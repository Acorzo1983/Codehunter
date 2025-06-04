package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ==============================================
// VERSION & BUILD INFO
// ==============================================
var (
	VERSION    = "2.5.5" // Incremented for exhaustive pattern matching log
	BUILD_DATE = "2024"  // This could be dynamically updated in a real build
	AUTHOR     = "Albert.C"
	TWITTER    = "@yz9yt"
	GITHUB     = "https://github.com/Acorzo1983/Codehunter"
)

// ==============================================
// BANNER & BRANDING
// ==============================================
const BANNER = `
╔══════════════════════════════════════════════════════════════╗
║                     🏴‍☠️ CodeHunter v2.5.5                    ║
║              Ultra-Fast Bug Bounty Scanner                  ║
║                                                              ║
║    ┌─────────────────────────────────────────────────────┐   ║
║    │  🎯 Hunt APIs    🔐 Find Secrets   👑 Admin Panels │   ║
║    │  📜 JS Analysis  📁 File Discovery  🔗 Endpoints  │   ║
║    └─────────────────────────────────────────────────────┘   ║
║                                                              ║
║  🏴‍☠️ Perfect for: Kali Linux | Bug Bounty | Pentesting     ║
║  ⚡ Features: Multi-threaded | Pipe-friendly | Exhaustive Log║
║                                                              ║
║             Made with ❤️  by Albert.C (@yz9yt)              ║
║          github.com/Acorzo1983/Codehunter                   ║
╚══════════════════════════════════════════════════════════════╝`

// ==============================================
// CONFIGURATION STRUCTURE
// ==============================================
type Config struct {
	PatternsFile     string
	UrlsFile         string
	OutputFile       string // Legacy -o, for backward compatibility or preference for found URLs
	Threads          int
	Verbose          bool
	ShowBanner       bool
	LogFile          string // New: for the detailed log file
	FoundUrlsLogFile string // New: for clean list of found URLs
}

// ==============================================
// PATTERN INFO STRUCTURE
// ==============================================
type PatternInfo struct {
	RegexStr   string         // The original regex string
	Compiled   *regexp.Regexp // The compiled regex
	SourceFile string         // The base name of the pattern file it came from
}

// Structure to hold detailed match results for logging
type MatchDetail struct {
	Pattern     PatternInfo // Information about the pattern that matched
	Occurrences []string    // All occurrences of this pattern in the URL
}

// ==============================================
// SCANNER STRUCTURE
// ==============================================
type Scanner struct {
	Config        Config
	Patterns      []PatternInfo // Changed to use PatternInfo
	Stats         ScanStats
	mu            sync.Mutex   // General mutex for Stats
	logFileMutex  sync.Mutex   // NEW: Mutex for concurrent writes to logDetailFile
	foundFile     *os.File     // File handle for --found-urls
	logDetailFile *os.File     // File handle for --log-file
}

type ScanStats struct {
	URLsProcessed int
	URLsMatched   int // Counts unique URLs that had at least one pattern match
	PatternsCount int
	StartTime     time.Time
	EndTime       time.Time
}

// ==============================================
// COLOR CONSTANTS
// ==============================================
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

