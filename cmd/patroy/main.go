package main

import (
"fmt"
"os"

"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
Use:   "patroy [url]",
Short: "Patroy - Undetected stealth web scraper & clean Markdown extractor",
Long:  `Patroy converts dynamic web pages into clean, LLM-ready Markdown using Rod+Stealth and Go-Trafilatura.`,
Args:  cobra.ExactArgs(1),
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
