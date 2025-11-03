package cmd

import (
	"backfill-tool/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	batchSize  int
	threads    int
	collection string
	csv        string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the backfill tool",
	Long:  `Run the backfill tool`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, World!")
		fmt.Println("Batch size:", batchSize)
		fmt.Println("Threads:", threads)
		fmt.Println("Collection:", collection)
		fmt.Println("Csv file path :", csv)
		internal.RunBatch(batchSize, threads, collection, csv)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVarP(&batchSize, "batch-size", "b", 1000, "Number of records per batch")
	runCmd.Flags().IntVarP(&threads, "threads", "t", 10, "Number of parallel threads")
	runCmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name to backfill")
	runCmd.Flags().StringVarP(&csv, "csv", "s", "", "CSV File path to take data from")
}
