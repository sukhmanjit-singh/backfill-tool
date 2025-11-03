package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "backfill-tool",
	Short: "A tool for running API calls and backfilling data using parallet ",
	Long:  `A tool for backfilling data into a database`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
