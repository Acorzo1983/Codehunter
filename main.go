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
	VERSION    = "2.5"
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
║                     🏴‍☠️ CodeHunter v2.5                      ║
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
	PatternsFile string
	UrlsFile     string
	OutputFile   string
	Threads      int
	Verbose      bool
	ShowBanner   bool
}

// ==============================================
// SCANNER STRUCTURE
// ==============================================
type Scanner struct {
	Config   Config
	Patterns []*regexp.Regexp
	Stats    ScanStats
	mu       sync.Mutex
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
	// Parse command line arguments
	config := parseFlags()

	// Show banner if enabled
	if config.ShowBanner {
		fmt.Println(BANNER)
		fmt.Printf("\n%s🚀 Starting CodeHunter v%s - Made with ❤️ by %s%s\n", 
			ColorGreen, VERSION, AUTHOR, ColorReset)
		fmt.Printf("%s📅 Build: %s | 🐹 Go: %s | 💻 OS: %s%s\n\n", 
			ColorBlue, BUILD_DATE, runtime.Version(), runtime.GOOS, ColorReset)
	}

	// Create and configure scanner
	scanner := &Scanner{
		Config: config,
		Stats: ScanStats{
			StartTime: time.Now(),
		},
	}

	// Load patterns
	if err := scanner.loadPatterns(); err != nil {
		fmt.Printf("%s[ERROR]%s Failed to load patterns: %v\n", ColorRed, ColorReset, err)
		os.Exit(1)
	}

	if config.Verbose {
		fmt.Printf("%s[INFO]%s Loaded %d patterns from %s\n", 
			ColorCyan, ColorReset, scanner.Stats.PatternsCount, config.PatternsFile)
	}

	// Setup input/output
	var input io.Reader
	var output io.Writer

	// Setup input (URLs source)
	if config.UrlsFile != "" {
		file, err := os.Open(config.UrlsFile)
		if err != nil {
			fmt.Printf("%s[ERROR]%s Cannot open URLs file: %v\n", ColorRed, ColorReset, err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
		if config.Verbose {
			fmt.Printf("%s[INFO]%s Reading URLs from: %s\n", ColorCyan, ColorReset, config.UrlsFile)
		}
	} else {
		input = os.Stdin
		if config.Verbose {
			fmt.Printf("%s[INFO]%s Reading URLs from stdin (pipe mode)\n", ColorCyan, ColorReset)
		}
	}

	// Setup output (results destination)
	if config.OutputFile != "" {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			fmt.Printf("%s[ERROR]%s Cannot create output file: %v\n", ColorRed, ColorReset, err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
		if config.Verbose {
			fmt.Printf("%s[INFO]%s Writing results to: %s\n", ColorCyan, ColorReset, config.OutputFile)
		}
	} else {
		output = os.Stdout
		if config.Verbose {
			fmt.Printf("%s[INFO]%s Writing results to stdout\n", ColorCyan, ColorReset)
		}
	}

	// Start scanning
	scanner.scan(input, output)

	// Show final statistics
	scanner.showFinalStats()
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

	// Custom usage function
	flag.Usage = func() {
		if config.ShowBanner {
			fmt.Println(BANNER)
		}
		fmt.Printf("\n%s🎯 CodeHunter v%s - Ultra-Fast Bug Bounty Scanner%s\n", ColorBold, VERSION, ColorReset)
		fmt.Printf("%sMade with ❤️ by %s (%s)%s\n\n", ColorPurple, AUTHOR, TWITTER, ColorReset)
		
		fmt.Printf("%s📋 Usage:%s\n", ColorYellow, ColorReset)
		fmt.Println("  codehunter -r patterns.txt -l urls.txt -o found.txt")
		fmt.Println("  katana -u tesla.com | codehunter -r secrets.txt")
		fmt.Println("  proxychains codehunter -r admin_panels.txt -l urls.txt")
		fmt.Println()
		
		fmt.Printf("%s💡 Examples:%s\n", ColorYellow, ColorReset)
		fmt.Println("  # Basic secret hunting")
		fmt.Println("  codehunter -r secrets.txt -l urls.txt -o secrets_found.txt")
		fmt.Println()
		fmt.Println("  # API endpoint discovery")
		fmt.Println("  codehunter -r api_endpoints.txt -l urls.txt -v")
		fmt.Println()
		fmt.Println("  # JavaScript analysis with pipeline")
		fmt.Println("  waybackurls tesla.com | grep '\\.js$' | codehunter -r js_secrets.txt")
		fmt.Println()
		fmt.Println("  # Complete Bug Bounty workflow")
		fmt.Println("  subfinder -d tesla.com | httpx | katana | codehunter -r secrets.txt,api_endpoints.txt")
		fmt.Println()
		fmt.Println("  # Anonymous scanning")
		fmt.Println("  proxychains codehunter -r admin_panels.txt -l targets.txt")
		fmt.Println()
		
		fmt.Printf("%s🔧 Flags:%s\n", ColorYellow, ColorReset)
		flag.PrintDefaults()
		fmt.Println()
		
		fmt.Printf("%s📋 Available Patterns:%s\n", ColorYellow, ColorReset)
		fmt.Println("  • secrets.txt      - API keys, tokens, credentials")
		fmt.Println("  • api_endpoints.txt - REST APIs, GraphQL, endpoints")
		fmt.Println("  • admin_panels.txt  - Admin areas, CMS panels")
		fmt.Println("  • js_secrets.txt   - JavaScript secrets, configs")
		fmt.Println("  • files.txt        - Sensitive files, backups")
		fmt.Println("  • custom.txt       - Your custom patterns")
		fmt.Println()
		
		fmt.Printf("%s⚡ Installation:%s\n", ColorYellow, ColorReset)
		fmt.Println("  git clone https://github.com/Acorzo1983/Codehunter.git && cd Codehunter && chmod +x installer.sh && ./installer.sh")
		fmt.Println()
		
		fmt.Printf("%s📞 Support:%s\n", ColorYellow, ColorReset)
		fmt.Printf("  🐙 GitHub: %s\n", GITHUB)
		fmt.Printf("  🐦 Twitter: %s\n", TWITTER)
		fmt.Println()
		
		fmt.Printf("%s🏴‍☠️ Happy Bug Hunting! 🏴‍☠️%s\n", ColorBold, ColorReset)
	}

	flag.StringVar(&config.PatternsFile, "r", "", "Patterns file (required)\n     Example: secrets.txt, api_endpoints.txt\n     Multiple: secrets.txt,admin_panels.txt")
	flag.StringVar(&config.UrlsFile, "l", "", "URLs file (optional, uses stdin if not provided)\n     Example: urls.txt, targets.txt")
	flag.StringVar(&config.OutputFile, "o", "", "Output file (optional, uses stdout if not provided)\n     Example: found.txt, results.txt")
	flag.IntVar(&config.Threads, "t", 10, "Number of threads for concurrent scanning\n     Example: -t 20 for faster scanning")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output (shows scanning progress)")
	flag.BoolVar(&config.ShowBanner, "b", true, "Show banner (default: true)")

	flag.Parse()

	// Validate required flags
	if config.PatternsFile == "" {
		fmt.Printf("%s[ERROR]%s Patterns file is required!\n", ColorRed, ColorReset)
		fmt.Printf("%s[TIP]%s Use: codehunter -r secrets.txt -l urls.txt\n", ColorYellow, ColorReset)
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	return config
}

// ==============================================
// PATTERN LOADING
// ==============================================
func (s *Scanner) loadPatterns() error {
	patternFiles := strings.Split(s.Config.PatternsFile, ",")
	
	for _, fileName := range patternFiles {
		fileName = strings.TrimSpace(fileName)
		
		// Try multiple locations for pattern files
		possiblePaths := []string{
			fileName,                                    // Direct path
			filepath.Join("patterns", fileName),         // Local patterns directory
			filepath.Join("/usr/share/codehunter/patterns", fileName), // System patterns
		}
		
		var file *os.File
		var err error
		var usedPath string
		
		for _, path := range possiblePaths {
			file, err = os.Open(path)
			if err == nil {
				usedPath = path
				break
			}
		}
		
		if err != nil {
			return fmt.Errorf("cannot open pattern file '%s' in any location", fileName)
		}
		defer file.Close()
		
		if s.Config.Verbose {
			fmt.Printf("%s[INFO]%s Loading patterns from: %s\n", ColorCyan, ColorReset, usedPath)
		}
		
		// Read patterns from file
		scanner := bufio.NewScanner(file)
		lineNum := 0
		
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			
			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			
			// Compile regex pattern
			pattern, err := regexp.Compile(line)
			if err != nil {
				if s.Config.Verbose {
					fmt.Printf("%s[WARN]%s Invalid regex at line %d in %s: %s\n", 
						ColorYellow, ColorReset, lineNum, fileName, err)
				}
				continue
			}
			
			s.Patterns = append(s.Patterns, pattern)
			s.Stats.PatternsCount++
		}
		
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading file %s: %v", fileName, err)
		}
	}
	
	if len(s.Patterns) == 0 {
		return fmt.Errorf("no valid patterns loaded")
	}
	
	return nil
}

// ==============================================
// MAIN SCANNING LOGIC
// ==============================================
func (s *Scanner) scan(input io.Reader, output io.Writer) {
	// Create channels for URL processing
	urlChan := make(chan string, s.Config.Threads*2)
	resultChan := make(chan string, s.Config.Threads*2)
	
	// Create wait groups
	var readerWg, writerWg, workerWg sync.WaitGroup
	
	// Start result writer goroutine
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		for result := range resultChan {
			fmt.Fprintln(output, result)
		}
	}()
	
	// Start worker goroutines
	for i := 0; i < s.Config.Threads; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			for url := range urlChan {
				s.processURL(url, resultChan, workerID)
			}
		}(i)
	}
	
	// Start URL reader goroutine
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
			fmt.Printf("%s[ERROR]%s Error reading URLs: %v\n", ColorRed, ColorReset, err)
		}
	}()
	
	// Wait for URL reading to complete
	readerWg.Wait()
	
	// Wait for all workers to finish
	workerWg.Wait()
	
	// Close result channel and wait for writer
	close(resultChan)
	writerWg.Wait()
	
	// Record end time
	s.Stats.EndTime = time.Now()
}

