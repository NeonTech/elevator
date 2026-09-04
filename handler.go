package elevator

import (
	"context"
	"log/slog"
	"os"
)

const (
	defaultInitialCapacity = 16
	defaultFlushLevel      = slog.LevelError
)

var (
	defaultDelegate = slog.NewJSONHandler(os.Stdout, nil)
)

// Handler is a [slog.Handler] that handles all records at all levels by
// buffering them. The buffer is flushed when a record is handled that is
// greater than or equal to the configured flush level. Flushing the buffer
// causes all records in the buffer to be handled by the configured delegate
// [slog.Handler] (at the time when the record was buffered).
//
// Use [NewHandler] to create a new Handler; do not construct directly.
//
// This implementation is most useful when used in conjunction with a scoped
// "id" attribute, for example:
//
//	logger := elevator.NewLogger()
//	// id is created in a "scope", for example, an HTTP request.
//	logger = logger.With("id", id)
//
// Debug, Info, and Warn logs will not be outputted until Error is called, in
// which case, all outputted logs can be correlated by "id". This enables
// applications to use as many debug/info/warn logs as desired to trivially
// debug errors without "noise" from those logs during normal operations.
type Handler struct {
	delegate slog.Handler
	state    *bufferState
}

// Enabled reports records are handled at all levels by always returning true.
func (*Handler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle buffers record. If the record's level is greater than or equal to the
// receiver's flush level, then the buffer is flushed.
func (handler *Handler) Handle(ctx context.Context, record slog.Record) error {
	state := handler.state

	state.locker.Lock()
	defer state.locker.Unlock()

	state.push(
		bufferedRecord{
			ctx:     ctx,
			record:  record.Clone(),
			handler: handler.delegate,
		},
	)

	if record.Level < state.flushLevel {
		return nil
	}

	return state.flush()
}

// WithAttrs returns a new [Handler] with a delegate set to the receiver's
// delegate returning [slog.Handler.WithAttrs]
// EXCEPT when attrs is empty, in which case, the receiver is returned.
func (handler *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return handler
	}

	delegate := handler.delegate
	return &Handler{
		delegate: delegate.WithAttrs(attrs),
		state:    handler.state,
	}
}

// WithGroup returns a new [Handler] with a delegate set to the receiver's
// delegate returning [slog.Handler.WithGroup]
// EXCEPT when name is empty, in which case, the receiver is returned.
func (handler *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return handler
	}

	delegate := handler.delegate
	return &Handler{
		delegate: delegate.WithGroup(name),
		state:    handler.state,
	}
}

// NewHandler returns a new [Handler] configured with options available via
// With* functions; for example:
//
//	elevator.NewHandler(
//	  elevator.WithBuffer(16),
//	  elevator.WithFlushLevel(slog.LevelWarn),
//	)
//
// Unless explicitly configured by options, the returned Handler will have the
// following default behavior:
//   - Delegate is a [slog.JSONHandler] writing to [os.Stdout]
//   - Buffer is unlimited with an initial capacity of 16
//   - Flush level is set to [slog.LevelError]
func NewHandler(options ...HandlerOption) *Handler {
	state := &bufferState{
		flushLevel: defaultFlushLevel,
	}

	handler := &Handler{
		delegate: defaultDelegate,
		state:    state,
	}

	for _, option := range options {
		option(handler)
	}

	if state.buffer == nil {
		option := WithUnlimitedBuffer(defaultInitialCapacity)
		option(handler)
	}

	return handler
}

// NewLogger returns a new [slog.Logger] whose handler is [NewHandler]
// configured with options; it is equivalent to:
//
//	slog.New(NewHandler(options...))
func NewLogger(options ...HandlerOption) *slog.Logger {
	return slog.New(NewHandler(options...))
}