// ==============================================
// MAIN FUNCTION
// ==============================================
func main() {
	config := parseFlags()

	if config.ShowBanner {
		fmt.Println(BANNER)
		fmt.Printf("\n%s🚀 Starting CodeHunter v%s - Made with ❤️ by %s%s\n",
			ColorGreen, VERSION, AUTHOR, ColorReset)
		fmt.Printf("%s📅 Build: %s | 🐹 Go: %s | 💻 OS: %s%s\n\n",
			ColorBlue, BUILD_DATE, runtime.Version(), runtime.GOOS, ColorReset)
	}

	scanner := &Scanner{
		Config: config,
		Stats: ScanStats{
			StartTime: time.Now(),
		},
	}
	defer scanner.CloseFiles() // Ensure files are closed at the end

	// Setup output files first
	if err := scanner.setupOutputFiles(); err != nil {
		// This initial error goes to stdout as log files might not be set up yet
		fmt.Printf("%s[ERROR]%s Setting up output files: %v\n", ColorRed, ColorReset, err)
		os.Exit(1)
	}

	// Load patterns
	if err := scanner.loadPatterns(); err != nil {
		// Log to detail log/stdout and exit
		scanner.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Failed to load patterns: %v\n", ColorRed, ColorReset, err), true)
		os.Exit(1)
	}

	if config.Verbose {
		scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Loaded %d patterns. Initial source: %s\n",
			ColorCyan, ColorReset, scanner.Stats.PatternsCount, config.PatternsFile), true)
	}

	// Setup input (URLs source)
	var input io.Reader
	if config.UrlsFile != "" {
		file, err := os.Open(config.UrlsFile)
		if err != nil {
			scanner.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Cannot open URLs file: %v\n", ColorRed, ColorReset, err), true)
			os.Exit(1)
		}
		defer file.Close()
		input = file
		if config.Verbose {
			scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Reading URLs from: %s\n", ColorCyan, ColorReset, config.UrlsFile), true)
		}
	} else {
		input = os.Stdin
		if config.Verbose {
			scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Reading URLs from stdin (pipe mode)\n", ColorCyan, ColorReset), true)
		}
	}

	// Start scanning
	scanner.scan(input)
	// Show final statistics (will also write to logDetailFile if configured)
	scanner.showFinalStats()

	// Inform user about created files after stats are shown/logged
	if scanner.Config.FoundUrlsLogFile != "" {
		finalMsg := fmt.Sprintf("%s[INFO]%s Matched URLs saved to: %s\n", ColorGreen, ColorReset, scanner.Config.FoundUrlsLogFile)
		fmt.Print(finalMsg) // Always print this to stdout for user visibility
		if scanner.logDetailFile != nil && scanner.logDetailFile != os.Stdout { // Avoid double printing if log is stdout
			fmt.Fprint(scanner.logDetailFile, finalMsg)
		}
	}
	if scanner.Config.LogFile != "" {
		finalMsg := fmt.Sprintf("%s[INFO]%s Detailed scan log (with all matched patterns and occurrences) saved to: %s\n", ColorGreen, ColorReset, scanner.Config.LogFile)
		fmt.Print(finalMsg) // Always print this to stdout
		// No need to write to logDetailFile itself again, as it's about the file itself.
	}
}

// ==============================================
// HELPER FUNCTIONS for file handling and logging
// ==============================================
func (s *Scanner) setupOutputFiles() error {
	if s.Config.FoundUrlsLogFile != "" {
		file, err := os.Create(s.Config.FoundUrlsLogFile)
		if err != nil {
			return fmt.Errorf("creating found URLs file '%s': %w", s.Config.FoundUrlsLogFile, err)
		}
		s.foundFile = file // Store the file handle
	}

	if s.Config.LogFile != "" {
		file, err := os.Create(s.Config.LogFile)
		if err != nil {
			// If foundFile was created, close it before erroring
			if s.foundFile != nil {
				s.foundFile.Close()
			}
			return fmt.Errorf("creating detailed log file '%s': %w", s.Config.LogFile, err)
		}
		s.logDetailFile = file // Store the file handle
	}
	return nil
}

func (s *Scanner) CloseFiles() {
	if s.foundFile != nil {
		s.foundFile.Close()
	}
	if s.logDetailFile != nil {
		// Avoid closing os.Stdout if it was assigned
		if f, ok := s.logDetailFile.(*os.File); ok && f != os.Stdout {
			s.logDetailFile.Close()
		}
	}
}

// logDetailMessage logs messages.
// If s.logDetailFile is set, it logs there.
// If forceToStdoutIfNotLogging is true OR specific conditions (verbose, error, warn) are met,
// AND s.logDetailFile is nil (meaning no dedicated log file), it logs to stdout.
func (s *Scanner) logDetailMessage(message string, forceToStdoutIfNotLogging bool) {
	if s.logDetailFile != nil {
		// Write to the designated log file.
		// Concurrent writes from processURL are handled by logFileMutex there.
		// Other calls (loadPatterns, main) are serial.
		fmt.Fprint(s.logDetailFile, message)
	}

	// If no dedicated log file, decide whether to print to stdout.
	if s.logDetailFile == nil {
		if forceToStdoutIfNotLogging || s.Config.Verbose || strings.Contains(message, "[ERROR]") || strings.Contains(message, "[WARN]") {
			fmt.Print(message)
		}
	}
}

