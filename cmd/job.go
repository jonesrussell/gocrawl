/*
Copyright © 2025 Russell Jones <russell@web.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// Update the job command to accept scheduling parameters
var schedule string

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse the schedule and run the multi command accordingly
		duration, err := time.ParseDuration(schedule)
		if err != nil {
			fmt.Println("Invalid schedule duration:", err)
			return
		}

		// Create a context for cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Schedule the multi command using a ticker
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					// Call the multi command here
					fmt.Println("Running multi command...")
					// You may need to invoke the multi command function directly or use cobra to execute it
					// Example: multiCmd.ExecuteContext(ctx)
				case <-ctx.Done():
					return
				}
			}
		}()

		// Block the main goroutine to keep the application running
		select {}
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Define the schedule flag
	jobCmd.Flags().StringVar(&schedule, "schedule", "1m", "Schedule the job to run at a specified interval (e.g., 1m, 2h)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
