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
	"github.com/symonk/vessel/internal/validation"
)

// TODO: Wire in cobra auto completion

const (
	// Version holds the current version of the binary.
	// It is overridden at build time using -ldflags.
	Version = "v0.0.1"
)

const (
	// flag long names
	versionFlag     = "version"
	quietFlag       = "quiet"
	maxRPSFlag      = "max-rps"
	concurrencyFlag = "concurrency"
	durationFlag    = "duration"
	outputFlag      = "output"
	methodFlag      = "method"
	timeoutFlag     = "timeout"
	http2Flag       = "http2"
	hostHeaderFlag  = "host"
	userAgentFlag   = "agent"
	basicAuthFlag   = "basic"
	headersFlag     = "headers"
)

const (
	// HTTP Headers
	userAgentHeader = "User-Agent"
)

var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "vessel",
	Short:   "HTTP Benchmarking utility",
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.Endpoint = args[0]

		// build the single template templateRequest to clone later.
		templateRequest, err := http.NewRequest(
			cfg.Method,
			cfg.Endpoint,
			nil, // TODO: Allow body string or path to file for non GET
		)

		if err != nil {
			return fmt.Errorf("unable to create request: %v", err)
		}
		/*
			QuietSet: TODO: Suppress output and be aware of suppressing output throughout. [done]
			BasicAuth: TODO: Add b64 authorisation basic auth header to requests. [done]
			Host: TODO: Add Host header to requests. [done]
			UserAgent: TODO: Append user defined user agent, default to something identifying the tool. [done]
			Headers: TODO: Allow arbitrary `-H K:V` header value pairs.
			MaxRPS: TODO: Somehow throttle max requests per second.
			Concurrency: TODO: fan out worker pool of concurrency count.
			Duration: TODO: Exit after fixed duration, smart use of contexts an proper cleanup.
			Output: TODO: Allow JSON/CSV resultsets, keep it extensible for future.
			Timeout: TODO: Per request timeouts (read etc).
			HTTP2: TODO: Enable http2 negotiation, careful we are using our own transport impl (not implicit).
		*/

		// handle -q to suppress output if required.
		var out io.Writer = os.Stdout
		if cfg.QuietSet {
			out = io.Discard
		}

		// Append user provided HTTP headers if provided
		// -H can be provided multiple times.
		// Do this early so we can enforce the special case headers later.
		if cmd.Flags().Changed(headersFlag) {
			headers := validation.ParseHTTPHeaders(cfg.Headers).Clone()
			templateRequest.Header = headers
		}

		// Handle basic auth if provided by the user
		if cmd.Flags().Changed(basicAuthFlag) {
			basicAuthUser, basicAuthPw, err := validation.ParseBasicAuth(cfg.BasicAuth)
			if err != nil {
				return err
			}
			templateRequest.SetBasicAuth(basicAuthUser, basicAuthPw)
		}

		// Handle custom host header if provided by the user
		// Host header has special treatment and is not a traditional header
		if cmd.Flags().Changed(hostHeaderFlag) {
			templateRequest.Host = cfg.Host
		}

		// Handle a custom user agent if provided by the user
		// the tool user agent is always appended for server tracability.
		uA := "vessel/" + Version
		if cmd.Flags().Changed(userAgentFlag) {
			uA = fmt.Sprintf("%s ", uA)
		}
		templateRequest.Header.Set(userAgentHeader, uA)

		collector := collector.New(out, &cfg)

		// command ctx already has the signalling capabilities.
		// if duration is specified, wrap the ctx with that dead line
		// to cause Go() to exit and Wait() to unblock.
		ctx := cmd.Context()
		var cancelFunc context.CancelFunc
		if d := cfg.Duration; d != 0 {
			ctx, cancelFunc = context.WithTimeout(ctx, d)
			defer cancelFunc()
		}

		requester := requester.New(
			cfg,
			collector,
			templateRequest,
		)
		requester.Wait()
		collector.Summarise()
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func ExecuteContext(ctx context.Context) error {
	rootCmd.SetErrPrefix("fatal error: ")
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.Flags().BoolVarP(&cfg.VersionSet, versionFlag, "v", false, "Shows the version of vessel")
	rootCmd.Flags().BoolVarP(&cfg.QuietSet, quietFlag, "q", false, "Suppresses output")
	rootCmd.Flags().IntVarP(&cfg.MaxRPS, maxRPSFlag, "r", 0, "Rate limit requests per second")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, concurrencyFlag, "c", 10, "Number of concurrent requests")
	rootCmd.Flags().DurationVarP(&cfg.Duration, durationFlag, "d", 0, "Duration to send requests for")
	rootCmd.Flags().StringVarP(&cfg.Output, outputFlag, "o", "", "File format to output")
	rootCmd.Flags().StringVarP(&cfg.Method, methodFlag, "m", "GET", "HTTP Verb to perform")
	rootCmd.Flags().DurationVarP(&cfg.Timeout, timeoutFlag, "t", 0, "Requests before terminating the request")
	rootCmd.Flags().BoolVar(&cfg.HTTP2, http2Flag, false, "Enable HTTP/2 support")
	rootCmd.Flags().StringVar(&cfg.Host, hostHeaderFlag, "", "Set a custom HOST header")
	rootCmd.Flags().StringVarP(&cfg.UserAgent, userAgentFlag, "u", "", "Set a custom user agent header, this is always suffixed with the tools user agent")
	rootCmd.Flags().StringVarP(&cfg.BasicAuth, basicAuthFlag, "b", "", "Colon separated user:pass for basic auth header")
	rootCmd.Flags().StringSliceVarP(&cfg.Headers, headersFlag, "H", make([]string, 0), "Colon separated header:value for arbitrary HTTP headers (appendable)")

	// Specify required flags
	rootCmd.MarkFlagsMutuallyExclusive(concurrencyFlag, durationFlag)

	// Only allow a single non flag argument, which is the url/endpoint.
	rootCmd.Args = cobra.ExactArgs(1)

}
