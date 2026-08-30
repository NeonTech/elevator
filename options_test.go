package elevator

import (
	"log/slog"
	"testing"
)

func TestWithDelegateNil(t *testing.T) {
	handler := NewHandler(WithDelegate(nil))
	if handler.delegate != defaultDelegate {
		t.Fatalf("delegate got %v, want %v", handler.delegate, defaultDelegate)
	}
}

func TestWithDelegateNonNil(t *testing.T) {
	delegate := &mockHandler{}
	handler := NewHandler(WithDelegate(delegate))
	if handler.delegate != delegate {
		t.Fatalf("delegate got %v, want %v", handler.delegate, delegate)
	}
}

func TestWithBuffer(t *testing.T) {
	tests := map[string]struct {
		capacity     int
		wantCapacity int
	}{
		"clamps zero":     {0, 1},
		"clamps negative": {-3, 1},
		"keeps explicit":  {3, 3},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(WithBuffer(test.capacity))
			state := handler.state

			if state.unlimitedBuffer {
				t.Fatalf("unlimitedBuffer got true, want false")
			}

			gotCapacity := len(state.buffer)
			wantCapacity := test.wantCapacity
			if gotCapacity != wantCapacity {
				t.Fatalf("buffer length got %d, want %d", gotCapacity, wantCapacity)
			}
		})
	}
}

func TestWithUnlimitedBuffer(t *testing.T) {
	tests := map[string]struct {
		initialCapacity     int
		wantInitialCapacity int
	}{
		"clamps zero":     {0, 1},
		"clamps negative": {-3, 1},
		"keeps explicit":  {3, 3},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(WithUnlimitedBuffer(test.initialCapacity))
			state := handler.state

			if !state.unlimitedBuffer {
				t.Fatalf("unlimitedBuffer got false, want true")
			}

			gotLength := len(state.buffer)
			if gotLength != 0 {
				t.Fatalf("buffer length got %d, want 0", gotLength)
			}

			gotCapacity := cap(state.buffer)
			wantCapacity := test.wantInitialCapacity
			if gotCapacity != wantCapacity {
				t.Fatalf("buffer capacity got %d, want %d", gotCapacity, wantCapacity)
			}
		})
	}
}

func TestWithFlushLevel(t *testing.T) {
	handler := NewHandler(WithFlushLevel(slog.LevelWarn))
	state := handler.state

	if state.flushLevel != slog.LevelWarn {
		t.Fatalf("flushLevel got %v, want %v", state.flushLevel, slog.LevelWarn)
	}
}
