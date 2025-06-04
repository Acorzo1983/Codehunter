//go:build linux || darwin
// +build linux darwin

// CodeHunter v2.5 - Ultra-Fast Bug Bounty Scanner
// Made with ❤️ by Albert.C @yz9yt
// https://github.com/Acorzo1983/Codehunter
// 🏴‍☠️ Exclusive for Kali Linux & Linux Distributions

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	VERSION = "2.5"
	BANNER  = `
╔══════════════════════════════════════════════════════╗
║                CodeHunter v2.5                      ║
║          Ultra-Fast Bug Bounty Scanner              ║
║              🏴‍☠️ Kali Linux Ready                   ║
║                                                      ║
║           Made with ❤️ by Albert.C @yz9yt           ║
║        https://github.com/Acorzo1983/Codehunter     ║
╚══════════════════════════════════════════════════════╝`
)

type Config struct {
	PatternsFile string
	URLsFile     string
	OutputFile   string
	Threads      int
	Verbose      bool
	ShowBanner   bool
}

type Result struct {
	URL     string `json:"url"`
	Pattern string `json:"pattern"`
	Match   string `json:"match"`
}

func main() {
	var config Config
	
	// Command line flags
	flag.StringVar(&config.PatternsFile, "r", "", "Patterns file (required)")
	flag.StringVar(&config.URLsFile, "l", "", "URLs file (optional, uses stdin if not provided)")
	flag.StringVar(&config.OutputFile, "o", "", "Output file (optional, uses stdout if not provided)")
	flag.IntVar(&config.Threads, "t", 10, "Number of threads")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.ShowBanner, "b", true, "Show banner")
	
	flag.Usage = func() {
		if config.ShowBanner {
			fmt.Println(BANNER)
		}
		fmt.Println("\n🎯 Usage:")
		fmt.Println("  codehunter -r patterns.txt -l urls.txt -o found.txt")
		fmt.Println("  katana -u tesla.com | codehunter -r secrets.txt")
		fmt.Println("  proxychains codehunter -r admin_panels.txt -l urls.txt")
		fmt.Println("\n📋 Flags:")
		flag.PrintDefaults()
		fmt.Println("\n🏴‍☠️ Made with ❤️ by Albert.C @yz9yt")
		fmt.Println("GitHub: https://github.com/Acorzo1983/Codehunter")
	}
	
	flag.Parse()
	
	// Show banner
	if config.ShowBanner {
		fmt.Println(BANNER)
		fmt.Printf("\n🚀 Starting CodeHunter v%s...\n", VERSION)
		fmt.Println("Made with ❤️ by Albert.C @yz9yt")
	}
	
	// Validate required flags
	if config.PatternsFile == "" {
		fmt.Println("❌ Error: Patterns file (-r) is required")
		flag.Usage()
		os.Exit(1)
	}
	
	// Load patterns
	patterns, err := loadPatterns(config.PatternsFile)
	if err != nil {
		fmt.Printf("❌ Error loading patterns: %v\n", err)
		os.Exit(1)
	}
	
	if config.Verbose {
		fmt.Printf("✅ Loaded %d patterns from %s\n", len(patterns), config.PatternsFile)
	}
	
	// Get URLs source
	var urlsReader io.Reader
	if config.URLsFile == "" || config.URLsFile == "stdin" {
		urlsReader = os.Stdin
		if config.Verbose {
			fmt.Println("📥 Reading URLs from stdin...")
		}
	} else {
		file, err := os.Open(config.URLsFile)
		if err != nil {
			fmt.Printf("❌ Error opening URLs file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		urlsReader = file
		if config.Verbose {
			fmt.Printf("📁 Reading URLs from %s\n", config.URLsFile)
		}
	}
	
	// Setup output
	var output io.Writer = os.Stdout
	if config.OutputFile != "" {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			fmt.Printf("❌ Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
		if config.Verbose {
			fmt.Printf("📝 Writing results to %s\n", config.OutputFile)
		}
	}
	
	// Process URLs
	results := processURLs(urlsReader, patterns, config)
	
	// Write results
	foundCount := 0
	for result := range results {
		fmt.Fprintln(output, result.URL)
		foundCount++
		
		if config.Verbose {
			fmt.Printf("🎯 Found: %s (pattern: %s)\n", result.URL, result.Pattern)
		}
	}
	
	if config.Verbose {
		fmt.Printf("\n✅ Scan complete! Found %d matching URLs\n", foundCount)
		fmt.Println("🏴‍☠️ Happy Bug Hunting!")
	}
}

func loadPatterns(filename string) ([]*regexp.Regexp, error) {
	// Try local file first, then system path
	paths := []string{
		filename,
		"patterns/" + filename,
		"/usr/share/codehunter/patterns/" + filename,
	}
	
	var file *os.File
	var err error
	
	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("could not find patterns file %s", filename)
	}
	defer file.Close()
	
	var patterns []*regexp.Regexp
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		pattern, err := regexp.Compile(line)
		if err != nil {
			fmt.Printf("⚠️  Warning: Invalid regex pattern: %s\n", line)
			continue
		}
		
		patterns = append(patterns, pattern)
	}
	
	return patterns, scanner.Err()
}

func processURLs(reader io.Reader, patterns []*regexp.Regexp, config Config) <-chan Result {
	results := make(chan Result, 100)
	urls := make(chan string, 100)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go worker(urls, patterns, results, &wg, config.Verbose)
	}
	
	// Read URLs
	go func() {
		defer close(urls)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" && !strings.HasPrefix(url, "#") {
				urls <- url
			}
		}
	}()
	
	// Close results when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()
	
	return results
}

func worker(urls <-chan string, patterns []*regexp.Regexp, results chan<- Result, wg *sync.WaitGroup, verbose bool) {
	defer wg.Done()
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	for url := range urls {
		if verbose {
			fmt.Printf("🔍 Scanning: %s\n", url)
		}
		
		resp, err := client.Get(url)
		if err != nil {
			if verbose {
				fmt.Printf("⚠️  Error fetching %s: %v\n", url, err)
			}
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			if verbose {
				fmt.Printf("⚠️  Error reading body %s: %v\n", url, err)
			}
			continue
		}
		
		content := string(body)
		
		// Check patterns
		for _, pattern := range patterns {
			if match := pattern.FindString(content); match != "" {
				results <- Result{
					URL:     url,
					Pattern: pattern.String(),
					Match:   match,
				}
				break // Only report first match per URL
			}
		}
	}
}
