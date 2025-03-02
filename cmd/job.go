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
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Define a struct to hold source information
type Source struct {
	Name      string   `yaml:"name"`
	URL       string   `yaml:"url"`
	Index     string   `yaml:"index"`
	RateLimit string   `yaml:"rate_limit"`
	MaxDepth  int      `yaml:"max_depth"`
	Time      []string `yaml:"time"` // Change to a slice of strings
}

// Load sources from YAML file
func loadSources(filename string) ([]Source, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var sources struct {
		Sources []Source `yaml:"sources"`
	}
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, err
	}
	return sources.Sources, nil
}

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		sources, err := loadSources("sources.yml")
		if err != nil {
			fmt.Println("Error loading sources:", err)
			return
		}

		// Infinite loop to keep the application running
		for {
			now := time.Now()

			for _, source := range sources {
				for _, t := range source.Time {
					scheduledTime, err := time.Parse("15:04", t)
					if err != nil {
						fmt.Println("Invalid time format for source:", source.Name)
						continue
					}

					// Check if it's time to run the job
					if now.Hour() == scheduledTime.Hour() && now.Minute() == scheduledTime.Minute() {
						fmt.Printf("Running job for %s...\n", source.Name)
						// Call the multi command here
						// Example: multiCmd.ExecuteContext(ctx)
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