// ==============================================
// COMMAND LINE FLAG PARSING
// ==============================================
func parseFlags() Config {
	config := Config{
		Threads:    10,
		Verbose:    false,
		ShowBanner: true,
	}

	flag.Usage = func() {
		if config.ShowBanner { // Show banner if not suppressed
			fmt.Println(BANNER)
		}
		fmt.Printf("\n%s🎯 CodeHunter v%s - Ultra-Fast Bug Bounty Scanner%s\n", ColorBold, VERSION, ColorReset)
		fmt.Printf("%sMade with ❤️ by %s (%s)%s\n\n", ColorPurple, AUTHOR, TWITTER, ColorReset)

		fmt.Printf("%s📋 Usage:%s\n", ColorYellow, ColorReset)
		fmt.Println("  codehunter -r patterns.txt -l urls.txt --found-urls found.txt --log-file scan.log")
		fmt.Println()

		fmt.Printf("%s💡 Test Examples (logs matched pattern & occurrences):%s\n", ColorYellow, ColorReset)
		fmt.Println("  echo \"https://example.com/api/v1/users\" | codehunter -r api_endpoints.txt --found-urls hits.txt --log-file detailed.log")
		fmt.Println("  echo \"http://test.com/config.js?token=abcdef1234567890&token=anotherToken\" | codehunter -r js_secrets.txt -v --log-file detailed_verbose.log")
		fmt.Println()

		fmt.Printf("%s💡 Full Workflow Example with Katana:%s\n", ColorYellow, ColorReset)
		fmt.Println("  katana -u http://testhtml5.vulnweb.com/ | codehunter -r /usr/share/codehunter/patterns/secrets.txt,/usr/share/codehunter/patterns/api_endpoints.txt --found-urls k_found.txt --log-file k_log.txt")
		fmt.Println()

		fmt.Printf("%s🔧 Flags:%s\n", ColorYellow, ColorReset)
		flag.PrintDefaults()
		fmt.Println()

		fmt.Printf("%s📋 Available Patterns (default location: patterns/ or /usr/share/codehunter/patterns/):%s\n", ColorYellow, ColorReset)
		fmt.Println("  • secrets.txt      - API keys, tokens, credentials")
		fmt.Println("  • api_endpoints.txt - REST APIs, GraphQL, endpoints")
		fmt.Println("  • admin_panels.txt  - Admin areas, CMS panels")
		fmt.Println("  • js_secrets.txt   - JavaScript secrets, configs")
		fmt.Println("  • files.txt        - Sensitive files, backups")
		fmt.Println("  • custom.txt       - Your custom patterns")
		fmt.Println()

		fmt.Printf("%s🏴‍☠️ Happy Bug Hunting! 🏴‍☠️%s\n", ColorBold, ColorReset)
	}

	flag.StringVar(&config.PatternsFile, "r", "", "Patterns file(s), comma-separated (required)")
	flag.StringVar(&config.UrlsFile, "l", "", "URLs file (optional, uses stdin if not provided)")
	flag.StringVar(&config.OutputFile, "o", "", "Output file for matched URLs (legacy, use --found-urls for clarity)")
	flag.IntVar(&config.Threads, "t", 10, "Number of threads")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output (logs progress to stdout/--log-file)")
	flag.BoolVar(&config.ShowBanner, "b", true, "Show banner (default: true, set to false with -b=false)")
	flag.StringVar(&config.LogFile, "log-file", "", "File to write detailed scan log (includes all matched patterns and occurrences)")
	flag.StringVar(&config.FoundUrlsLogFile, "found-urls", "", "File to write clean list of matched URLs (optional)")

	flag.Parse()

	// Validate required flags
	if config.PatternsFile == "" {
		// This error message will go to stdout because log files are not set up yet.
		fmt.Printf("%s[ERROR]%s Patterns file is required! Use -r <patterns_file>%s\n", ColorRed, ColorReset, ColorReset)
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
	// If -o is used and --found-urls is not, make --found-urls same as -o for backward compatibility
	if config.OutputFile != "" && config.FoundUrlsLogFile == "" {
		config.FoundUrlsLogFile = config.OutputFile
	}

	return config
}

// ==============================================
// PATTERN LOADING
// ==============================================
func (s *Scanner) loadPatterns() error {
	patternFileSources := strings.Split(s.Config.PatternsFile, ",")
	var loadedPatterns []PatternInfo // Use PatternInfo
	var filesSuccessfullyProcessed int
	var patternLoadingLog strings.Builder // Accumulate log messages for patterns

	for _, sourceName := range patternFileSources {
		sourceName = strings.TrimSpace(sourceName)
		if sourceName == "" {
			continue
		}

		// Try multiple locations for pattern files
		possiblePaths := []string{
			sourceName,                                    // Direct path
			filepath.Join("patterns", sourceName),         // Local patterns directory
			filepath.Join("/usr/share/codehunter/patterns", sourceName), // System patterns
		}

		var file *os.File
		var err error
		var usedPath string
		opened := false

		for _, path := range possiblePaths {
			file, err = os.Open(path)
			if err == nil {
				usedPath = path
				opened = true
				break
			}
		}

		if !opened {
			msg := fmt.Sprintf("%s[WARN]%s Cannot open pattern file '%s' in any location. Skipping.\n", ColorYellow, ColorReset, sourceName)
			patternLoadingLog.WriteString(msg)
			continue
		}

		baseSourceName := filepath.Base(usedPath) // For logging which file the pattern came from

		if s.Config.Verbose { // Log this specific info only if verbose
			msg := fmt.Sprintf("%s[INFO]%s Loading patterns from: %s\n", ColorCyan, ColorReset, usedPath)
			patternLoadingLog.WriteString(msg)
		}

		fileScanner := bufio.NewScanner(file)
		lineNum := 0
		patternsInThisFile := 0

		for fileScanner.Scan() {
			lineNum++
			line := strings.TrimSpace(fileScanner.Text()) // This is RegexStr

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Compile regex pattern
			compiledPattern, errRegex := regexp.Compile(line)
			if errRegex != nil {
				// Log warning about invalid regex
				msg := fmt.Sprintf("%s[WARN]%s Invalid regex at line %d in %s ('%s'): %s. Skipping pattern.\n",
					ColorYellow, ColorReset, lineNum, usedPath, line, errRegex)
				patternLoadingLog.WriteString(msg)
				continue // Skip this invalid pattern
			}
			loadedPatterns = append(loadedPatterns, PatternInfo{
				RegexStr:   line, // Store the original string
				Compiled:   compiledPattern,
				SourceFile: baseSourceName,
			})
			patternsInThisFile++
		}

		// Check for errors during file scan (e.g., permission issues mid-file)
		if errScan := fileScanner.Err(); errScan != nil {
			msg := fmt.Sprintf("%s[WARN]%s Error reading file %s: %v. Patterns loaded from this file might be incomplete.\n",
				ColorYellow, ColorReset, usedPath, errScan)
			patternLoadingLog.WriteString(msg)
		}
		file.Close() // Close file after processing it

		if patternsInThisFile > 0 {
			filesSuccessfullyProcessed++
		}
	}

	// Log all accumulated pattern loading messages.
	// The 'true' forces write to stdout if no log file AND message is error/warn OR verbose is on.
	s.logDetailMessage(patternLoadingLog.String(), true)


	s.Patterns = loadedPatterns
	s.Stats.PatternsCount = len(loadedPatterns)

	if len(s.Patterns) == 0 {
		// This is a fatal error, so it should go to stderr/stdout if log isn't set up.
		return fmt.Errorf("no valid patterns loaded from any specified sources ('%s'). Please check pattern file paths and content", s.Config.PatternsFile)
	}

	// Log if some files were processed but not all (if multiple were specified)
	if filesSuccessfullyProcessed < len(patternFileSources) && len(patternFileSources) > 1 {
		s.logDetailMessage(fmt.Sprintf("%s[INFO]%s Successfully processed patterns from %d out of %d specified source(s).\n",
			ColorCyan, ColorReset, filesSuccessfullyProcessed, len(patternFileSources)), true)
	}
	return nil
}

// ==============================================
// MAIN SCANNING LOGIC
// ==============================================
func (s *Scanner) scan(input io.Reader) {
	urlChan := make(chan string, s.Config.Threads*2)
	resultChan := make(chan string, s.Config.Threads*2) // For --found-urls output

	var readerWg, writerWg, workerWg sync.WaitGroup

	// Start result writer goroutine (for --found-urls file)
	if s.foundFile != nil { // Only start writer if foundFile is configured
		writerWg.Add(1)
		go func() {
			defer writerWg.Done()
			for resultURL := range resultChan {
				fmt.Fprintln(s.foundFile, resultURL)
			}
		}()
	}

	// Start worker goroutines
	for i := 0; i < s.Config.Threads; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for url := range urlChan {
				s.processURL(url, resultChan, workerID) // Pass resultChan
			}
		}(i)
	}

	// Start URL reader goroutine
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		defer close(urlChan) // Close urlChan when reader is done
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" && !strings.HasPrefix(url, "#") {
				urlChan <- url
			}
		}
		if err := scanner.Err(); err != nil {
			// Force to stdout if no log file, as this is an error
			s.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Error reading URLs: %v\n", ColorRed, ColorReset, err), true)
		}
	}()

	readerWg.Wait()   // Wait for reader to finish sending all URLs
	workerWg.Wait()   // Wait for all workers to finish processing
	close(resultChan) // Now close resultChan (used for --found-urls)

	if s.foundFile != nil {
		writerWg.Wait() // Wait for found-urls writer to finish
	}

	s.Stats.EndTime = time.Now()
}

