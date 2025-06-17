package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.Endpoint = args[0]

		// build the single template templateRequest to clone later.
		templateRequest, err := http.NewRequest(
			cfg.Method,
			cfg.Endpoint,
			nil, // TODO: Allow body string or path to file for non GET
		)
		/*
			TODO:
				QuietSet: TODO: Suppress output and be aware of suppressing output throughout.
				MaxRPS: TODO: Somehow throttle max requests per second.
				Concurrency: TODO: fan out worker pool of concurrency count.
				Duration: TODO: Exit after fixed duration, smart use of contexts an proper cleanup.
				Output: TODO: Allow JSON/CSV resultsets, keep it extensible for future.
				Timeout: TODO: Per request timeouts (read etc).
				HTTP2: TODO: Enable http2 negotiation, careful we are using our own transport impl (not implicit).
				Host: TODO: Add Host header to requests.
				UserAgent: TODO: Append user defined user agent, default to something identifying the tool.
				BasicAuth: TODO: Add b64 authorisation basic auth header to requests.
		*/
		var out io.Writer = os.Stdout
		if cfg.QuietSet {
			out = io.Discard
		}

		collector := collector.New(out, &cfg)

		if err != nil {
			return fmt.Errorf("unable to create request: %v", err)
		}

		requester := requester.New(
			cfg,
			collector,
			templateRequest,
		)

		// command ctx already has the signalling capabilities.
		// if duration is specified, wrap the ctx with that dead line
		// to cause Go() to exit and Wait() to unblock.
		ctx := cmd.Context()
		var cancelFunc context.CancelFunc
		if d := cfg.Duration; d != 0 {
			ctx, cancelFunc = context.WithTimeout(ctx, d)
			defer cancelFunc()
		}
		requester.Go(ctx)
		requester.Wait()
		collector.Summarise()
		return nil
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
	rootCmd.Flags().StringVarP(&cfg.BasicAuth, "basic", "b", "", "Colon separated user:pass for basic auth header")

	// Specify required flags
	rootCmd.MarkFlagsOneRequired("concurrency", "duration")

	// Only allow a single non flag argument, which is the url/endpoint.
	rootCmd.Args = cobra.ExactArgs(1)

}
