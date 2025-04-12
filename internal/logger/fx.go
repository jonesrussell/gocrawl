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
		l.logOnStartExecuting(e)
	case *fxevent.OnStartExecuted:
		l.logOnStartExecuted(e)
	case *fxevent.OnStopExecuting:
		l.logOnStopExecuting(e)
	case *fxevent.OnStopExecuted:
		l.logOnStopExecuted(e)
	case *fxevent.Supplied:
		l.logSupplied(e)
	case *fxevent.Provided:
		l.logProvided(e)
	case *fxevent.Replaced:
		l.logReplaced(e)
	case *fxevent.Decorated:
		l.logDecorated(e)
	case *fxevent.Invoked:
		l.logInvoked(e)
	case *fxevent.Stopped:
		l.logStopped(e)
	case *fxevent.RolledBack:
		l.logRolledBack(e)
	case *fxevent.Started:
		l.logStarted(e)
	}
}

func (l *fxLogger) logOnStartExecuting(e *fxevent.OnStartExecuting) {
	l.logger.Info("OnStart hook executing",
		zap.String("callee", e.FunctionName),
		zap.String("caller", e.CallerName),
	)
}

func (l *fxLogger) logOnStartExecuted(e *fxevent.OnStartExecuted) {
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
}

func (l *fxLogger) logOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.logger.Info("OnStop hook executing",
		zap.String("callee", e.FunctionName),
		zap.String("caller", e.CallerName),
	)
}

func (l *fxLogger) logOnStopExecuted(e *fxevent.OnStopExecuted) {
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
}

func (l *fxLogger) logSupplied(e *fxevent.Supplied) {
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
}

func (l *fxLogger) logProvided(e *fxevent.Provided) {
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
}

func (l *fxLogger) logReplaced(e *fxevent.Replaced) {
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
}

func (l *fxLogger) logDecorated(e *fxevent.Decorated) {
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
}

func (l *fxLogger) logInvoked(e *fxevent.Invoked) {
	if e.Err != nil {
		l.logger.Error("Error encountered while applying options",
			zap.String("function", e.FunctionName),
			zap.Error(e.Err),
		)
	} else {
		l.logger.Info("Invoked",
			zap.String("function", e.FunctionName),
		)
	}
}

func (l *fxLogger) logStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.logger.Error("Error encountered while stopping",
			zap.Error(e.Err),
		)
	} else {
		l.logger.Info("Stopped")
	}
}

func (l *fxLogger) logRolledBack(e *fxevent.RolledBack) {
	l.logger.Error("Start failed, rolling back",
		zap.Error(e.Err),
	)
}

func (l *fxLogger) logStarted(e *fxevent.Started) {
	l.logger.Info("Started")
}
