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
	VERSION    = "2.5.2" // Incremented version for new features
	BUILD_DATE = "2024"  // Esto podría actualizarse dinámicamente en un build real
	AUTHOR     = "Albert.C"
	TWITTER    = "@yz9yt"
	GITHUB     = "https://github.com/Acorzo1983/Codehunter"
)

// ==============================================
// BANNER & BRANDING
// ==============================================
const BANNER = `
╔══════════════════════════════════════════════════════════════╗
║                     🏴‍☠️ CodeHunter v2.5.2                    ║
║              Ultra-Fast Bug Bounty Scanner                  ║
║                                                              ║
║    ┌─────────────────────────────────────────────────────┐   ║
║    │  🎯 Hunt APIs    🔐 Find Secrets   👑 Admin Panels │   ║
║    │  📜 JS Analysis  📁 File Discovery  🔗 Endpoints  │   ║
║    └─────────────────────────────────────────────────────┘   ║
║                                                              ║
║  🏴‍☠️ Perfect for: Kali Linux | Bug Bounty | Pentesting     ║
║  ⚡ Features: Multi-threaded | Pipe-friendly | 325+ Patterns║
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
	OutputFile       string // Legacy -o, could be for found URLs if new flags aren't used
	Threads          int
	Verbose          bool
	ShowBanner       bool
	LogFile          string // New: for the detailed log file
	FoundUrlsLogFile string // New: for clean list of found URLs
}

// ==============================================
// SCANNER STRUCTURE
// ==============================================
type Scanner struct {
	Config        Config
	Patterns      []*regexp.Regexp
	Stats         ScanStats
	mu            sync.Mutex
	foundFile     *os.File // File handle for found URLs
	logDetailFile *os.File // File handle for detailed log
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
		fmt.Printf("%s[ERROR]%s Setting up output files: %v\n", ColorRed, ColorReset, err)
		os.Exit(1)
	}

	if err := scanner.loadPatterns(); err != nil {
		scanner.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Failed to load patterns: %v\n", ColorRed, ColorReset, err))
		os.Exit(1)
	}

	if config.Verbose {
		scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Loaded %d patterns. Initial source: %s\n",
			ColorCyan, ColorReset, scanner.Stats.PatternsCount, config.PatternsFile))
	}

	var input io.Reader
	if config.UrlsFile != "" {
		file, err := os.Open(config.UrlsFile)
		if err != nil {
			scanner.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Cannot open URLs file: %v\n", ColorRed, ColorReset, err))
			os.Exit(1)
		}
		defer file.Close()
		input = file
		if config.Verbose {
			scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Reading URLs from: %s\n", ColorCyan, ColorReset, config.UrlsFile))
		}
	} else {
		input = os.Stdin
		if config.Verbose {
			scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Reading URLs from stdin (pipe mode)\n", ColorCyan, ColorReset))
		}
	}

	scanner.scan(input)
	scanner.showFinalStats() // This will now also write to logDetailFile if configured

	// Inform user about created files
	if scanner.Config.FoundUrlsLogFile != "" {
		fmt.Printf("%s[INFO]%s Matched URLs saved to: %s\n", ColorGreen, ColorReset, scanner.Config.FoundUrlsLogFile)
	}
	if scanner.Config.LogFile != "" {
		fmt.Printf("%s[INFO]%s Detailed scan log saved to: %s\n", ColorGreen, ColorReset, scanner.Config.LogFile)
	}
}

// Helper to setup output files for the Scanner
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

// Helper to close files, to be deferred
func (s *Scanner) CloseFiles() {
	if s.foundFile != nil {
		s.foundFile.Close()
	}
	if s.logDetailFile != nil {
		s.logDetailFile.Close()
	}
}

// Helper to log messages to detailed log file or stdout
func (s *Scanner) logDetailMessage(message string) {
	if s.logDetailFile != nil {
		fmt.Fprint(s.logDetailFile, message)
	} else {
		// If no detail log file, print to stdout only if verbose or it's an error
		if s.Config.Verbose || strings.Contains(message, "[ERROR]") || strings.Contains(message, "[WARN]") {
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
		fmt.Println("  echo \"http://httpbin.org/get?api_key=TEST1234567890123456\" | codehunter -r secrets.txt -v")
		fmt.Println()

		fmt.Printf("%s💡 Test Example (should find a match):%s\n", ColorYellow, ColorReset)
		fmt.Println("  echo \"https://example.com/api/v1/users\" | codehunter -r api_endpoints.txt --found-urls test_hits.txt")
		fmt.Println("  echo \"http://test.com/config.js?token=abcdef1234567890\" | codehunter -r js_secrets.txt -v")
		fmt.Println()

		fmt.Printf("%s💡 More Examples:%s\n", ColorYellow, ColorReset)
		fmt.Println("  katana -u tesla.com | codehunter -r secrets.txt,api_endpoints.txt --found-urls tesla_found.txt")
		fmt.Println()

		fmt.Printf("%s🔧 Flags:%s\n", ColorYellow, ColorReset)
		flag.PrintDefaults()
		fmt.Println()

		// ... (rest of the usage message like available patterns, installation, support) ...
		fmt.Printf("%s📋 Available Patterns (default location: patterns/ or /usr/share/codehunter/patterns/):%s\n", ColorYellow, ColorReset) //
		fmt.Println("  • secrets.txt      - API keys, tokens, credentials")                                                                    //
		fmt.Println("  • api_endpoints.txt - REST APIs, GraphQL, endpoints")                                                                   //
		fmt.Println("  • admin_panels.txt  - Admin areas, CMS panels")                                                                         //
		fmt.Println("  • js_secrets.txt   - JavaScript secrets, configs")                                                                      //
		fmt.Println("  • files.txt        - Sensitive files, backups")                                                                         //
		fmt.Println("  • custom.txt       - Your custom patterns")                                                                             //
		fmt.Println()

		fmt.Printf("%s🏴‍☠️ Happy Bug Hunting! 🏴‍☠️%s\n", ColorBold, ColorReset)
	}

	flag.StringVar(&config.PatternsFile, "r", "", "Patterns file(s), comma-separated (required)")
	flag.StringVar(&config.UrlsFile, "l", "", "URLs file (optional, uses stdin if not provided)")
	flag.StringVar(&config.OutputFile, "o", "", "Output file for matched URLs (legacy, use --found-urls for clarity)")
	flag.IntVar(&config.Threads, "t", 10, "Number of threads")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output (logs progress to stdout/--log-file)")
	flag.BoolVar(&config.ShowBanner, "b", true, "Show banner (default: true, set to false with -b=false)")
	flag.StringVar(&config.LogFile, "log-file", "", "File to write detailed scan log (optional)")
	flag.StringVar(&config.FoundUrlsLogFile, "found-urls", "", "File to write clean list of matched URLs (optional)")

	flag.Parse()

	if config.PatternsFile == "" {
		fmt.Printf("%s[ERROR]%s Patterns file is required! Use -r <patterns_file>%s\n", ColorRed, ColorReset, ColorReset)
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
    // If -o is used and --found-urls is not, make --found-urls same as -o for backward compatibility or preference
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
	var loadedPatterns []*regexp.Regexp
	var filesSuccessfullyProcessed int
	var patternLoadingLog strings.Builder // To accumulate log messages for patterns

	for _, sourceName := range patternFileSources {
		sourceName = strings.TrimSpace(sourceName)
		if sourceName == "" {
			continue
		}

		possiblePaths := []string{
			sourceName,
			filepath.Join("patterns", sourceName),
			filepath.Join("/usr/share/codehunter/patterns", sourceName), //
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
			pattern, errRegex := regexp.Compile(line)
			if errRegex != nil {
				msg := fmt.Sprintf("%s[WARN]%s Invalid regex at line %d in %s: %s. Skipping pattern.\n",
					ColorYellow, ColorReset, lineNum, usedPath, errRegex)
				patternLoadingLog.WriteString(msg)
				continue
			}
			loadedPatterns = append(loadedPatterns, pattern)
			patternsInThisFile++
		}
		
		if errScan := fileScanner.Err(); errScan != nil {
			msg := fmt.Sprintf("%s[WARN]%s Error reading file %s: %v. Patterns loaded may be incomplete.\n",
				ColorYellow, ColorReset, usedPath, errScan)
			patternLoadingLog.WriteString(msg)
		}
		file.Close() // Close file after processing it

		if patternsInThisFile > 0 {
			filesSuccessfullyProcessed++
		}
	}
    
	s.logDetailMessage(patternLoadingLog.String()) // Log all pattern loading messages

	s.Patterns = loadedPatterns
	s.Stats.PatternsCount = len(loadedPatterns)

	if len(s.Patterns) == 0 {
		return fmt.Errorf("no valid patterns loaded from any specified sources ('%s'). Please check pattern file paths and content", s.Config.PatternsFile)
	}

	if filesSuccessfullyProcessed < len(patternFileSources) && len(patternFileSources) > 1 {
		s.logDetailMessage(fmt.Sprintf("%s[INFO]%s Successfully processed patterns from %d out of %d specified source(s).\n",
			ColorCyan, ColorReset, filesSuccessfullyProcessed, len(patternFileSources)))
	}
	return nil
}

// ==============================================
// MAIN SCANNING LOGIC
// ==============================================
func (s *Scanner) scan(input io.Reader) {
	urlChan := make(chan string, s.Config.Threads*2)
	resultChan := make(chan string, s.Config.Threads*2) // This channel now only carries URLs to be written to foundFile

	var readerWg, writerWg, workerWg sync.WaitGroup

	// Start result writer goroutine (for found URLs file)
	if s.foundFile != nil { // Only start writer if foundFile is configured
		writerWg.Add(1)
		go func() {
			defer writerWg.Done()
			for resultURL := range resultChan {
				fmt.Fprintln(s.foundFile, resultURL)
			}
		}()
	} else { // If no foundFile, results will be printed to stdout by processURL if verbose
        // To prevent deadlock if resultChan is never read from, we need to drain it.
        // Or, ensure processURL doesn't write to resultChan if s.foundFile is nil.
        // For simplicity now, let's make processURL aware.
    }


	for i := 0; i < s.Config.Threads; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for url := range urlChan {
				s.processURL(url, resultChan, workerID)
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
			s.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Error reading URLs: %v\n", ColorRed, ColorReset, err))
		}
	}()

	readerWg.Wait()
	workerWg.Wait()
	
    close(resultChan) // Close resultChan after all workers are done.
	if s.foundFile != nil {
		writerWg.Wait() // Wait for writer only if it was started
	}


	s.Stats.EndTime = time.Now()
}

// ==============================================
// URL PROCESSING
// ==============================================
func (s *Scanner) processURL(url string, resultChan chan<- string, workerID int) {
	s.mu.Lock()
	s.Stats.URLsProcessed++
	currentProcessed := s.Stats.URLsProcessed
	s.mu.Unlock()

	if s.Config.Verbose && currentProcessed%100 == 0 {
		s.logDetailMessage(fmt.Sprintf("%s[PROGRESS]%s Processed %d URLs (Worker %d)\n",
			ColorBlue, ColorReset, currentProcessed, workerID))
	}

	matched := false
	for _, pattern := range s.Patterns {
		if pattern.MatchString(url) {
			matched = true
			break
		}
	}

	if matched {
		s.mu.Lock()
		s.Stats.URLsMatched++
		s.mu.Unlock()

		if s.Config.Verbose { // Verbose always logs to detail/stdout
			s.logDetailMessage(fmt.Sprintf("%s[MATCH]%s %s\n", ColorGreen, ColorReset, url))
		}
        
        if s.foundFile != nil { // If foundFile is configured, send to its channel
            resultChan <- url
        } else if !s.Config.Verbose { 
            // If no foundFile AND not verbose, print matched URL to stdout so user sees it
            // This maintains previous behavior of -o or stdout for matches
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

	// Tips
	if s.Stats.URLsMatched == 0 {
		statsBuilder.WriteString(fmt.Sprintf("\n%s💡 Tips for better results:%s\n", ColorYellow, ColorReset))
		statsBuilder.WriteString("  • Try different pattern files or combine them (e.g., -r secrets.txt,api_endpoints.txt)\n")
		statsBuilder.WriteString("  • Ensure your input URLs actually contain data that patterns might match.\n")
		statsBuilder.WriteString("  • Use -v for verbose progress, or --log-file for a persistent log.\n")
		statsBuilder.WriteString("  • Verify pattern file syntax and regex validity.\n")
	} else {
		statsBuilder.WriteString(fmt.Sprintf("\n%s🎉 Great! Found potential targets.%s\n", ColorGreen, ColorReset))
		if s.Config.FoundUrlsLogFile != "" { // If found URLs were saved to a file
			// This message is now printed at the end of main()
		} else { // If matches were printed to stdout
			statsBuilder.WriteString("  🔍 Review matched URLs printed above or use --found-urls <file> to save them.\n")
		}
		statsBuilder.WriteString("  🛡️ Follow responsible disclosure practices.\n")
	}

	statsBuilder.WriteString(fmt.Sprintf("\n%s🏴‍☠️ Made with ❤️ by %s (%s) | %s 🏴‍☠️%s\n",
		ColorBold, AUTHOR, TWITTER, GITHUB, ColorReset))
	statsBuilder.WriteString(fmt.Sprintf("%s🎯 Happy Bug Hunting! 🎯%s\n\n", ColorBold, ColorReset))

	// Write the stats to the detailed log file or stdout
	s.logDetailMessage(statsBuilder.String())
	if s.logDetailFile == nil && (s.Config.ShowBanner || s.Config.Verbose) { // if no log file, and banner/verbose is on, ensure it hits stdout
		fmt.Print(statsBuilder.String())
	} else if s.logDetailFile != nil && (!s.Config.ShowBanner && !s.Config.Verbose && s.Stats.URLsMatched > 0) {
        // If not verbose and no banner, but matches found and log file exists, also print minimal summary to stdout
        // This might be too noisy, consider if needed. For now, main prints file locations.
    }
}
