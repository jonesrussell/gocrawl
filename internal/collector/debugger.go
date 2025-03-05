package collector

import (
	"github.com/gocolly/colly/v2/debug"
)

// DebuggerInterface is an interface for the debugger
type DebuggerInterface interface {
	Init() error
	Event(e *debug.Event)
}

// Ensure debugger implementations satisfy the interface
var _ DebuggerInterface = (debug.Debugger)(nil)
