package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"

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
	methodFlag      = "method"
	timeoutFlag     = "timeout"
	http2Flag       = "http2"
	hostHeaderFlag  = "host"
	userAgentFlag   = "agent"
	basicAuthFlag   = "basic"
	headersFlag     = "headers"
	numberFlag      = "number"
	followFlag      = "follow"
	showCfgFlag     = "show"
	insecureFlag    = "insecure"
)

const (
	// HTTP Headers
	userAgentHeader = "User-Agent"
)

var (
	cfg     *config.Config
	showCfg bool
)

func init() {
	cfg = &config.Config{}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "vessel",
	Short:   "HTTP Benchmarking utility",
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.Endpoint = args[0]
		if showCfg {
			fmt.Println(cfg)
		}

		if cfg.Amount == 0 && cfg.Duration == 0 {
			return errors.New("-n or -d must not be zero when supplied")
		}

		// build the single req req to clone later.
		req, err := requester.GenerateTemplateRequest(cfg)
		if err != nil {
			return fmt.Errorf("unable to create request: %v", err)
		}

		// Ensure the endpoint is actual a valid URL
		// TODO: Do we want to enforce host/scheme specifics?
		_, err = url.ParseRequestURI(cfg.Endpoint)
		if err != nil {
			return fmt.Errorf("bad endpoint provided: %v", err)
		}

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
			req.Header = headers
		}

		// Disallow negative MaxRPS.
		if cmd.Flags().Changed(maxRPSFlag) {
			cfg.MaxRPS = max(0, cfg.MaxRPS)
		}

		// Disallow negative concurrency.
		if cmd.Flags().Changed(concurrencyFlag) {
			cfg.Concurrency = max(0, cfg.Concurrency)
		}

		// Handle basic auth if provided by the user
		if cmd.Flags().Changed(basicAuthFlag) {
			basicAuthUser, basicAuthPw, err := validation.ParseBasicAuth(cfg.BasicAuth)
			if err != nil {
				return err
			}
			req.SetBasicAuth(basicAuthUser, basicAuthPw)
		}

		// Handle custom host header if provided by the user
		// Host header has special treatment and is not a traditional header
		if cmd.Flags().Changed(hostHeaderFlag) {
			req.Host = cfg.Host
		}

		// Handle a custom user agent if provided by the user
		// the tool user agent is always appended for server tracability.
		uA := "vessel/" + Version
		if cmd.Flags().Changed(userAgentFlag) {
			cfg.UserAgent += fmt.Sprintf("%s ", uA)
		}
		req.Header.Set(userAgentHeader, cfg.UserAgent)

		collector := collector.New(out, cfg)

		// command ctx already has the signalling capabilities.
		// if duration is specified, wrap the ctx with that dead line
		// to cause Go() to exit and Wait() to unblock.
		parent := cmd.Context()
		ctx, cancel := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
		defer cancel()

		requester := requester.New(
			ctx,
			cfg,
			collector,
			req,
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
	rootCmd.Flags().BoolVarP(&cfg.QuietSet, quietFlag, "q", false, "Suppresses output")
	rootCmd.Flags().IntVarP(&cfg.MaxRPS, maxRPSFlag, "r", 0, "Rate limit requests per second")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, concurrencyFlag, "c", 10, "Number of concurrent requests")
	rootCmd.Flags().DurationVarP(&cfg.Duration, durationFlag, "d", 0, "Duration to send requests for (must be parsable by time.ParseDuration)")
	rootCmd.Flags().StringVarP(&cfg.Method, methodFlag, "m", "GET", "HTTP Verb to perform")
	rootCmd.Flags().DurationVarP(&cfg.Timeout, timeoutFlag, "t", 0, "Per Request timeout before terminating the request (must be parsable by time.ParseDuration)")
	rootCmd.Flags().BoolVar(&cfg.HTTP2, http2Flag, false, "Enable HTTP/2 support")
	rootCmd.Flags().StringVar(&cfg.Host, hostHeaderFlag, "", "Set a custom HOST header")
	rootCmd.Flags().StringVarP(&cfg.UserAgent, userAgentFlag, "u", "", "Set a custom user agent header, this is always suffixed with the tools user agent")
	rootCmd.Flags().StringVarP(&cfg.BasicAuth, basicAuthFlag, "b", "", "Colon separated user:pass for basic auth header")
	rootCmd.Flags().StringSliceVarP(&cfg.Headers, headersFlag, "H", make([]string, 0), "Colon separated header:value for arbitrary HTTP headers (appendable)")
	rootCmd.Flags().Int64VarP(&cfg.Amount, numberFlag, "n", 50, "The total number of requests, cannot be used with -d")
	rootCmd.Flags().BoolVarP(&cfg.FollowRedirects, followFlag, "f", true, "Automatically follow redirects")
	rootCmd.Flags().BoolVarP(&showCfg, showCfgFlag, "s", false, "Print cfg to stdout on startup")
	rootCmd.Flags().BoolVarP(&cfg.Insecure, insecureFlag, "i", false, "Do not verify server certificate and host name")

	// Specify required flags
	rootCmd.MarkFlagsMutuallyExclusive(durationFlag, numberFlag)

	// Only allow a single non flag argument, which is the url/endpoint.
	rootCmd.Args = cobra.ExactArgs(1)

	// Apply the current working version of vessel into the config
	cfg.Version = Version

}
