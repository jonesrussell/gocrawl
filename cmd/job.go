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
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// JobParams holds the parameters for the job scheduler
type JobParams struct {
	fx.In

	Logger  common.Logger
	Sources common.Sources
}

// startJobScheduler starts the job scheduler and handles graceful shutdown
func startJobScheduler(p JobParams, rootCmd string) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle) {
		ctx, cancel := context.WithCancel(context.Background())

		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go runScheduler(ctx, p, rootCmd)
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})
	})
}

func runScheduler(ctx context.Context, p JobParams, rootCmd string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.Logger.Info("Job scheduler shutting down")
			return
		case now := <-ticker.C:
			checkAndRunJobs(p, rootCmd, now)
		}
	}
}

func checkAndRunJobs(p JobParams, rootCmd string, now time.Time) {
	for _, source := range p.Sources.Sources {
		for _, t := range source.Time {
			scheduledTime, parseErr := time.Parse("15:04", t)
			if parseErr != nil {
				p.Logger.Error("Error parsing time", "error", parseErr, "source", source.Name, "time", t)
				continue
			}

			p.Logger.Debug("Checking scheduled time",
				"source", source.Name,
				"current_time", now.Format("15:04"),
				"scheduled_time", t,
				"current_hour", now.Hour(),
				"scheduled_hour", scheduledTime.Hour(),
				"current_minute", now.Minute(),
				"scheduled_minute", scheduledTime.Minute(),
			)

			if now.Hour() == scheduledTime.Hour() && now.Minute() == scheduledTime.Minute() {
				p.Logger.Info("Running scheduled crawl",
					"source", source.Name,
					"time", t,
					"current_time", now.Format("15:04"),
				)

				if err := runCrawlCommand(rootCmd, source.Name, p.Logger); err != nil {
					p.Logger.Error("Error running crawl command",
						"error", err,
						"source", source.Name,
						"time", t,
					)
				}
			}
		}
	}
}

func runCrawlCommand(rootCmd, sourceName string, logger common.Logger) error {
	cmdArgs := []string{"crawl", sourceName}
	cmd := exec.Command(rootCmd, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug("Executing crawl command", "command", rootCmd, "args", cmdArgs)
	return cmd.Run()
}

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Schedule and run crawl jobs",
	Long:  `Schedule and run crawl jobs based on the times specified in sources.yml`,
	Run: func(cmd *cobra.Command, _ []string) {
		rootPath := cmd.Root().Name()

		app := fx.New(
			common.Module,
			fx.Invoke(func(p JobParams) {
				startJobScheduler(p, rootPath)
			}),
		)

		if err := app.Start(context.Background()); err != nil {
			common.PrintErrorf("Error starting application: %v", err)
			os.Exit(1)
		}

		// Wait for termination signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)
		defer func() {
			cancel()
			if err := app.Stop(ctx); err != nil {
				common.PrintErrorf("Error during shutdown: %v", err)
				os.Exit(1)
			}
		}()
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)
}
