package logger

import (
	"reflect"
	"strings"
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

// logSuccessOrError is a helper method for logging success or error events.
func (l *fxLogger) logSuccessOrError(successMsg, errorMsg string, err error, fields ...any) {
	if err != nil {
		l.log.Error(errorMsg, append(fields, "error", err)...)
	} else {
		l.log.Debug(successMsg, fields...)
	}
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
	switch event.(type) {
	case *fxevent.Provided, *fxevent.Replaced, *fxevent.Decorated:
		var outputTypes []string
		var constructorName, decoratorName string
		var err error

		switch ev := event.(type) {
		case *fxevent.Provided:
			outputTypes = ev.OutputTypeNames
			constructorName = ev.ConstructorName
			err = ev.Err
		case *fxevent.Replaced:
			outputTypes = ev.OutputTypeNames
			err = ev.Err
		case *fxevent.Decorated:
			outputTypes = ev.OutputTypeNames
			decoratorName = ev.DecoratorName
			err = ev.Err
		}

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

		for _, rtype := range outputTypes {
			if !importantTypes[rtype] {
				continue
			}

			fields := []any{"type", rtype}
			if constructorName != "" {
				msg := "Initialized " + cleanConstructorName(constructorName)
				log.Debug(msg, fields...)
			} else if decoratorName != "" {
				fields = append(fields, "decorator", decoratorName)
				log.Debug(h.successMsg, fields...)
			} else {
				log.Debug(h.successMsg, fields...)
			}
		}

		if err != nil {
			log.Error(h.errorMsg, "error", err)
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
			"callee", cleanConstructorName(functionName),
			"caller", cleanConstructorName(callerName),
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
		(&fxLogger{log: log}).logSuccessOrError(
			h.prefix+" hook executed",
			h.prefix+" hook failed",
			err,
			"callee", cleanConstructorName(functionName),
			"caller", cleanConstructorName(callerName),
			"runtime", runtime,
		)
	}
}

// simpleEventHandler handles simple success/error events.
type simpleEventHandler struct {
	successMsg string
	errorMsg   string
}

// Handle implements the eventHandler interface for simple events.
func (h *simpleEventHandler) Handle(log Interface, event fxevent.Event) {
	switch event.(type) {
	case *fxevent.Supplied, *fxevent.Invoked, *fxevent.LoggerInitialized, *fxevent.Invoking, *fxevent.Run:
		var err error
		var fields []any

		switch ev := event.(type) {
		case *fxevent.Supplied:
			err = ev.Err
			fields = []any{"type", ev.TypeName}
		case *fxevent.Invoked:
			err = ev.Err
			fields = []any{"function", cleanConstructorName(ev.FunctionName)}
		case *fxevent.LoggerInitialized:
			err = ev.Err
			fields = []any{"function", cleanConstructorName(ev.ConstructorName)}
		case *fxevent.Invoking:
			fields = []any{"function", cleanConstructorName(ev.FunctionName)}
		case *fxevent.Run:
			fields = []any{"function", cleanConstructorName(ev.Name)}
		}

		(&fxLogger{log: log}).logSuccessOrError(h.successMsg, h.errorMsg, err, fields...)
	}
}

// signalEventHandler handles signal-related events.
type signalEventHandler struct{}

// Handle implements the eventHandler interface for signal events.
func (h *signalEventHandler) Handle(log Interface, event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.Stopping:
		log.Info("Received signal", "signal", e.Signal)
	case *fxevent.Stopped, *fxevent.RolledBack, *fxevent.Started:
		var err error
		var msg string
		switch ev := event.(type) {
		case *fxevent.Stopped:
			err = ev.Err
			msg = "Stop failed"
		case *fxevent.RolledBack:
			err = ev.Err
			msg = "Start failed, rolling back"
		case *fxevent.Started:
			err = ev.Err
			msg = "Start failed"
		}

		if err != nil {
			log.Error(msg, "error", err)
		} else if _, ok := event.(*fxevent.Started); ok {
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
	reflect.TypeOf(&fxevent.Invoking{}): &simpleEventHandler{
		successMsg: "Invoking",
		errorMsg:   "Invoke failed",
	},
	reflect.TypeOf(&fxevent.Run{}): &simpleEventHandler{
		successMsg: "Running",
		errorMsg:   "Run failed",
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
