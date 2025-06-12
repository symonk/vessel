package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/requester"
)

const (
	// Version holds the current version of the binary.
	// It is overridden at build time using -ldflags.
	Version = "v0.0.1"
)

var (
	// Defines flag variables for ease of use in commands.
	versionSet  bool
	quietSet    bool
	maxRPS      int
	concurrency int
	duration    time.Duration
	output      string
	method      string
	timeout     time.Duration
	http2       bool
	hostHeader  string
	userAgent   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vessel",
	Short: "HTTP Benchmarking utility",
	Run: func(cmd *cobra.Command, args []string) {
		collector := collector.New()
		requester := requester.New(collector,
			timeout,
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
	rootCmd.Flags().BoolVarP(&versionSet, "version", "v", false, "Shows the version of vessel")
	rootCmd.Flags().BoolVarP(&quietSet, "quiet", "q", false, "Suppresses output")
	rootCmd.Flags().IntVarP(&maxRPS, "max-rps", "r", 0, "Rate limit requests per second")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "Number of concurrent requests")
	rootCmd.Flags().DurationVarP(&duration, "duration", "d", 0, "Duration to send requests for")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "File format to output")
	rootCmd.Flags().StringVarP(&method, "method", "m", "GET", "HTTP Verb to perform")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "Requests before terminating the request")
	rootCmd.Flags().BoolVar(&http2, "http2", false, "Enable HTTP/2 support")
	rootCmd.Flags().StringVar(&hostHeader, "host", "", "Set a custom HOST header")
	rootCmd.Flags().StringVarP(&userAgent, "agent", "u", "", "Set a custom user agent header")

	// Specify required flags
	rootCmd.MarkFlagsOneRequired("concurrency", "duration")

}
