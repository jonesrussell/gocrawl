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

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// httpdCmd represents the httpd command
var httpdCmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the Fx application with the HTTP server
		app := fx.New(
			api.Module,
			fx.Provide(
				func() logger.Interface {
					return globalLogger // Use the global logger
				},
			),
		)

		// Start the application
		if err := app.Start(context.Background()); err != nil {
			fmt.Println("Error starting HTTP server:", err)
			return
		}

		// Wait for the application to stop
		defer app.Stop(context.Background())
		fmt.Println("HTTP server started on :8080")
	},
}

func init() {
	rootCmd.AddCommand(httpdCmd)

	// Here you can define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// httpdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// httpdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