// ==============================================
// URL PROCESSING (Logs all occurrences of all matching patterns)
// ==============================================
func (s *Scanner) processURL(url string, resultChan chan<- string, workerID int) {
	s.mu.Lock() // Protect s.Stats
	s.Stats.URLsProcessed++
	currentProcessed := s.Stats.URLsProcessed
	s.mu.Unlock()

	if s.Config.Verbose && currentProcessed%100 == 0 {
		// Not forced to stdout if no log, as it's just verbose progress
		s.logDetailMessage(fmt.Sprintf("%s[PROGRESS]%s Processed %d URLs (Worker %d)\n",
			ColorBlue, ColorReset, currentProcessed, workerID), false)
	}

	var overallMatchForURL bool = false       // Flag to track if this URL matched ANY pattern
	var allMatchesForThisURL []MatchDetail // Store all pattern matches for this URL

	for _, pInfo := range s.Patterns { // NO BREAK HERE - Test ALL patterns
		foundOccurrences := pInfo.Compiled.FindAllString(url, -1) // Find ALL occurrences for THIS pattern

		if len(foundOccurrences) > 0 {
			overallMatchForURL = true // Mark that this URL had at least one match
			allMatchesForThisURL = append(allMatchesForThisURL, MatchDetail{
				Pattern:     pInfo,
				Occurrences: foundOccurrences,
			})
		}
	}

	if overallMatchForURL {
		s.mu.Lock() // Protect s.Stats
		s.Stats.URLsMatched++ // Increment for the URL if it matched any pattern
		s.mu.Unlock()

		// Log all detailed matches for this URL to the log file
		if s.logDetailFile != nil && len(allMatchesForThisURL) > 0 {
			var logEntry strings.Builder
			logEntry.WriteString(fmt.Sprintf("URL: %s\n", url))
			for _, match := range allMatchesForThisURL {
				logEntry.WriteString(fmt.Sprintf("  MATCHED_PATTERN: %s (From: %s)\n", match.Pattern.RegexStr, match.Pattern.SourceFile))
				logEntry.WriteString(fmt.Sprintf("  FOUND [%d time(s)]:\n", len(match.Occurrences)))
				for _, occurrence := range match.Occurrences {
					logEntry.WriteString(fmt.Sprintf("    - %s\n", occurrence))
				}
			}
			logEntry.WriteString("---\n")

			s.logFileMutex.Lock() // Protect concurrent writes to the log file
			fmt.Fprint(s.logDetailFile, logEntry.String())
			s.logFileMutex.Unlock()
		}

		if s.Config.Verbose { // Verbose output to stdout/log for a quick indication
			// For verbose, just indicate a match and maybe the number of patterns that hit
			// Not forced to stdout if no log, as it's verbose progress
			s.logDetailMessage(fmt.Sprintf("%s[MATCH]%s %s (Matched by %d patterns)\n",
				ColorGreen, ColorReset, url, len(allMatchesForThisURL)), false)
		}

		// Send to --found-urls file (only the URL, once)
		if s.foundFile != nil {
			resultChan <- url
		} else if !s.Config.Verbose { // If no --found-urls AND not verbose, print matched URL to stdout
			fmt.Println(url)
		}
	}
}

