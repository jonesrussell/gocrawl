package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gocrawl",
	Short: "A web crawler and search tool",
	Long: `gocrawl is a web crawler that stores data in Elasticsearch 
and provides search capabilities.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
