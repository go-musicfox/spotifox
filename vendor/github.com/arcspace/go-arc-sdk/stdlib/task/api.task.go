package task

import (
	"context"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/log"
)

// NilContext is used to start Contexts with no parent Context.
var NilContext = Context((*ctx)(nil))

// Start starts the given context as its own task root.
func Start(task *Task) (Context, error) {
	return NilContext.StartChild(task)
}

// Go starts a new root Context with the given label and function.
func Go(parent Context, label string, fn func(ctx Context)) (Context, error) {
	return parent.StartChild(&Task{
		Label: label,
		OnRun: fn,
	})
}

// Task is an optional set of callbacks for a Context
type Task struct {

	// If > 0, CloseWhenIdle() is automatically called after the last remaining child is closed or after OnRun() completes (if set) -- whichever occurs later.
	// Note how this will not enter into effect unless OnRun is given or a child is started.
	IdleClose time.Duration

	TaskRef        any                     // Offered to you for open-ended use.
	Label          string                  // Label is a log label and debugging
	OnStart        func(ctx Context) error // Blocking fn called in StartChild(). If err, ctx.Close() is called and Go() returns the err and OnRun is never called.
	OnRun          func(ctx Context)       // Async work body. If non-nil, ctx.Close() will be automatically called after OnRun() completes
	OnClosing      func()                  // Called immediately after Close() is first called while self & children are still closing
	OnChildClosing func(child Context)     // Called immediately after the child's OnClosing() is called
	OnClosed       func()                  // Called after Close() and all children have completed Close() (but immediately before Done() is released)
}

type Context interface {
	log.Logger

	// A task.Context is an extension of context.Context.
	context.Context

	// Returns Task.Ref passed into StartChild()
	TaskRef() interface{}

	// The context's public label
	Label() string

	// A guaranteed unique ID assigned after Start() is called.
	ContextID() int64

	// Creates a new child Context with for given Task.
	// If OnStart() returns an error error is encountered, then child.Close() is immediately called and the error is returned.
	StartChild(task *Task) (Context, error)

	// Convenience function for StartChild() and is equivalent to:
	//
	//      parent.StartChild(label, &Task{
	//  		IdleClose: time.Nanosecond,
	// 	        OnRun: fn,
	//      })
	Go(label string, fn func(ctx Context)) (Context, error)

	// Appends all currently open/active child Contexts to the given slice and returns the given slice.
	// Naturally, the returned items are back-ward looking as any could close at any time.
	// Context implementations wishing to remain lightweight may opt to not retain a list of children (and just return the given slice as-is).
	GetChildren(in []Context) []Context

	// Async call that initiates task shutdown and causes all children's Close() to be called.
	// Close can be called multiple times but calls after the first are in effect ignored.
	// First, children get Close() in breath-first order.
	// After all children are done closing, OnClosing(), then OnClosed() are executed.
	Close() error

	// Inserts a pending Close() on this Context once it is idle after the given delay.
	// Subsequent calls will update the delay but the previously pending delay must run out first.
	// If at the end of the period Task.OnRun() is complete, there are no children, PreventIdleClose() is not in effect, then Close() is called.
	CloseWhenIdle(delay time.Duration)

	// Ensures that that this Context will not automatically idle-close until the given delay has passed.
	// If previous PreventIdleClose calls were made, the more limiting delay is retained.
	//
	// Returns false if this Context has already been closed.
	PreventIdleClose(delay time.Duration) bool

	// Signals when Close() has been called.
	// First, Children get Close(),  then OnClosing, then OnClosed are executing
	Closing() <-chan struct{}

	// Signals when Close() has fully executed, no children remain, and OnClosed() has been completed.
	Done() <-chan struct{}
}