// ==============================================
// STATISTICS DISPLAY
// ==============================================
func (s *Scanner) showFinalStats() {
	// Prepare the statistics string using a builder
	var statsBuilder strings.Builder

	duration := s.Stats.EndTime.Sub(s.Stats.StartTime)
	speed := 0.0
	if duration.Seconds() > 0 {
		speed = float64(s.Stats.URLsProcessed) / duration.Seconds()
	}

	statsBuilder.WriteString(fmt.Sprintf("\n%s╔══════════════════════════════════════════════════════════╗%s\n", ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║                    🏴‍☠️ HUNT COMPLETE 🏴‍☠️                  ║%s\n", ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s╠══════════════════════════════════════════════════════════╣%s\n", ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  📊 URLs Processed: %s%-10d%s                     %s║%s\n",
		ColorPurple, ColorReset, ColorCyan, s.Stats.URLsProcessed, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  🎯 URLs Matched:   %s%-10d%s                     %s║%s\n",
		ColorPurple, ColorReset, ColorGreen, s.Stats.URLsMatched, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  🔍 Patterns Used:  %s%-10d%s                     %s║%s\n",
		ColorPurple, ColorReset, ColorYellow, s.Stats.PatternsCount, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  ⏱️ Duration:       %s%-10s%s                     %s║%s\n",
		ColorPurple, ColorReset, ColorBlue, duration.Truncate(time.Millisecond).String(), ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  🚀 Speed:          %s%-10.1f URLs/sec%s           %s║%s\n",
		ColorPurple, ColorReset, ColorCyan, speed, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s╠══════════════════════════════════════════════════════════╣%s\n", ColorPurple, ColorReset))

	var matchStatus, matchColor string
	if s.Stats.URLsMatched == 0 {
		matchStatus = "No matches found"
		matchColor = ColorYellow
	} else if s.Stats.URLsMatched < 10 && s.Stats.URLsMatched > 0 { // Positive but low
		matchStatus = "Low match rate"
		matchColor = ColorYellow
	} else { // s.Stats.URLsMatched >= 10
		matchStatus = "Good match rate"
		matchColor = ColorGreen
	}
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  🎯 Status:         %s%-25s%s %s║%s\n",
		ColorPurple, ColorReset, matchColor, matchStatus, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s╚══════════════════════════════════════════════════════════╝%s\n", ColorPurple, ColorReset))

	// Tips
	if s.Stats.URLsMatched == 0 {
		statsBuilder.WriteString(fmt.Sprintf("\n%s💡 Tips for better results:%s\n", ColorYellow, ColorReset))
		statsBuilder.WriteString("  • Try different pattern files or combine them (e.g., -r secrets.txt,api_endpoints.txt)\n")
		statsBuilder.WriteString("  • Ensure your input URLs actually contain data that patterns might match.\n")
		statsBuilder.WriteString("  • Use -v for verbose progress, or --log-file for a persistent log with matched patterns and occurrences.\n")
		statsBuilder.WriteString("  • Verify pattern file syntax and regex validity.\n")
	} else { // Matches were found
		statsBuilder.WriteString(fmt.Sprintf("\n%s🎉 Great! Found potential targets.%s\n", ColorGreen, ColorReset))
		if s.Config.FoundUrlsLogFile == "" && !s.Config.Verbose { // If matches were printed to stdout because no found-file and not verbose
			statsBuilder.WriteString("  🔍 Review matched URLs printed above or use --found-urls <file> to save them.\n")
		}
		statsBuilder.WriteString("  🛡️ Follow responsible disclosure practices.\n")
	}

	statsBuilder.WriteString(fmt.Sprintf("\n%s🏴‍☠️ Made with ❤️ by %s (%s) | %s 🏴‍☠️%s\n",
		ColorBold, AUTHOR, TWITTER, GITHUB, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s🎯 Happy Bug Hunting! 🎯%s\n\n", ColorBold, ColorReset))

	// Write the stats to the detailed log file or stdout
	finalStatsMessage := statsBuilder.String()
	if s.logDetailFile != nil {
		fmt.Fprint(s.logDetailFile, finalStatsMessage) // Always write to the log file if it exists
		// Additionally, if banner or verbose is on, also print stats to stdout for immediate visibility
		if s.Config.ShowBanner || s.Config.Verbose {
             // Add an extra newline if not already well-spaced by verbose output
			if !strings.HasSuffix(finalStatsMessage, "\n\n") && !strings.HasSuffix(finalStatsMessage, "\n") {
                 fmt.Println()
            }
			fmt.Print(finalStatsMessage)
		}
	} else { // No dedicated log file, so print to stdout if banner/verbose
		if s.Config.ShowBanner || s.Config.Verbose {
			fmt.Print(finalStatsMessage)
		}
	}
}
