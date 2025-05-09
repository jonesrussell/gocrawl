package logger

import (
	"strings"

	"go.uber.org/fx/fxevent"
)

// fxLogger is a logger that implements fxevent.Logger.
type fxLogger struct {
	log Interface
}

// NewFxLogger creates a new fx logger.
func NewFxLogger(log Interface) fxevent.Logger {
	return &fxLogger{
		log: log,
	}
}

// cleanConstructorName cleans up a constructor name for logging
func cleanConstructorName(name string) string {
	// Remove package path
	if idx := strings.LastIndex(name, "/"); idx != -1 {
		name = name[idx+1:]
	}
	// Remove function suffix
	if idx := strings.Index(name, "("); idx != -1 {
		name = name[:idx]
	}
	// Remove fx annotations
	if idx := strings.Index(name, "fx."); idx != -1 {
		name = name[:idx]
	}
	return name
}

// LogEvent logs an fx event.
func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.Provided:
		// Only log important dependencies
		importantTypes := map[string]bool{
			"*article.Manager":      true,
			"*elasticsearch.Client": true,
			"*sources.Sources":      true,
			"*events.EventBus":      true,
			"signal.Interface":      true,
			"common.JobService":     true,
			"config.Interface":      true,
			"storage.Interface":     true,
			"content.Interface":     true,
			"crawler.Interface":     true,
		}

		for _, rtype := range e.OutputTypeNames {
			if !importantTypes[rtype] {
				continue
			}

			fields := []any{"type", rtype}
			if e.ConstructorName != "" {
				msg := "Initialized " + cleanConstructorName(e.ConstructorName)
				l.log.Debug(msg, fields...)
			} else {
				l.log.Debug("Provided", fields...)
			}
		}

		if e.Err != nil {
			l.log.Error("Error encountered while applying options", "error", e.Err)
		}

	case *fxevent.Invoked:
		l.log.Debug("Invoked",
			"function", cleanConstructorName(e.FunctionName),
		)
		if e.Err != nil {
			l.log.Error("Invoke failed", "error", e.Err)
		}

	case *fxevent.Run:
		l.log.Debug("Running",
			"function", cleanConstructorName(e.Name),
			"duration", e.Runtime,
		)
		if e.Err != nil {
			l.log.Error("Run failed", "error", e.Err)
		}

	case *fxevent.OnStartExecuting:
		l.log.Debug("Starting",
			"function", cleanConstructorName(e.FunctionName),
			"caller", cleanConstructorName(e.CallerName),
		)

	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.log.Error("Start failed",
				"function", cleanConstructorName(e.FunctionName),
				"caller", cleanConstructorName(e.CallerName),
				"duration", e.Runtime,
				"error", e.Err,
			)
		} else {
			l.log.Debug("Started",
				"function", cleanConstructorName(e.FunctionName),
				"caller", cleanConstructorName(e.CallerName),
				"duration", e.Runtime,
			)
		}

	case *fxevent.OnStopExecuting:
		l.log.Debug("Stopping",
			"function", cleanConstructorName(e.FunctionName),
			"caller", cleanConstructorName(e.CallerName),
		)

	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.log.Error("Stop failed",
				"function", cleanConstructorName(e.FunctionName),
				"caller", cleanConstructorName(e.CallerName),
				"duration", e.Runtime,
				"error", e.Err,
			)
		} else {
			l.log.Debug("Stopped",
				"function", cleanConstructorName(e.FunctionName),
				"caller", cleanConstructorName(e.CallerName),
				"duration", e.Runtime,
			)
		}

	case *fxevent.Started:
		if e.Err != nil {
			l.log.Error("Application start failed", "error", e.Err)
		} else {
			l.log.Info("Application started")
		}

	case *fxevent.Stopping:
		l.log.Info("Application stopping", "signal", e.Signal)

	case *fxevent.Stopped:
		if e.Err != nil {
			l.log.Error("Application stop failed", "error", e.Err)
		} else {
			l.log.Info("Application stopped")
		}

	case *fxevent.RolledBack:
		l.log.Error("Application start failed, rolling back", "error", e.Err)
	}
}
