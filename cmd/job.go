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
	Run: func(cmd *cobra.Command, _ []string) {
		// Get the root command path
		rootPath := cmd.Root().Name()

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
					scheduledTime, parseErr := time.Parse("15:04", t)
					if parseErr != nil {
						globalLogger.Error("Error parsing time", "error", parseErr)
						continue
					}

					// Add debug logging for time comparison
					globalLogger.Debug("Checking scheduled time",
						"source", source.Name,
						"current_time", now.Format("15:04"),
						"scheduled_time", t,
						"current_hour", now.Hour(),
						"scheduled_hour", scheduledTime.Hour(),
						"current_minute", now.Minute(),
						"scheduled_minute", scheduledTime.Minute(),
					)

					// Check if it's time to run the job
					if now.Hour() == scheduledTime.Hour() && now.Minute() == scheduledTime.Minute() {
						globalLogger.Info("Running scheduled crawl",
							"source", source.Name,
							"time", t,
							"current_time", now.Format("15:04"),
						)

						// This is safe because:
						// 1. rootPath is the binary name from cobra
						// 2. source.Name comes from the validated sources.yml file
						// 3. The command structure is fixed
						cmdArgs := []string{"crawl", "--source", source.Name}
						cmd := exec.Command(rootPath, cmdArgs...)
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr

						if runErr := cmd.Run(); runErr != nil {
							globalLogger.Error("Error running command", "error", runErr)
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
