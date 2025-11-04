package cmd

import (
	"backfill-tool/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	batchSize   int
	threads     int
	collection  string
	csv         string
	metricsFile string
	noProgress  bool
	bearerToken string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute API requests from Postman collection with CSV data",
	Long: `Execute bulk API requests using a Postman collection and CSV data.

This command reads a Postman collection (exported as JSON) and executes all
requests using data from a CSV file. Each row in the CSV becomes one request,
with CSV columns replacing {{variable}} placeholders in URLs, headers, and bodies.

The tool automatically:
  • Tracks progress with real-time metrics
  • Logs failed requests to a CSV file for retry
  • Saves execution metrics to a JSON file
  • Processes nested folders in collections
  • Uses concurrent workers for high performance
  • Supports authentication (Bearer, API Key, Basic)

Template Variables:
  Use {{columnName}} syntax in your Postman collection to reference CSV columns.
  Supported in: URLs, query parameters, headers, request bodies, and auth tokens.

Example Collection URL:
  https://api.example.com/users/{{userId}}/posts/{{postId}}?tag={{tag}}

Example CSV:
  userId,postId,tag
  123,456,important
  789,012,draft`,

	Example: `  # Basic usage with 10 concurrent workers
  backfill-tool run -c collection.json -s data.csv -t 10

  # With bearer token authentication (overrides collection auth)
  backfill-tool run -c collection.json -s data.csv -t 10 -a "your_token_here"

  # Bearer token from environment variable
  backfill-tool run -c collection.json -s data.csv -t 10 -a "$API_TOKEN"

  # High concurrency for internal APIs
  backfill-tool run -c collection.json -s data.csv -t 50

  # Conservative approach for rate-limited APIs
  backfill-tool run -c collection.json -s data.csv -t 2

  # Retry failed requests from previous run
  backfill-tool run -c collection.json -s failed_requests_20251103_114230.csv -t 5

  # Quiet mode for CI/CD (no progress bars)
  backfill-tool run -c collection.json -s data.csv -t 20 --quiet

  # Custom metrics file location
  backfill-tool run -c collection.json -s data.csv -t 10 --metrics-file ./results/metrics.json`,

	Run: func(cmd *cobra.Command, args []string) {
		// Get global flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		quiet, _ := cmd.Flags().GetBool("quiet")

		// Show startup info
		if !quiet {
			fmt.Println("🚀 Backfill Tool v2.3.0")
			fmt.Printf("📦 Collection: %s\n", collection)
			fmt.Printf("📊 CSV Data: %s\n", csv)
			fmt.Printf("⚙️  Workers: %d\n", threads)
			if metricsFile != "" {
				fmt.Printf("📈 Metrics: %s\n", metricsFile)
			}
			fmt.Println()
		}

		// Create run configuration
		config := internal.RunConfig{
			BatchSize:    batchSize,
			Threads:      threads,
			Collection:   collection,
			CSV:          csv,
			MetricsFile:  metricsFile,
			Verbose:      verbose,
			Quiet:        quiet,
			BearerToken:  bearerToken,
		}

		// Execute the batch run
		internal.RunBatch(config)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Required flags
	runCmd.Flags().StringVarP(&collection, "collection", "c", "", "Path to Postman collection JSON file (required)")
	runCmd.Flags().StringVarP(&csv, "csv", "s", "", "Path to CSV file with data (required)")
	runCmd.MarkFlagRequired("collection")
	runCmd.MarkFlagRequired("csv")

	// Optional flags with sensible defaults
	runCmd.Flags().IntVarP(&threads, "threads", "t", 10, "Number of concurrent worker threads (1-100)")
	runCmd.Flags().IntVarP(&batchSize, "batch-size", "b", 1000, "Number of records per batch (for future use)")

	// Output configuration
	runCmd.Flags().StringVarP(&metricsFile, "metrics-file", "m", "", "Path to save execution metrics JSON (default: metrics_<timestamp>.json)")
	runCmd.Flags().BoolVar(&noProgress, "no-progress", false, "Disable progress bars (deprecated: use --quiet instead)")

	// Authentication
	runCmd.Flags().StringVarP(&bearerToken, "bearer-token", "a", "", "Bearer token for authentication (overrides collection auth)")

	// Add examples to help
	runCmd.SetUsageTemplate(usageTemplate)
}

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