// ==============================================
// URL PROCESSING
// ==============================================
func (s *Scanner) processURL(url string, resultChan chan<- string, workerID int) {
	// Update stats
	s.mu.Lock()
	s.Stats.URLsProcessed++
	currentProcessed := s.Stats.URLsProcessed
	s.mu.Unlock()
	
	// Verbose progress reporting
	if s.Config.Verbose && currentProcessed%100 == 0 {
		fmt.Printf("%s[PROGRESS]%s Processed %d URLs (Worker %d)\n", 
			ColorBlue, ColorReset, currentProcessed, workerID)
	}
	
	// Check URL against all patterns
	matched := false
	for _, pattern := range s.Patterns {
		if pattern.MatchString(url) {
			matched = true
			break
		}
	}
	
	// If matched, send to results
	if matched {
		s.mu.Lock()
		s.Stats.URLsMatched++
		s.mu.Unlock()
		
		if s.Config.Verbose {
			fmt.Printf("%s[MATCH]%s %s\n", ColorGreen, ColorReset, url)
		}
		
		resultChan <- url
	}
}

// ==============================================
// STATISTICS DISPLAY
// ==============================================
func (s *Scanner) showFinalStats() {
	if !s.Config.Verbose && !s.Config.ShowBanner {
		return
	}
	
	duration := s.Stats.EndTime.Sub(s.Stats.StartTime)
	
	fmt.Printf("\n%s╔══════════════════════════════════════════════════════════╗%s\n", ColorPurple, ColorReset)
	fmt.Printf("%s║                    🏴‍☠️ HUNT COMPLETE 🏴‍☠️                  ║%s\n", ColorPurple, ColorReset)
	fmt.Printf("%s╠══════════════════════════════════════════════════════════╣%s\n", ColorPurple, ColorReset)
	fmt.Printf("%s║%s  📊 URLs Processed: %s%-10d%s                     %s║%s\n", 
		ColorPurple, ColorReset, ColorCyan, s.Stats.URLsProcessed, ColorReset, ColorPurple, ColorReset)
	fmt.Printf("%s║%s  🎯 URLs Matched:   %s%-10d%s                     %s║%s\n", 
		ColorPurple, ColorReset, ColorGreen, s.Stats.URLsMatched, ColorReset, ColorPurple, ColorReset)
	fmt.Printf("%s║%s  🔍 Patterns Used:  %s%-10d%s                     %s║%s\n", 
		ColorPurple, ColorReset, ColorYellow, s.Stats.PatternsCount, ColorReset, ColorPurple, ColorReset)
	fmt.Printf("%s║%s  ⏱️ Duration:       %s%-10s%s                     %s║%s\n", 
		ColorPurple, ColorReset, ColorBlue, duration.Truncate(time.Millisecond).String(), ColorReset, ColorPurple, ColorReset)
	fmt.Printf("%s║%s  🚀 Speed:          %s%-10.1f URLs/sec%s           %s║%s\n", 
		ColorPurple, ColorReset, ColorCyan, float64(s.Stats.URLsProcessed)/duration.Seconds(), ColorReset, ColorPurple, ColorReset)
	fmt.Printf("%s╠══════════════════════════════════════════════════════════╣%s\n", ColorPurple, ColorReset)
	
	// Match rate indicator
	var matchStatus, matchColor string
	if s.Stats.URLsMatched == 0 {
		matchStatus = "No matches found"
		matchColor = ColorYellow
	} else if s.Stats.URLsMatched < 10 {
		matchStatus = "Low match rate"
		matchColor = ColorYellow
	} else {
		matchStatus = "Good match rate"
		matchColor = ColorGreen
	}
	
	fmt.Printf("%s║%s  🎯 Status:         %s%-25s%s %s║%s\n", 
		ColorPurple, ColorReset, matchColor, matchStatus, ColorReset, ColorPurple, ColorReset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════╝%s\n", ColorPurple, ColorReset)
	
	// Show tips based on results
	if s.Stats.URLsMatched == 0 {
		fmt.Printf("\n%s💡 Tips for better results:%s\n", ColorYellow, ColorReset)
		fmt.Println("  • Try different pattern files")
		fmt.Println("  • Check if URLs contain the target patterns")
		fmt.Println("  • Use -v flag for verbose output")
		fmt.Println("  • Verify pattern file syntax")
	} else if s.Stats.URLsMatched > 0 {
		fmt.Printf("\n%s🎉 Great! Found potential targets:%s\n", ColorGreen, ColorReset)
		if s.Config.OutputFile != "" {
			fmt.Printf("  📁 Results saved to: %s\n", s.Config.OutputFile)
		}
		fmt.Println("  🔍 Review each URL manually")
		fmt.Println("  🛡️ Follow responsible disclosure")
	}
	
	fmt.Printf("\n%s🏴‍☠️ Made with ❤️ by %s (%s) | %s 🏴‍☠️%s\n", 
		ColorBold, AUTHOR, TWITTER, GITHUB, ColorReset)
	fmt.Printf("%s🎯 Happy Bug Hunting! 🎯%s\n\n", ColorBold, ColorReset)
}

// ==============================================
// ERROR HANDLING
// ==============================================
func init() {
	// Handle panics gracefully
	if r := recover(); r != nil {
		fmt.Printf("%s[PANIC]%s CodeHunter crashed: %v\n", ColorRed, ColorReset, r)
		fmt.Printf("%s[HELP]%s Report this at: %s\n", ColorYellow, ColorReset, GITHUB)
		os.Exit(1)
	}
}
