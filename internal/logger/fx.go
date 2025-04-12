package logger

import (
	"go.uber.org/fx/fxevent"
)

// fxLogger is a logger that implements fxevent.Logger.
type fxLogger struct {
	log Interface
}

// NewFxLogger creates a new fx logger.
func NewFxLogger(log Interface) fxevent.Logger {
	return &fxLogger{log: log}
}

// LogEvent logs an fx event.
func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.log.Debug("OnStart hook executing",
			"callee", e.FunctionName,
			"caller", e.CallerName,
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.log.Error("OnStart hook failed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("OnStart hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime,
			)
		}
	case *fxevent.OnStopExecuting:
		l.log.Debug("OnStop hook executing",
			"callee", e.FunctionName,
			"caller", e.CallerName,
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.log.Error("OnStop hook failed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("OnStop hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime,
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.log.Error("Error encountered while applying options",
				"type", e.TypeName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("Supplied",
				"type", e.TypeName,
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.log.Debug("Provided",
				"constructor", e.ConstructorName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			l.log.Error("Error encountered while applying options",
				"error", e.Err,
			)
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.log.Debug("Replaced",
				"type", rtype,
			)
		}
		if e.Err != nil {
			l.log.Error("Error encountered while replacing",
				"error", e.Err,
			)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.log.Debug("Decorated",
				"decorator", e.DecoratorName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			l.log.Error("Error encountered while applying options",
				"error", e.Err,
			)
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			l.log.Error("Invoke failed",
				"function", e.FunctionName,
				"error", e.Err,
			)
		} else {
			l.log.Debug("Invoked",
				"function", e.FunctionName,
			)
		}
	case *fxevent.Stopping:
		l.log.Info("Received signal",
			"signal", e.Signal,
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			l.log.Error("Stop failed",
				"error", e.Err,
			)
		}
	case *fxevent.RolledBack:
		l.log.Error("Start failed, rolling back",
			"error", e.Err,
		)
	case *fxevent.Started:
		if e.Err != nil {
			l.log.Error("Start failed",
				"error", e.Err,
			)
		} else {
			l.log.Info("Started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.log.Error("Custom logger initialization failed",
				"error", e.Err,
			)
		} else {
			l.log.Debug("Initialized custom fxevent.Logger",
				"function", e.ConstructorName,
			)
		}
	}
}
