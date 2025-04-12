package logger

import (
	"reflect"
	"time"

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

// logSuccessOrError is a helper method for logging success or error events.
func (l *fxLogger) logSuccessOrError(successMsg, errorMsg string, err error, fields ...interface{}) {
	if err != nil {
		l.log.Error(errorMsg, append(fields, "error", err)...)
	} else {
		l.log.Debug(successMsg, fields...)
	}
}

// hookEvent represents an event with function and caller information.
type hookEvent interface {
	GetFunctionName() string
	GetCallerName() string
}

// errorEvent represents an event with an error field.
type errorEvent interface {
	GetErr() error
}

// multiTypeEvent represents an event with multiple types and an error.
type multiTypeEvent interface {
	errorEvent
	GetOutputTypeNames() []string
}

// logHookEvent is a helper method for logging hook events (OnStart/OnStop).
func (l *fxLogger) logHookEvent(prefix string, e hookEvent, err error, runtime time.Duration) {
	l.logSuccessOrError(
		prefix+" hook executed",
		prefix+" hook failed",
		err,
		"callee", e.GetFunctionName(),
		"caller", e.GetCallerName(),
		"runtime", runtime,
	)
}

// logTypeEvent is a helper method for logging events with type information.
func (l *fxLogger) logTypeEvent(msg string, e errorEvent, typeName string) {
	l.logSuccessOrError(
		msg,
		"Error encountered while "+msg,
		e.GetErr(),
		"type", typeName,
	)
}

// logMultiTypeEvent is a helper method for logging events with multiple types.
func (l *fxLogger) logMultiTypeEvent(msg string, e multiTypeEvent, additionalFields ...interface{}) {
	for _, rtype := range e.GetOutputTypeNames() {
		l.log.Debug(msg, append(additionalFields, "type", rtype)...)
	}
	if e.GetErr() != nil {
		l.log.Error("Error encountered while "+msg, append(additionalFields, "error", e.GetErr())...)
	}
}

// LogOnStartExecuting logs the OnStartExecuting event.
func (l *fxLogger) LogOnStartExecuting(e *fxevent.OnStartExecuting) {
	l.log.Debug("OnStart hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

// LogOnStartExecuted logs the OnStartExecuted event.
func (l *fxLogger) LogOnStartExecuted(e *fxevent.OnStartExecuted) {
	l.logSuccessOrError(
		"OnStart hook executed",
		"OnStart hook failed",
		e.Err,
		"callee", e.FunctionName,
		"caller", e.CallerName,
		"runtime", e.Runtime,
	)
}

// LogOnStopExecuting logs the OnStopExecuting event.
func (l *fxLogger) LogOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.log.Debug("OnStop hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

// LogOnStopExecuted logs the OnStopExecuted event.
func (l *fxLogger) LogOnStopExecuted(e *fxevent.OnStopExecuted) {
	l.logSuccessOrError(
		"OnStop hook executed",
		"OnStop hook failed",
		e.Err,
		"callee", e.FunctionName,
		"caller", e.CallerName,
		"runtime", e.Runtime,
	)
}

// LogSupplied logs the Supplied event.
func (l *fxLogger) LogSupplied(e *fxevent.Supplied) {
	l.logSuccessOrError(
		"Supplied",
		"Error encountered while applying options",
		e.Err,
		"type", e.TypeName,
	)
}

// LogProvided logs the Provided event.
func (l *fxLogger) LogProvided(e *fxevent.Provided) {
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
}

// LogReplaced logs the Replaced event.
func (l *fxLogger) LogReplaced(e *fxevent.Replaced) {
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
}

// LogDecorated logs the Decorated event.
func (l *fxLogger) LogDecorated(e *fxevent.Decorated) {
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
}

// LogInvoked logs the Invoked event.
func (l *fxLogger) LogInvoked(e *fxevent.Invoked) {
	l.logSuccessOrError(
		"Invoked",
		"Invoke failed",
		e.Err,
		"function", e.FunctionName,
	)
}

// LogStopping logs the Stopping event.
func (l *fxLogger) LogStopping(e *fxevent.Stopping) {
	l.log.Info("Received signal",
		"signal", e.Signal,
	)
}

// LogStopped logs the Stopped event.
func (l *fxLogger) LogStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.log.Error("Stop failed",
			"error", e.Err,
		)
	}
}

// LogRolledBack logs the RolledBack event.
func (l *fxLogger) LogRolledBack(e *fxevent.RolledBack) {
	l.log.Error("Start failed, rolling back",
		"error", e.Err,
	)
}

// LogStarted logs the Started event.
func (l *fxLogger) LogStarted(e *fxevent.Started) {
	if e.Err != nil {
		l.log.Error("Start failed",
			"error", e.Err,
		)
	} else {
		l.log.Info("Started")
	}
}

// LogLoggerInitialized logs the LoggerInitialized event.
func (l *fxLogger) LogLoggerInitialized(e *fxevent.LoggerInitialized) {
	l.logSuccessOrError(
		"Initialized custom fxevent.Logger",
		"Custom logger initialization failed",
		e.Err,
		"function", e.ConstructorName,
	)
}

// eventHandler defines the interface for handling fx events.
type eventHandler interface {
	Handle(log Interface, event fxevent.Event)
}

// multiTypeEventHandler handles events with multiple types.
type multiTypeEventHandler struct {
	successMsg string
	errorMsg   string
}

// Handle implements the eventHandler interface for multi-type events.
func (h *multiTypeEventHandler) Handle(log Interface, event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			log.Debug(h.successMsg,
				"constructor", e.ConstructorName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			log.Error(h.errorMsg,
				"error", e.Err,
			)
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			log.Debug(h.successMsg,
				"type", rtype,
			)
		}
		if e.Err != nil {
			log.Error(h.errorMsg,
				"error", e.Err,
			)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			log.Debug(h.successMsg,
				"decorator", e.DecoratorName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			log.Error(h.errorMsg,
				"error", e.Err,
			)
		}
	}
}

// hookEventHandler handles hook events (OnStart/OnStop).
type hookEventHandler struct {
	prefix string
}

// Handle implements the eventHandler interface for hook events.
func (h *hookEventHandler) Handle(log Interface, event fxevent.Event) {
	switch event.(type) {
	case *fxevent.OnStartExecuting, *fxevent.OnStopExecuting:
		var functionName, callerName string
		switch ev := event.(type) {
		case *fxevent.OnStartExecuting:
			functionName = ev.FunctionName
			callerName = ev.CallerName
		case *fxevent.OnStopExecuting:
			functionName = ev.FunctionName
			callerName = ev.CallerName
		}
		log.Debug(h.prefix+" hook executing",
			"callee", functionName,
			"caller", callerName,
		)
	case *fxevent.OnStartExecuted, *fxevent.OnStopExecuted:
		var functionName, callerName string
		var err error
		var runtime time.Duration
		switch ev := event.(type) {
		case *fxevent.OnStartExecuted:
			functionName = ev.FunctionName
			callerName = ev.CallerName
			err = ev.Err
			runtime = ev.Runtime
		case *fxevent.OnStopExecuted:
			functionName = ev.FunctionName
			callerName = ev.CallerName
			err = ev.Err
			runtime = ev.Runtime
		}
		if err != nil {
			log.Error(h.prefix+" hook failed",
				"callee", functionName,
				"caller", callerName,
				"error", err,
			)
		} else {
			log.Debug(h.prefix+" hook executed",
				"callee", functionName,
				"caller", callerName,
				"runtime", runtime,
			)
		}
	}
}

// simpleEventHandler handles simple success/error events.
type simpleEventHandler struct {
	successMsg string
	errorMsg   string
}

// Handle implements the eventHandler interface for simple events.
func (h *simpleEventHandler) Handle(log Interface, event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.Supplied:
		if e.Err != nil {
			log.Error(h.errorMsg,
				"type", e.TypeName,
				"error", e.Err,
			)
		} else {
			log.Debug(h.successMsg,
				"type", e.TypeName,
			)
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			log.Error(h.errorMsg,
				"function", e.FunctionName,
				"error", e.Err,
			)
		} else {
			log.Debug(h.successMsg,
				"function", e.FunctionName,
			)
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			log.Error(h.errorMsg,
				"function", e.ConstructorName,
				"error", e.Err,
			)
		} else {
			log.Debug(h.successMsg,
				"function", e.ConstructorName,
			)
		}
	}
}

// signalEventHandler handles signal-related events.
type signalEventHandler struct{}

// Handle implements the eventHandler interface for signal events.
func (h *signalEventHandler) Handle(log Interface, event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.Stopping:
		log.Info("Received signal",
			"signal", e.Signal,
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			log.Error("Stop failed",
				"error", e.Err,
			)
		}
	case *fxevent.RolledBack:
		log.Error("Start failed, rolling back",
			"error", e.Err,
		)
	case *fxevent.Started:
		if e.Err != nil {
			log.Error("Start failed",
				"error", e.Err,
			)
		} else {
			log.Info("Started")
		}
	}
}

