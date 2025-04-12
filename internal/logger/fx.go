package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// NewFxLogger creates a new Fx logger that uses our custom logger.
func NewFxLogger(logger *zap.Logger) *fxLogger {
	return &fxLogger{logger: logger}
}

type fxLogger struct {
	logger *zap.Logger
}

func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logger.Info("OnStart hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logger.Error("OnStart hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Info("OnStart hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		l.logger.Info("OnStop hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logger.Error("OnStop hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Info("OnStop hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logger.Error("Error encountered while applying options",
				zap.String("type", e.TypeName),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Info("Supplied",
				zap.String("type", e.TypeName),
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.logger.Info("Provided",
				zap.String("constructor", e.ConstructorName),
				zap.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logger.Error("Error encountered while applying options",
				zap.String("constructor", e.ConstructorName),
				zap.Error(e.Err),
			)
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.logger.Info("Replaced",
				zap.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logger.Error("Error encountered while replacing",
				zap.Error(e.Err),
			)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.logger.Info("Decorated",
				zap.String("decorator", e.DecoratorName),
				zap.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logger.Error("Error encountered while applying options",
				zap.String("decorator", e.DecoratorName),
				zap.Error(e.Err),
			)
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logger.Error("Invoke failed",
				zap.String("function", e.FunctionName),
				zap.Error(e.Err),
			)
		} else {
			l.logger.Info("Invoked",
				zap.String("function", e.FunctionName),
			)
		}
	case *fxevent.Stopping:
		l.logger.Info("Received signal",
			zap.String("signal", e.Signal.String()),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logger.Error("Stop failed",
				zap.Error(e.Err),
			)
		}
	case *fxevent.RolledBack:
		l.logger.Error("Start failed, rolling back",
			zap.Error(e.Err),
		)
	case *fxevent.Started:
		if e.Err != nil {
			l.logger.Error("Start failed",
				zap.Error(e.Err),
			)
		} else {
			l.logger.Info("Started application")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logger.Error("Custom logger initialization failed",
				zap.Error(e.Err),
			)
		} else {
			l.logger.Info("Initialized custom fxevent.Logger")
		}
	}
}
