package main

import (
	"fmt"
	"os"
	"runtime/debug"

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

var rootCmd = &cobra.Command{
	Use:     "patroy [url]",
	Short:   "Patroy - Undetected stealth web scraper & clean Markdown extractor",
	Long:    `Patroy converts dynamic web pages into clean, LLM-ready Markdown using Rod+Stealth and Go-Trafilatura.`,
	Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetURL := args[0]
		fmt.Printf("Scraping %s with Patroy...\n", targetURL)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
