package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "2.2.0"

var (
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "backfill-tool",
	Short: "High-performance CLI for bulk API operations using Postman collections",
	Long: `Backfill Tool - Production-ready CLI for bulk API operations

A powerful command-line tool for executing large-scale API requests using
Postman collections and CSV data. Perfect for data migration, backfilling,
bulk testing, and API automation tasks.

Features:
  • Execute Postman collections with CSV data
  • Template variables in URLs, headers, and bodies ({{variable}})
  • Concurrent request processing with configurable workers
  • Real-time progress tracking with metrics
  • Automatic logging of failed requests
  • Support for all HTTP methods (GET, POST, PUT, PATCH, DELETE, etc.)
  • Nested folder support in Postman collections

Quick Start:
  1. Export your Postman collection to JSON
  2. Prepare a CSV file with your data
  3. Run: backfill-tool run -c collection.json -s data.csv -t 10

Documentation: https://github.com/sukhmanjit-singh/backfill-tool`,
	Example: `  # Basic usage with 10 concurrent workers
  backfill-tool run -c api-collection.json -s users.csv -t 10

  # High concurrency for bulk operations
  backfill-tool run -c collection.json -s data.csv -t 50

  # Quiet mode for CI/CD pipelines
  backfill-tool run -c collection.json -s data.csv -t 20 --quiet

  # Show version information
  backfill-tool version`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the current version of backfill-tool and build information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Backfill Tool v%s\n", version)
		fmt.Println("\nFeatures:")
		fmt.Println("  ✓ Postman Collection Support")
		fmt.Println("  ✓ CSV Data Integration")
		fmt.Println("  ✓ Template Variable Replacement")
		fmt.Println("  ✓ Concurrent Execution")
		fmt.Println("  ✓ Progress Tracking & Metrics")
		fmt.Println("  ✓ Failed Request Logging")
		fmt.Println("  ✓ Nested Folder Support")
		fmt.Println("\nGo version: 1.21+")
		fmt.Println("Repository: https://github.com/sukhmanjit-singh/backfill-tool")
	},
}

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Show detailed usage examples",
	Long:  `Display comprehensive examples for common use cases with backfill-tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Backfill Tool - Usage Examples")
		fmt.Println("================================\n")

		fmt.Println("1. SIMPLE POST REQUEST")
		fmt.Println("   CSV file (users.csv):")
		fmt.Println("     name,email,age")
		fmt.Println("     John,john@example.com,30")
		fmt.Println("")
		fmt.Println("   Postman URL: https://api.example.com/users")
		fmt.Println("   Postman Body: {\"name\": \"{{name}}\", \"email\": \"{{email}}\"}")
		fmt.Println("")
		fmt.Println("   Command:")
		fmt.Println("     backfill-tool run -c collection.json -s users.csv -t 10")
		fmt.Println("")

		fmt.Println("2. GET REQUEST WITH PATH VARIABLES")
		fmt.Println("   CSV file (ids.csv):")
		fmt.Println("     userId")
		fmt.Println("     123")
		fmt.Println("     456")
		fmt.Println("")
		fmt.Println("   Postman URL: https://api.example.com/users/{{userId}}")
		fmt.Println("")
		fmt.Println("   Command:")
		fmt.Println("     backfill-tool run -c collection.json -s ids.csv -t 5")
		fmt.Println("")

		fmt.Println("3. QUERY PARAMETERS")
		fmt.Println("   CSV file (search.csv):")
		fmt.Println("     query,limit")
		fmt.Println("     smartphones,10")
		fmt.Println("     laptops,20")
		fmt.Println("")
		fmt.Println("   Postman URL: https://api.example.com/search?q={{query}}&limit={{limit}}")
		fmt.Println("")
		fmt.Println("   Command:")
		fmt.Println("     backfill-tool run -c collection.json -s search.csv -t 10")
		fmt.Println("")

		fmt.Println("4. DYNAMIC HEADERS")
		fmt.Println("   CSV file (auth.csv):")
		fmt.Println("     token,userId")
		fmt.Println("     abc123,user1")
		fmt.Println("     def456,user2")
		fmt.Println("")
		fmt.Println("   Postman Header: Authorization: Bearer {{token}}")
		fmt.Println("   Postman URL: https://api.example.com/users/{{userId}}/profile")
		fmt.Println("")
		fmt.Println("   Command:")
		fmt.Println("     backfill-tool run -c collection.json -s auth.csv -t 10")
		fmt.Println("")

		fmt.Println("5. RETRY FAILED REQUESTS")
		fmt.Println("   After a run, retry only the failed requests:")
		fmt.Println("")
		fmt.Println("   Command:")
		fmt.Println("     backfill-tool run -c collection.json -s failed_requests_20251103_114230.csv -t 5")
		fmt.Println("")

		fmt.Println("6. QUIET MODE FOR CI/CD")
		fmt.Println("   Run without progress bars (suitable for logs):")
		fmt.Println("")
		fmt.Println("   Command:")
		fmt.Println("     backfill-tool run -c collection.json -s data.csv -t 20 --quiet")
		fmt.Println("")

		fmt.Println("For more information, visit:")
		fmt.Println("https://github.com/sukhmanjit-singh/backfill-tool")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output with detailed logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode - suppress progress bars (useful for CI/CD)")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(examplesCmd)
}
