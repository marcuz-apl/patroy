package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/marcuz-apl/patroy/pkg/patroy"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0-dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	if version == "0.1.0-dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
}

var (
	flagOutput       string
	flagFormat       string
	flagHeadless     bool
	flagWaitFor      string
	flagTimeout      time.Duration
	flagFallbackHTTP bool
	flagSilent       bool
)

var rootCmd = &cobra.Command{
	Use:     "patroy <url>",
	Short:   "Patroy - Undetected stealth web scraper & clean Markdown extractor",
	Long:    `Patroy converts dynamic, JavaScript-heavy, bot-defended web pages into clean, LLM-ready Markdown and structured data using Rod+Stealth and Go-Trafilatura.`,
	Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL := args[0]
		if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
			targetURL = "https://" + targetURL
		}

		if !flagSilent {
			fmt.Fprintf(os.Stderr, "Scraping %s with Patroy...\n", targetURL)
		}

		opts := []patroy.Option{
			patroy.WithHeadless(flagHeadless),
			patroy.WithTimeout(flagTimeout),
			patroy.WithFallbackHTTP(flagFallbackHTTP),
			patroy.WithIncludeCleanHTML(true),
			patroy.WithIncludeRawHTML(true),
		}

		if flagWaitFor != "" {
			opts = append(opts, patroy.WithWaitSelector(flagWaitFor))
		}

		client, err := patroy.NewClient(opts...)
		if err != nil {
			return fmt.Errorf("initialize client: %w", err)
		}
		defer client.Close()

		ctx := context.Background()
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

		var outputContent string
		switch strings.ToLower(flagFormat) {
		case "json":
			outputContent, err = result.ToFormattedJSON()
			if err != nil {
				return fmt.Errorf("format JSON: %w", err)
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
	},
}

func init() {
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write output to specified file path instead of stdout")
	rootCmd.Flags().StringVarP(&flagFormat, "format", "f", "markdown", "Output format: markdown, json, html")
	rootCmd.Flags().BoolVar(&flagHeadless, "headless", true, "Run browser in headless mode")
	rootCmd.Flags().StringVar(&flagWaitFor, "wait-for", "", "Wait for CSS selector to appear in DOM before extracting")
	rootCmd.Flags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "Scraping timeout duration")
	rootCmd.Flags().BoolVar(&flagFallbackHTTP, "fallback-http", true, "Automatically fall back to net/http if browser navigation fails")
	rootCmd.Flags().BoolVar(&flagSilent, "silent", false, "Suppress progress and status messages on stderr")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
