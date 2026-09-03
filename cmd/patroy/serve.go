package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marcuz-apl/patroy/internal/server"
	"github.com/marcuz-apl/patroy/pkg/patroy"
	"github.com/spf13/cobra"
)

var (
	flagServeHost string
	flagServePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Patroy scraping HTTP REST API microservice",
	Long:  `Launches a high-performance REST API microservice providing /health, /scrape, and /scrape/batch endpoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("%s:%d", flagServeHost, flagServePort)
		fmt.Fprintf(os.Stderr, "Starting Patroy API server on %s (version: %s)...\n", addr, version)

		clientOpts := []patroy.Option{
			patroy.WithHeadless(flagHeadless),
			patroy.WithFallbackHTTP(flagFallbackHTTP),
			patroy.WithConcurrency(flagConcurrency),
			patroy.WithIncludeCleanHTML(true),
			patroy.WithIncludeRawHTML(true),
		}

		client, err := patroy.NewClient(clientOpts...)
		if err != nil {
			return fmt.Errorf("initialize client: %w", err)
		}
		defer client.Close()

		srv := server.NewServer(client, version)

		// Handle graceful shutdown on OS signals
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-stopChan
			fmt.Fprintln(os.Stderr, "\nReceived shutdown signal, terminating gracefully...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
		}()

		fmt.Fprintf(os.Stderr, "Ready to receive scrape requests. Press Ctrl+C to stop.\n")
		if err := srv.Start(addr); err != nil && err.Error() != "http: Server closed" {
			return fmt.Errorf("server error: %w", err)
		}

		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&flagServeHost, "host", "0.0.0.0", "Network host interface to bind")
	serveCmd.Flags().IntVarP(&flagServePort, "port", "p", 4023, "HTTP port to listen on")

	rootCmd.AddCommand(serveCmd)
}
