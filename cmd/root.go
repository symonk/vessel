package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/requester"
)

const (
	// Version holds the current version of the binary.
	// It is overridden at build time using -ldflags.
	Version = "v0.0.1"
)

var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vessel",
	Short: "HTTP Benchmarking utility",
	Run: func(cmd *cobra.Command, args []string) {
		cfg.Endpoint = args[0]
		collector := collector.New(&cfg)
		req, _ := http.NewRequest(
			cfg.Method,
			cfg.Endpoint,
			nil, // TODO: Allow body string or path to file for non GET
		)
		requester := requester.New(collector,
			cfg.Timeout,
			req,
		)
		requester.Go()
		requester.Wait()
		fmt.Println(collector.Summarise())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.Flags().BoolVarP(&cfg.VersionSet, "version", "v", false, "Shows the version of vessel")
	rootCmd.Flags().BoolVarP(&cfg.QuietSet, "quiet", "q", false, "Suppresses output")
	rootCmd.Flags().IntVarP(&cfg.MaxRPS, "max-rps", "r", 0, "Rate limit requests per second")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 10, "Number of concurrent requests")
	rootCmd.Flags().DurationVarP(&cfg.Duration, "duration", "d", 0, "Duration to send requests for")
	rootCmd.Flags().StringVarP(&cfg.Output, "output", "o", "", "File format to output")
	rootCmd.Flags().StringVarP(&cfg.Method, "method", "m", "GET", "HTTP Verb to perform")
	rootCmd.Flags().DurationVarP(&cfg.Timeout, "timeout", "t", 0, "Requests before terminating the request")
	rootCmd.Flags().BoolVar(&cfg.HTTP2, "http2", false, "Enable HTTP/2 support")
	rootCmd.Flags().StringVar(&cfg.Host, "host", "", "Set a custom HOST header")
	rootCmd.Flags().StringVarP(&cfg.UserAgent, "agent", "u", "", "Set a custom user agent header")

	// Specify required flags
	rootCmd.MarkFlagsOneRequired("concurrency", "duration")

	// Only allow a single non flag argument, which is the url/endpoint.
	rootCmd.Args = cobra.ExactArgs(1)

}
