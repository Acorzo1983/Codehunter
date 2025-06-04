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
	VERSION    = "2.5.7" // Incremented for new log format
	BUILD_DATE = "2024"
	AUTHOR     = "Albert.C"
	TWITTER    = "@yz9yt"
	GITHUB     = "https://github.com/Acorzo1983/Codehunter"
)

// ==============================================
// BANNER & BRANDING
// ==============================================
const BANNER = `
╔══════════════════════════════════════════════════════════════╗
║                     🏴‍☠️ CodeHunter v2.5.7                    ║
║              Ultra-Fast Bug Bounty Scanner                  ║
║                                                              ║
║    ┌─────────────────────────────────────────────────────┐   ║
║    │  🎯 Hunt APIs    🔐 Find Secrets   👑 Admin Panels │   ║
║    │  📜 JS Analysis  📁 File Discovery  🔗 Endpoints  │   ║
║    └─────────────────────────────────────────────────────┘   ║
║                                                              ║
║  🏴‍☠️ Perfect for: Kali Linux | Bug Bounty | Pentesting     ║
║  ⚡ Features: Multi-threaded | Pipe-friendly | Detailed Log  ║
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
	OutputFile       string
	Threads          int
	Verbose          bool
	ShowBanner       bool
	LogFile          string // For detailed, one-line-per-match-detail logs
	FoundUrlsLogFile string // For clean list of unique matched URLs
}

// ==============================================
// PATTERN INFO STRUCTURE
// ==============================================
type PatternInfo struct {
	RegexStr   string
	Compiled   *regexp.Regexp
	SourceFile string
}

// Structure to hold detailed match results for logging
type MatchDetail struct {
	Pattern     PatternInfo
	Occurrences []string
}

// ==============================================
// SCANNER STRUCTURE
// ==============================================
type Scanner struct {
	Config        Config
	Patterns      []PatternInfo
	Stats         ScanStats
	mu            sync.Mutex
	logFileMutex  sync.Mutex
	foundFile     *os.File
	logDetailFile *os.File // This will now be the structured, one-line-per-match log
}

type ScanStats struct {
	URLsProcessed int
	URLsMatched   int
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

	// Initial banner and startup messages to stdout
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
	defer scanner.CloseFiles()

	if err := scanner.setupOutputFiles(); err != nil {
		fmt.Printf("%s[ERROR]%s Setting up output files: %v\n", ColorRed, ColorReset, err) // To stdout
		os.Exit(1)
	}

	// Pattern loading messages will use logGeneralMessage (which might write to logDetailFile or stdout)
	if err := scanner.loadPatterns(); err != nil {
		scanner.logGeneralMessage(fmt.Sprintf("%s[ERROR]%s Failed to load patterns: %v\n", ColorRed, ColorReset, err), true)
		os.Exit(1)
	}

	if config.Verbose {
		scanner.logGeneralMessage(fmt.Sprintf("%s[INFO]%s Loaded %d patterns. Initial source: %s\n",
			ColorCyan, ColorReset, scanner.Stats.PatternsCount, config.PatternsFile), true)
	}

	var input io.Reader
	if config.UrlsFile != "" {
		file, err := os.Open(config.UrlsFile)
		if err != nil {
			scanner.logGeneralMessage(fmt.Sprintf("%s[ERROR]%s Cannot open URLs file: %v\n", ColorRed, ColorReset, err), true)
			os.Exit(1)
		}
		defer file.Close()
		input = file
		if config.Verbose {
			scanner.logGeneralMessage(fmt.Sprintf("%s[INFO]%s Reading URLs from: %s\n", ColorCyan, ColorReset, config.UrlsFile), true)
		}
	} else {
		input = os.Stdin
		if config.Verbose {
			scanner.logGeneralMessage(fmt.Sprintf("%s[INFO]%s Reading URLs from stdin (pipe mode)\n", ColorCyan, ColorReset), true)
		}
	}

	scanner.scan(input)
	// showFinalStats will now only print to stdout if banner/verbose, not to logDetailFile
	scanner.showFinalStats()

	// Final messages about where files were saved (to stdout)
	if scanner.Config.FoundUrlsLogFile != "" {
		fmt.Printf("%s[INFO]%s Matched URLs saved to: %s\n", ColorGreen, ColorReset, scanner.Config.FoundUrlsLogFile)
	}
	if scanner.Config.LogFile != "" {
		fmt.Printf("%s[INFO]%s Detailed match log saved to: %s\n", ColorGreen, ColorReset, scanner.Config.LogFile)
	}
}

// ==============================================
// HELPER FUNCTIONS
// ==============================================
func (s *Scanner) setupOutputFiles() error {
	if s.Config.FoundUrlsLogFile != "" {
		file, err := os.Create(s.Config.FoundUrlsLogFile)
		if err != nil {
			return fmt.Errorf("creating found URLs file '%s': %w", s.Config.FoundUrlsLogFile, err)
		}
		s.foundFile = file
	}

	if s.Config.LogFile != "" {
		file, err := os.Create(s.Config.LogFile)
		if err != nil {
			if s.foundFile != nil {
				s.foundFile.Close() // Clean up if partially successful
			}
			return fmt.Errorf("creating detailed match log file '%s': %w", s.Config.LogFile, err)
		}
		s.logDetailFile = file // This is for the structured one-line-per-match log
	}
	return nil
}

func (s *Scanner) CloseFiles() {
	if s.foundFile != nil {
		s.foundFile.Close()
	}
	if s.logDetailFile != nil {
		s.logDetailFile.Close()
	}
}

// logGeneralMessage is for startup, errors, verbose progress, pattern loading info.
// It writes to s.logDetailFile IF it's meant to be a general log (not the case anymore)
// OR to stdout under certain conditions.
// For v2.5.7, s.logDetailFile is a structured match log, so general messages primarily go to stdout.
func (s *Scanner) logGeneralMessage(message string, forceToStdout bool) {
	// In this version, s.logDetailFile is for specific match structures.
	// General messages (errors, verbose progress, pattern loading) should primarily go to stdout.
	// However, if a user *only* specifies --log-file and expects everything there, this might need adjustment.
	// For now, keeping it simple: these messages go to stdout if conditions are met.
	if forceToStdout || s.Config.Verbose || strings.Contains(message, "[ERROR]") || strings.Contains(message, "[WARN]") {
		fmt.Print(message)
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
		if config.ShowBanner {
			fmt.Println(BANNER)
		}
		fmt.Printf("\n%s🎯 CodeHunter v%s - Ultra-Fast Bug Bounty Scanner%s\n", ColorBold, VERSION, ColorReset)
		fmt.Printf("%sMade with ❤️ by %s (%s)%s\n\n", ColorPurple, AUTHOR, TWITTER, ColorReset)

		fmt.Printf("%s📋 Usage:%s\n", ColorYellow, ColorReset)
		fmt.Println("  codehunter -r patterns.txt -l urls.txt --found-urls found.txt --log-file detailed_matches.log")
		fmt.Println()

		fmt.Printf("%s💡 Test Examples (logs matched pattern & occurrences to --log-file):%s\n", ColorYellow, ColorReset)
		fmt.Println("  echo \"https://example.com/api/v1/users\" | codehunter -r api_endpoints.txt --found-urls hits.txt --log-file detailed_matches.log")
		fmt.Println("  echo \"http://test.com/api/key1?api_key=ABC&api_key=DEF\" | codehunter -r secrets.txt -v --log-file detailed_matches_verbose.log")
		fmt.Println()

		fmt.Printf("%s💡 Full Workflow Example with Katana:%s\n", ColorYellow, ColorReset)
		fmt.Println("  katana -u http://testhtml5.vulnweb.com/ | codehunter -r /usr/share/codehunter/patterns/secrets.txt,/usr/share/codehunter/patterns/api_endpoints.txt --found-urls k_found.txt --log-file k_matches.log")
		fmt.Println()

		fmt.Printf("%s🔧 Flags:%s\n", ColorYellow, ColorReset)
		flag.PrintDefaults()
		fmt.Println()

		fmt.Printf("%s📋 Available Patterns (default location: patterns/ or /usr/share/codehunter/patterns/):%s\n", ColorYellow, ColorReset)
		fmt.Println("  • secrets.txt      - API keys, tokens, credentials")
		fmt.Println("  • api_endpoints.txt - REST APIs, GraphQL, endpoints")
		// ... (other patterns)
		fmt.Println()
		fmt.Printf("%s🏴‍☠️ Happy Bug Hunting! 🏴‍☠️%s\n", ColorBold, ColorReset)
	}

	flag.StringVar(&config.PatternsFile, "r", "", "Patterns file(s), comma-separated (required)")
	flag.StringVar(&config.UrlsFile, "l", "", "URLs file (optional, uses stdin if not provided)")
	flag.StringVar(&config.OutputFile, "o", "", "Output file for matched URLs (legacy, use --found-urls for clarity)")
	flag.IntVar(&config.Threads, "t", 10, "Number of threads")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output (logs progress to stdout)")
	flag.BoolVar(&config.ShowBanner, "b", true, "Show banner (default: true, set to false with -b=false)")
	flag.StringVar(&config.LogFile, "log-file", "", "File to write detailed one-line-per-match log (URL, Pattern, Occurrences)")
	flag.StringVar(&config.FoundUrlsLogFile, "found-urls", "", "File to write clean list of unique matched URLs (optional)")

	flag.Parse()

	if config.PatternsFile == "" {
		fmt.Printf("%s[ERROR]%s Patterns file is required! Use -r <patterns_file>%s\n", ColorRed, ColorReset, ColorReset)
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
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
	var loadedPatterns []PatternInfo
	var filesSuccessfullyProcessed int
	var patternLoadingLog strings.Builder

	for _, sourceName := range patternFileSources {
		sourceName = strings.TrimSpace(sourceName)
		if sourceName == "" {
			continue
		}
		possiblePaths := []string{sourceName, filepath.Join("patterns", sourceName), filepath.Join("/usr/share/codehunter/patterns", sourceName)}
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
			msg := fmt.Sprintf("%s[WARN]%s Cannot open pattern file '%s'. Skipping.\n", ColorYellow, ColorReset, sourceName)
			patternLoadingLog.WriteString(msg) // Collect messages
			continue
		}
		baseSourceName := filepath.Base(usedPath)
		if s.Config.Verbose {
			msg := fmt.Sprintf("%s[INFO]%s Loading patterns from: %s\n", ColorCyan, ColorReset, usedPath)
			patternLoadingLog.WriteString(msg)
		}
		fileScanner := bufio.NewScanner(file)
		lineNum := 0
		patternsInThisFile := 0
		for fileScanner.Scan() {
			lineNum++
			line := strings.TrimSpace(fileScanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			compiledPattern, errRegex := regexp.Compile(line)
			if errRegex != nil {
				msg := fmt.Sprintf("%s[WARN]%s Invalid regex (line %d) in '%s': '%s' (%v). Skipping.\n", ColorYellow, ColorReset, lineNum, usedPath, line, errRegex)
				patternLoadingLog.WriteString(msg)
				continue
			}
			loadedPatterns = append(loadedPatterns, PatternInfo{RegexStr: line, Compiled: compiledPattern, SourceFile: baseSourceName})
			patternsInThisFile++
		}
		if errScan := fileScanner.Err(); errScan != nil {
			msg := fmt.Sprintf("%s[WARN]%s Error reading '%s': %v.\n", ColorYellow, ColorReset, usedPath, errScan)
			patternLoadingLog.WriteString(msg)
		}
		file.Close()
		if patternsInThisFile > 0 {
			filesSuccessfullyProcessed++
		}
	}

	s.logGeneralMessage(patternLoadingLog.String(), true) // Log collected messages

	s.Patterns = loadedPatterns
	s.Stats.PatternsCount = len(loadedPatterns)
	if len(s.Patterns) == 0 {
		return fmt.Errorf("no valid patterns loaded from any specified sources ('%s')", s.Config.PatternsFile)
	}
	if filesSuccessfullyProcessed < len(patternFileSources) && len(patternFileSources) > 1 {
		s.logGeneralMessage(fmt.Sprintf("%s[INFO]%s Processed patterns from %d of %d sources.\n", ColorCyan, ColorReset, filesSuccessfullyProcessed, len(patternFileSources)), true)
	}
	return nil
}


// ==============================================
// MAIN SCANNING LOGIC
// ==============================================
func (s *Scanner) scan(input io.Reader) {
	urlChan := make(chan string, s.Config.Threads*2)
	uniqueMatchedURLChan := make(chan string, s.Config.Threads*2) // For --found-urls output

	var readerWg, writerWg, workerWg sync.WaitGroup

	if s.foundFile != nil {
		writerWg.Add(1)
		go func() {
			defer writerWg.Done()
			// Use a map to write unique URLs to the foundFile
			writtenURLs := make(map[string]bool)
			for resultURL := range uniqueMatchedURLChan {
				if !writtenURLs[resultURL] {
					fmt.Fprintln(s.foundFile, resultURL)
					writtenURLs[resultURL] = true
				}
			}
		}()
	}

	for i := 0; i < s.Config.Threads; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for url := range urlChan {
				s.processURL(url, uniqueMatchedURLChan, workerID)
			}
		}(i)
	}

	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		defer close(urlChan)
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" && !strings.HasPrefix(url, "#") {
				urlChan <- url
			}
		}
		if err := scanner.Err(); err != nil {
			s.logGeneralMessage(fmt.Sprintf("%s[ERROR]%s Error reading URLs: %v\n", ColorRed, ColorReset, err), true)
		}
	}()

	readerWg.Wait()
	workerWg.Wait()
	close(uniqueMatchedURLChan)

	if s.foundFile != nil {
		writerWg.Wait()
	}
	s.Stats.EndTime = time.Now()
}

// ==============================================
// URL PROCESSING (Writes one line per pattern match detail to logDetailFile)
// ==============================================
func (s *Scanner) processURL(url string, uniqueMatchedURLChan chan<- string, workerID int) {
	s.mu.Lock()
	s.Stats.URLsProcessed++
	currentProcessed := s.Stats.URLsProcessed
	s.mu.Unlock()

	if s.Config.Verbose && currentProcessed%100 == 0 {
		s.logGeneralMessage(fmt.Sprintf("%s[PROGRESS]%s Processed %d URLs (Worker %d)\n",
			ColorBlue, ColorReset, currentProcessed, workerID), false)
	}

	var overallMatchForURL bool = false

	for _, pInfo := range s.Patterns { // Test ALL patterns against the URL
		foundOccurrences := pInfo.Compiled.FindAllString(url, -1)

		if len(foundOccurrences) > 0 {
			if !overallMatchForURL { // If this is the first pattern to match this URL
				overallMatchForURL = true
				s.mu.Lock()
				s.Stats.URLsMatched++ // Increment unique matched URL count
				s.mu.Unlock()

				// Send to --found-urls file (only the URL, once per URL)
				if s.foundFile != nil {
					uniqueMatchedURLChan <- url
				} else if !s.Config.Verbose { // If no --found-urls AND not verbose, print unique matched URL to stdout
					fmt.Println(url)
				}
			}

			// Log this specific pattern match detail to the --log-file
			if s.logDetailFile != nil {
				occurrencesString := strings.Join(foundOccurrences, " - ")
				logLine := fmt.Sprintf("%s MATCHED_PATTERN: %s (From: %s) FOUND [%d time(s)]:- %s\n",
					url, pInfo.RegexStr, pInfo.SourceFile, len(foundOccurrences), occurrencesString)

				s.logFileMutex.Lock()
				fmt.Fprint(s.logDetailFile, logLine)
				s.logFileMutex.Unlock()
			}

			if s.Config.Verbose { // Verbose output for each pattern hit
				s.logGeneralMessage(fmt.Sprintf("%s[MATCH_DETAIL]%s %s (Pattern: %s, Occurrences: %d)\n",
					ColorGreen, ColorReset, url, pInfo.RegexStr, len(foundOccurrences)), false)
			}
		}
	}
}

// ==============================================
// STATISTICS DISPLAY (Now only prints to STDOUT if banner/verbose)
// ==============================================
func (s *Scanner) showFinalStats() {
	// Only proceed to build and print stats if banner or verbose mode is on
	if !s.Config.ShowBanner && !s.Config.Verbose {
		return
	}

	var statsBuilder strings.Builder
	duration := s.Stats.EndTime.Sub(s.Stats.StartTime)
	speed := 0.0
	if duration.Seconds() > 0 {
		speed = float64(s.Stats.URLsProcessed) / duration.Seconds()
	}

	// ... (rest of the statsBuilder formatting as in v2.5.5) ...
	statsBuilder.WriteString(fmt.Sprintf("\n%s╔══════════════════════════════════════════════════════════╗%s\n", ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║                    🏴‍☠️ HUNT COMPLETE 🏴‍☠️                  ║%s\n", ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s╠══════════════════════════════════════════════════════════╣%s\n", ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  📊 URLs Processed: %s%-10d%s                     %s║%s\n",
		ColorPurple, ColorReset, ColorCyan, s.Stats.URLsProcessed, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  🎯 URLs Matched:   %s%-10d%s                     %s║%s\n", // Unique URLs matched
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
	} else if s.Stats.URLsMatched < 10 && s.Stats.URLsMatched > 0 {
		matchStatus = "Low match rate"
		matchColor = ColorYellow
	} else {
		matchStatus = "Good match rate"
		matchColor = ColorGreen
	}
	statsBuilder.WriteString(fmt.Sprintf("%s║%s  🎯 Status:         %s%-25s%s %s║%s\n",
		ColorPurple, ColorReset, matchColor, matchStatus, ColorReset, ColorPurple, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s╚══════════════════════════════════════════════════════════╝%s\n", ColorPurple, ColorReset))

	if s.Stats.URLsMatched == 0 {
		statsBuilder.WriteString(fmt.Sprintf("\n%s💡 Tips for better results:%s\n", ColorYellow, ColorReset))
		statsBuilder.WriteString("  • Try different pattern files or combine them (e.g., -r secrets.txt,api_endpoints.txt)\n")
		statsBuilder.WriteString("  • Ensure your input URLs actually contain data that patterns might match.\n")
		statsBuilder.WriteString("  • Use -v for verbose progress. Check --log-file for detailed match occurrences.\n")
		statsBuilder.WriteString("  • Verify pattern file syntax and regex validity.\n")
	} else {
		statsBuilder.WriteString(fmt.Sprintf("\n%s🎉 Great! Found potential targets.%s\n", ColorGreen, ColorReset))
		if s.Config.FoundUrlsLogFile == "" && !s.Config.Verbose {
			statsBuilder.WriteString("  🔍 Review matched URLs printed above or use --found-urls <file> to save them.\n")
		}
		statsBuilder.WriteString("  🛡️ Follow responsible disclosure practices.\n")
	}

	statsBuilder.WriteString(fmt.Sprintf("\n%s🏴‍☠️ Made with ❤️ by %s (%s) | %s 🏴‍☠️%s\n",
		ColorBold, AUTHOR, TWITTER, GITHUB, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s🎯 Happy Bug Hunting! 🎯%s\n\n", ColorBold, ColorReset))

	// Print final stats summary to STDOUT only if banner or verbose is enabled.
	// The detailed match log (--log-file) will NOT contain this summary.
	fmt.Print(statsBuilder.String())
}
