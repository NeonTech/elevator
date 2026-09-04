package elevator

import "log/slog"

// HandlerOption configures a [Handler].
type HandlerOption func(handler *Handler)

// WithDelegate configures the [slog.Handler] that handles records when the
// buffer is flushed.
//
// If delegate is nil, then a default delegate is configured.
func WithDelegate(delegate slog.Handler) HandlerOption {
	if delegate == nil {
		delegate = defaultDelegate
	}

	return func(handler *Handler) {
		handler.delegate = delegate
	}
}

// WithBuffer configures the buffer to discard the oldest records when it
// reaches capacity.
//
// If capacity is less than 1, then a capacity of 1 is configured.
func WithBuffer(capacity int) HandlerOption {
	if capacity < 1 {
		capacity = 1
	}

	return func(handler *Handler) {
		state := handler.state
		state.buffer = make([]bufferedRecord, capacity)
		state.unlimitedBuffer = false
	}
}

// WithUnlimitedBuffer configures the buffer to grow when it reaches capacity.
//
// If initial capacity is less than 1, then an initial capacity of 1 is
// configured.
func WithUnlimitedBuffer(initialCapacity int) HandlerOption {
	if initialCapacity < 1 {
		initialCapacity = 1
	}

	return func(handler *Handler) {
		state := handler.state
		state.buffer = make([]bufferedRecord, 0, initialCapacity)
		state.unlimitedBuffer = true
	}
}

// WithFlushLevel configures the minimum record level to trigger a flush of the
// buffer.
func WithFlushLevel(flushLevel slog.Level) HandlerOption {
	return func(handler *Handler) {
		state := handler.state
		state.flushLevel = flushLevel
	}
}