// eventHandlers maps event types to their handlers.
var eventHandlers = map[reflect.Type]eventHandler{
	reflect.TypeOf(&fxevent.Provided{}): &multiTypeEventHandler{
		successMsg: "Provided",
		errorMsg:   "Error encountered while applying options",
	},
	reflect.TypeOf(&fxevent.Replaced{}): &multiTypeEventHandler{
		successMsg: "Replaced",
		errorMsg:   "Error encountered while replacing",
	},
	reflect.TypeOf(&fxevent.Decorated{}): &multiTypeEventHandler{
		successMsg: "Decorated",
		errorMsg:   "Error encountered while applying options",
	},
	reflect.TypeOf(&fxevent.OnStartExecuting{}): &hookEventHandler{prefix: "OnStart"},
	reflect.TypeOf(&fxevent.OnStartExecuted{}):  &hookEventHandler{prefix: "OnStart"},
	reflect.TypeOf(&fxevent.OnStopExecuting{}):  &hookEventHandler{prefix: "OnStop"},
	reflect.TypeOf(&fxevent.OnStopExecuted{}):   &hookEventHandler{prefix: "OnStop"},
	reflect.TypeOf(&fxevent.Supplied{}): &simpleEventHandler{
		successMsg: "Supplied",
		errorMsg:   "Error encountered while applying options",
	},
	reflect.TypeOf(&fxevent.Invoked{}): &simpleEventHandler{
		successMsg: "Invoked",
		errorMsg:   "Invoke failed",
	},
	reflect.TypeOf(&fxevent.LoggerInitialized{}): &simpleEventHandler{
		successMsg: "Initialized custom fxevent.Logger",
		errorMsg:   "Custom logger initialization failed",
	},
	reflect.TypeOf(&fxevent.Stopping{}):   &signalEventHandler{},
	reflect.TypeOf(&fxevent.Stopped{}):    &signalEventHandler{},
	reflect.TypeOf(&fxevent.RolledBack{}): &signalEventHandler{},
	reflect.TypeOf(&fxevent.Started{}):    &signalEventHandler{},
}

// LogEvent logs an fx event.
func (l *fxLogger) LogEvent(event fxevent.Event) {
	handler, ok := eventHandlers[reflect.TypeOf(event)]
	if !ok {
		l.log.Error("Unknown event type",
			"type", reflect.TypeOf(event),
		)
		return
	}
	handler.Handle(l.log, event)
}
