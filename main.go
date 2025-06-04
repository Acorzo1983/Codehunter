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
	VERSION    = "2.5.6" // Incremented for build fix in CloseFiles
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
║                     🏴‍☠️ CodeHunter v2.5.6                    ║
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
	OutputFile       string
	Threads          int
	Verbose          bool
	ShowBanner       bool
	LogFile          string
	FoundUrlsLogFile string
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
	logDetailFile *os.File
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
	defer scanner.CloseFiles()

	if err := scanner.setupOutputFiles(); err != nil {
		fmt.Printf("%s[ERROR]%s Setting up output files: %v\n", ColorRed, ColorReset, err)
		os.Exit(1)
	}

	if err := scanner.loadPatterns(); err != nil {
		scanner.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Failed to load patterns: %v\n", ColorRed, ColorReset, err), true)
		os.Exit(1)
	}

	if config.Verbose {
		scanner.logDetailMessage(fmt.Sprintf("%s[INFO]%s Loaded %d patterns. Initial source: %s\n",
			ColorCyan, ColorReset, scanner.Stats.PatternsCount, config.PatternsFile), true)
	}

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

	scanner.scan(input)
	scanner.showFinalStats()

	if scanner.Config.FoundUrlsLogFile != "" {
		finalMsg := fmt.Sprintf("%s[INFO]%s Matched URLs saved to: %s\n", ColorGreen, ColorReset, scanner.Config.FoundUrlsLogFile)
		fmt.Print(finalMsg)
		if scanner.logDetailFile != nil && scanner.logDetailFile != os.Stdout {
			fmt.Fprint(scanner.logDetailFile, finalMsg)
		}
	}
	if scanner.Config.LogFile != "" {
		finalMsg := fmt.Sprintf("%s[INFO]%s Detailed scan log (with all matched patterns and occurrences) saved to: %s\n", ColorGreen, ColorReset, scanner.Config.LogFile)
		fmt.Print(finalMsg)
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
		s.foundFile = file
	}

	if s.Config.LogFile != "" {
		file, err := os.Create(s.Config.LogFile)
		if err != nil {
			if s.foundFile != nil {
				s.foundFile.Close()
			}
			return fmt.Errorf("creating detailed log file '%s': %w", s.Config.LogFile, err)
		}
		s.logDetailFile = file
	}
	return nil
}

// Corrected CloseFiles function
func (s *Scanner) CloseFiles() {
	if s.foundFile != nil {
		s.foundFile.Close()
	}
	if s.logDetailFile != nil {
		// s.logDetailFile is an *os.File created by os.Create, so it's safe to close.
		// It will not be os.Stdout based on current setupOutputFiles logic.
		s.logDetailFile.Close()
	}
}

func (s *Scanner) logDetailMessage(message string, forceToStdoutIfNotLogging bool) {
	if s.logDetailFile != nil {
		fmt.Fprint(s.logDetailFile, message)
	}
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
		if config.ShowBanner {
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

		possiblePaths := []string{
			sourceName,
			filepath.Join("patterns", sourceName),
			filepath.Join("/usr/share/codehunter/patterns", sourceName),
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
				msg := fmt.Sprintf("%s[WARN]%s Invalid regex at line %d in %s ('%s'): %s. Skipping pattern.\n",
					ColorYellow, ColorReset, lineNum, usedPath, line, errRegex)
				patternLoadingLog.WriteString(msg)
				continue
			}
			loadedPatterns = append(loadedPatterns, PatternInfo{
				RegexStr:   line,
				Compiled:   compiledPattern,
				SourceFile: baseSourceName,
			})
			patternsInThisFile++
		}

		if errScan := fileScanner.Err(); errScan != nil {
			msg := fmt.Sprintf("%s[WARN]%s Error reading file %s: %v. Patterns loaded from this file might be incomplete.\n",
				ColorYellow, ColorReset, usedPath, errScan)
			patternLoadingLog.WriteString(msg)
		}
		file.Close()

		if patternsInThisFile > 0 {
			filesSuccessfullyProcessed++
		}
	}
	s.logDetailMessage(patternLoadingLog.String(), true)

	s.Patterns = loadedPatterns
	s.Stats.PatternsCount = len(loadedPatterns)

	if len(s.Patterns) == 0 {
		return fmt.Errorf("no valid patterns loaded from any specified sources ('%s'). Please check pattern file paths and content", s.Config.PatternsFile)
	}

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
	resultChan := make(chan string, s.Config.Threads*2)

	var readerWg, writerWg, workerWg sync.WaitGroup

	if s.foundFile != nil {
		writerWg.Add(1)
		go func() {
			defer writerWg.Done()
			for resultURL := range resultChan {
				fmt.Fprintln(s.foundFile, resultURL)
			}
		}()
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
			s.logDetailMessage(fmt.Sprintf("%s[ERROR]%s Error reading URLs: %v\n", ColorRed, ColorReset, err), true)
		}
	}()

	readerWg.Wait()
	workerWg.Wait()
	close(resultChan)

	if s.foundFile != nil {
		writerWg.Wait()
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
			ColorBlue, ColorReset, currentProcessed, workerID), false)
	}

	var overallMatchForURL bool = false
	var allMatchesForThisURL []MatchDetail

	for _, pInfo := range s.Patterns {
		foundOccurrences := pInfo.Compiled.FindAllString(url, -1)

		if len(foundOccurrences) > 0 {
			overallMatchForURL = true
			allMatchesForThisURL = append(allMatchesForThisURL, MatchDetail{
				Pattern:     pInfo,
				Occurrences: foundOccurrences,
			})
		}
	}

	if overallMatchForURL {
		s.mu.Lock()
		s.Stats.URLsMatched++
		s.mu.Unlock()

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

			s.logFileMutex.Lock()
			fmt.Fprint(s.logDetailFile, logEntry.String())
			s.logFileMutex.Unlock()
		}

		if s.Config.Verbose {
			s.logDetailMessage(fmt.Sprintf("%s[MATCH]%s %s (Matched by %d patterns)\n",
				ColorGreen, ColorReset, url, len(allMatchesForThisURL)), false)
		}

		if s.foundFile != nil {
			resultChan <- url
		} else if !s.Config.Verbose {
			fmt.Println(url)
		}
	}
}

// ==============================================
// STATISTICS DISPLAY
// ==============================================
func (s *Scanner) showFinalStats() {
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

	if s.Stats.URLsMatched == 0 {
		statsBuilder.WriteString(fmt.Sprintf("\n%s💡 Tips for better results:%s\n", ColorYellow, ColorReset))
		statsBuilder.WriteString("  • Try different pattern files or combine them (e.g., -r secrets.txt,api_endpoints.txt)\n")
		statsBuilder.WriteString("  • Ensure your input URLs actually contain data that patterns might match.\n")
		statsBuilder.WriteString("  • Use -v for verbose progress, or --log-file for a persistent log with matched patterns and occurrences.\n")
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

	finalStatsMessage := statsBuilder.String()
	if s.logDetailFile != nil {
		fmt.Fprint(s.logDetailFile, finalStatsMessage)
		if s.Config.ShowBanner || s.Config.Verbose {
			if !strings.HasSuffix(finalStatsMessage, "\n\n") && !strings.HasSuffix(finalStatsMessage, "\n") {
				fmt.Println()
			}
			fmt.Print(finalStatsMessage)
		}
	} else {
		if s.Config.ShowBanner || s.Config.Verbose {
			fmt.Print(finalStatsMessage)
		}
	}
}
