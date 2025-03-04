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
	"os"
	"os/exec"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
)

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Schedule and run crawl-source crawl jobs",
	Long:  `Schedule and run crawl-source crawl jobs based on the times specified in sources.yml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load sources using our package
		s, err := sources.Load("sources.yml")
		if err != nil {
			globalLogger.Error("Error loading sources", "error", err)
			return
		}

		// Infinite loop to keep the application running
		for {
			now := time.Now()

			for _, source := range s.Sources {
				for _, t := range source.Time {
					scheduledTime, err := time.Parse("15:04", t)
					if err != nil {
						globalLogger.Error("Invalid time format for source", "source", source.Name, "time", t)
						continue
					}

					// Check if it's time to run the job
					if now.Hour() == scheduledTime.Hour() && now.Minute() == scheduledTime.Minute() {
						globalLogger.Info("Running scheduled crawl", "source", source.Name, "time", t)

						// Create a new command instance for the crawl command
						args := []string{"crawl", "--source", source.Name}
						cmd := exec.Command(os.Args[0], args...)
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr

						if err := cmd.Run(); err != nil {
							globalLogger.Error("Error executing crawl command", "source", source.Name, "error", err)
						}
					}
				}
			}

			// Sleep for a minute before checking again
			time.Sleep(1 * time.Minute)
		}
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)
}
