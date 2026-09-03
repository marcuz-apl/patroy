package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/marcuz-apl/patroy/pkg/patroy"
	"github.com/spf13/cobra"
)

var (
	version = "1.1.0"
	commit  = "none"
	date    = "unknown"
)

func init() {
	if version == "1.1.0" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
}

var (
	flagOutput          string
	flagFormat          string
	flagHeadless        bool
	flagWaitFor         string
	flagTimeout         time.Duration
	flagFallbackHTTP    bool
	flagSilent          bool
	flagScreenshot      string
	flagFullScreenshot  string
	flagPDF             string
	flagProxy           string
	flagProxyList       string
	flagProxyStrategy   string
	flagConcurrency     int
	flagDelay           time.Duration
	flagSchema          string
	flagBlockPrivateIPs bool
)

var rootCmd = &cobra.Command{
	Use:     "patroy <url|urls_file.txt>",
	Short:   "Patroy - Lightweight, high-performance web scraper & clean Markdown extractor",
	Long:    `Patroy converts dynamic, JavaScript-heavy, and client-rendered web pages into clean, LLM-ready Markdown, screenshots, PDFs, and structured data using Rod and Go-Trafilatura.`,
	Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetInput := args[0]

		// Check if input is a text file containing URLs
		var urls []string
		if fileInfo, err := os.Stat(targetInput); err == nil && !fileInfo.IsDir() {
			f, err := os.Open(targetInput)
			if err != nil {
				return fmt.Errorf("open URL file %s: %w", targetInput, err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "#") {
					if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
						line = "https://" + line
					}
					urls = append(urls, line)
				}
			}
		} else {
			targetURL := targetInput
			if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
				targetURL = "https://" + targetURL
			}
			urls = append(urls, targetURL)
		}

		if len(urls) == 0 {
			return fmt.Errorf("no URLs to scrape")
		}

		// Read proxies from proxy-list file if specified
		var proxies []string
		if flagProxyList != "" {
			pf, err := os.Open(flagProxyList)
			if err != nil {
				return fmt.Errorf("open proxy list file %s: %w", flagProxyList, err)
			}
			defer pf.Close()

			scanner := bufio.NewScanner(pf)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "#") {
					proxies = append(proxies, line)
				}
			}
		} else if flagProxy != "" {
			proxies = append(proxies, flagProxy)
		}

		opts := []patroy.Option{
			patroy.WithHeadless(flagHeadless),
			patroy.WithTimeout(flagTimeout),
			patroy.WithFallbackHTTP(flagFallbackHTTP),
			patroy.WithIncludeCleanHTML(true),
			patroy.WithIncludeRawHTML(true),
			patroy.WithConcurrency(flagConcurrency),
			patroy.WithDelay(flagDelay),
		}

		if len(proxies) > 0 {
			opts = append(opts, patroy.WithProxies(proxies, flagProxyStrategy))
		}

		if flagWaitFor != "" {
			opts = append(opts, patroy.WithWaitSelector(flagWaitFor))
		}

		if flagFullScreenshot != "" {
			opts = append(opts, patroy.WithScreenshot(true))
		} else if flagScreenshot != "" {
			opts = append(opts, patroy.WithScreenshot(false))
		}

		if flagPDF != "" {
			opts = append(opts, patroy.WithPDF(false))
		}

		if flagDelay > 0 {
			opts = append(opts, patroy.WithDelay(flagDelay))
		}

		if flagBlockPrivateIPs {
			opts = append(opts, patroy.WithBlockPrivateIPs(true))
		}

		if flagSchema != "" {
			rawSchema := flagSchema
			if _, err := os.Stat(flagSchema); err == nil {
				data, err := os.ReadFile(flagSchema)
				if err == nil {
					rawSchema = string(data)
				}
			}
			var parsedSchema map[string]interface{}
			if err := json.Unmarshal([]byte(rawSchema), &parsedSchema); err != nil {
				return fmt.Errorf("invalid schema JSON: %w", err)
			}
			opts = append(opts, patroy.WithSchema(parsedSchema))
		}

		client, err := patroy.NewClient(opts...)
		if err != nil {
			return fmt.Errorf("initialize client: %w", err)
		}
		defer client.Close()

		ctx := context.Background()

		// Single URL execution
		if len(urls) == 1 {
			targetURL := urls[0]
			if !flagSilent {
				fmt.Fprintf(os.Stderr, "Scraping %s with Patroy...\n", targetURL)
			}

			result, err := client.Scrape(ctx, targetURL, opts...)
			if err != nil {
				return fmt.Errorf("scrape failed: %w", err)
			}

			if !flagSilent {
				mode := "rod+stealth browser"
				if result.IsFallback {
					mode = "net/http fallback"
				}
				fmt.Fprintf(os.Stderr, "Extracted successfully via %s in %dms\n", mode, result.DurationMs)
			}

			// Save screenshot if requested
			screenshotPath := flagScreenshot
			if flagFullScreenshot != "" {
				screenshotPath = flagFullScreenshot
			}
			if screenshotPath != "" && len(result.Screenshot) > 0 {
				if err := os.WriteFile(screenshotPath, result.Screenshot, 0644); err != nil {
					return fmt.Errorf("write screenshot %s: %w", screenshotPath, err)
				}
				if !flagSilent {
					fmt.Fprintf(os.Stderr, "Screenshot saved to %s\n", screenshotPath)
				}
			}

			// Save PDF if requested
			if flagPDF != "" && len(result.PDF) > 0 {
				if err := os.WriteFile(flagPDF, result.PDF, 0644); err != nil {
					return fmt.Errorf("write PDF %s: %w", flagPDF, err)
				}
				if !flagSilent {
					fmt.Fprintf(os.Stderr, "PDF saved to %s\n", flagPDF)
				}
			}

			// Auto-detect format from file extension if -f was not explicitly specified
			effectiveFormat := strings.ToLower(flagFormat)
			if !cmd.Flags().Changed("format") && flagOutput != "" {
				switch filepath.Ext(strings.ToLower(flagOutput)) {
				case ".json":
					effectiveFormat = "json"
				case ".csv":
					effectiveFormat = "csv"
				case ".html", ".htm":
					effectiveFormat = "html"
				case ".md", ".markdown":
					effectiveFormat = "markdown"
				}
			}

			var outputContent string
			switch effectiveFormat {
			case "json":
				outputContent, err = result.ToFormattedJSON()
				if err != nil {
					return fmt.Errorf("format JSON: %w", err)
				}
			case "csv":
				if result.CSV != "" {
					outputContent = result.CSV
				} else {
					outputContent = fmt.Sprintf("URL,Title,Content\n%q,%q,%q\n", result.URL, result.Title, result.Markdown)
				}
			case "html":
				if result.CleanHTML != "" {
					outputContent = result.CleanHTML
				} else {
					outputContent = result.RawHTML
				}
			case "markdown", "md":
				fallthrough
			default:
				outputContent = result.Markdown
			}

			if flagOutput != "" {
				if err := os.WriteFile(flagOutput, []byte(outputContent), 0644); err != nil {
					return fmt.Errorf("write output file %s: %w", flagOutput, err)
				}
				if !flagSilent {
					fmt.Fprintf(os.Stderr, "Output written to %s\n", flagOutput)
				}
			} else {
				fmt.Println(outputContent)
			}

			return nil
		}

		// Batch URL execution
		if !flagSilent {
			fmt.Fprintf(os.Stderr, "Batch scraping %d URLs with concurrency=%d...\n", len(urls), flagConcurrency)
		}

		// If flagOutput is set, treat it as a directory for batch output
		outDir := flagOutput
		if outDir != "" {
			_ = os.MkdirAll(outDir, 0755)
		}

		stream := client.ScrapeMany(ctx, urls, opts...)
		idx := 0
		for item := range stream {
			idx++
			if item.Err != nil {
				fmt.Fprintf(os.Stderr, "[%d/%d] FAIL %s: %v\n", idx, len(urls), item.URL, item.Err)
				continue
			}

			if !flagSilent {
				fmt.Fprintf(os.Stderr, "[%d/%d] OK %s in %dms (%s)\n", idx, len(urls), item.URL, item.Result.DurationMs, item.Result.Title)
			}

			if outDir != "" {
				parsed, _ := url.Parse(item.URL)
				host := "page"
				if parsed != nil && parsed.Host != "" {
					host = parsed.Host + strings.ReplaceAll(parsed.Path, "/", "_")
				}
				ext := "md"
				var data string
				switch strings.ToLower(flagFormat) {
				case "json":
					ext = "json"
					data, _ = item.Result.ToFormattedJSON()
				case "csv":
					ext = "csv"
					if item.Result.CSV != "" {
						data = item.Result.CSV
					} else {
						data = fmt.Sprintf("URL,Title,Content\n%q,%q,%q\n", item.Result.URL, item.Result.Title, item.Result.Markdown)
					}
				case "html":
					ext = "html"
					data = item.Result.CleanHTML
				default:
					data = item.Result.Markdown
				}

				outFile := filepath.Join(outDir, fmt.Sprintf("%03d_%s.%s", idx, host, ext))
				_ = os.WriteFile(outFile, []byte(data), 0644)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write output to file (or directory if batch scraping)")
	rootCmd.Flags().StringVarP(&flagFormat, "format", "f", "markdown", "Output format: markdown, json, html, csv (auto-detected from -o extension)")
	rootCmd.Flags().BoolVar(&flagHeadless, "headless", true, "Run browser in headless mode")
	rootCmd.Flags().StringVar(&flagWaitFor, "wait-for", "", "Wait for CSS selector to appear in DOM before extracting")
	rootCmd.Flags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "Scraping timeout duration")
	rootCmd.Flags().BoolVar(&flagFallbackHTTP, "fallback-http", true, "Automatically fall back to net/http if browser navigation fails")
	rootCmd.Flags().BoolVar(&flagSilent, "silent", false, "Suppress progress and status messages on stderr")

	// Media flags
	rootCmd.Flags().StringVar(&flagScreenshot, "screenshot", "", "Capture viewport screenshot and save to image path (.png, .jpeg, .webp)")
	rootCmd.Flags().StringVar(&flagFullScreenshot, "full-screenshot", "", "Capture full-page screenshot and save to image path")
	rootCmd.Flags().StringVar(&flagPDF, "pdf", "", "Export page to PDF file")

	// Proxy flags
	rootCmd.Flags().StringVar(&flagProxy, "proxy", "", "Single HTTP/SOCKS5 proxy endpoint")
	rootCmd.Flags().StringVar(&flagProxyList, "proxy-list", "", "File containing proxy endpoints (one per line)")
	rootCmd.Flags().StringVar(&flagProxyStrategy, "proxy-strategy", "round-robin", "Proxy rotation strategy: round-robin, random, failover")

	// Batch concurrency flags
	rootCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 4, "Number of concurrent workers for batch scraping")
	rootCmd.Flags().DurationVar(&flagDelay, "delay", 0, "Polite delay between consecutive requests per domain (e.g. 500ms, 1s)")

	// Enterprise schema and security flags
	rootCmd.Flags().StringVar(&flagSchema, "schema", "", "Custom CSS extraction schema (JSON string or path to .json file)")
	rootCmd.Flags().BoolVar(&flagBlockPrivateIPs, "block-private-ips", false, "Block internal loopback, private networks, and cloud metadata (SSRF protection)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
