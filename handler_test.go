package elevator

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"
)

func equalAttr(attr1 slog.Attr, attr2 slog.Attr) bool {
	return attr1.Equal(attr2)
}

func TestHandlerEnabled(t *testing.T) {
	handler := NewHandler()

	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		got := handler.Enabled(context.Background(), level)
		if !got {
			t.Fatalf("Enabled(%v) got %v, want true", level, got)
		}
	}
}

func TestHandlerHandleBelowFlushLevel(t *testing.T) {
	delegate := &mockHandler{}
	handler := NewHandler(WithDelegate(delegate))

	gotErr := handler.Handle(context.Background(), newRecord(slog.LevelInfo, "buffered"))
	if gotErr != nil {
		t.Fatalf("Handle() got %v, want nil", gotErr)
	}

	gotMessages := delegate.messages()
	wantMessages := []string{}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("delegate got %v, want %v", gotMessages, wantMessages)
	}
}

func TestHandlerHandleAtFlushLevel(t *testing.T) {
	delegate := &mockHandler{}
	handler := NewHandler(WithDelegate(delegate))

	gotErr := handler.Handle(context.Background(), newRecord(slog.LevelInfo, "buffered"))
	if gotErr != nil {
		t.Fatalf("Handle() got %v, want nil", gotErr)
	}

	gotErr = handler.Handle(context.Background(), newRecord(slog.LevelError, "trigger"))
	if gotErr != nil {
		t.Fatalf("Handle() got %v, want nil", gotErr)
	}

	gotMessages := delegate.messages()
	wantMessages := []string{"buffered", "trigger"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("delegate got %v, want %v", gotMessages, wantMessages)
	}
}

func TestHandlerHandleFlushError(t *testing.T) {
	failingErr := errors.New("boom")
	delegate := &mockHandler{err: failingErr}
	handler := NewHandler(WithDelegate(delegate))

	gotErr := handler.Handle(context.Background(), newRecord(slog.LevelError, "trigger"))
	if !errors.Is(gotErr, failingErr) {
		t.Fatalf("Handle() got %v, want %v", gotErr, failingErr)
	}
}

func TestHandlerWithAttrsEmpty(t *testing.T) {
	handler := NewHandler()

	gotHandler := handler.WithAttrs(nil)
	if gotHandler != handler {
		t.Fatalf("WithAttrs(nil) got %v, want %v", gotHandler, handler)
	}

	gotHandler = handler.WithAttrs([]slog.Attr{})
	if gotHandler != handler {
		t.Fatalf("WithAttrs([]) got %v, want %v", gotHandler, handler)
	}
}

func TestHandlerWithAttrsNonEmpty(t *testing.T) {
	delegate := &mockHandler{}
	handler := NewHandler(WithDelegate(delegate))

	attr := slog.String("id", "1234567890")

	derived := handler.WithAttrs([]slog.Attr{attr})
	if derived == handler {
		t.Fatalf("WithAttrs() got %v, do NOT want %v", derived, handler)
	}

	derivedHandler, ok := derived.(*Handler)
	if !ok {
		t.Fatalf("derived handler is %T, want *Handler", derived)
	}

	rawDerivedDelegate := derivedHandler.delegate
	derivedDelegate, ok := rawDerivedDelegate.(*mockHandler)
	if !ok {
		t.Fatalf("derived delegate is %T, want *mockHandler", rawDerivedDelegate)
	}

	gotAttrs := derivedDelegate.attrs
	wantAttrs := []slog.Attr{attr}
	if !slices.EqualFunc(gotAttrs, wantAttrs, equalAttr) {
		t.Fatalf("derived delegate attrs got %v, want %v", gotAttrs, wantAttrs)
	}
}

func TestHandlerWithGroupEmpty(t *testing.T) {
	handler := NewHandler()

	gotHandler := handler.WithGroup("")
	if gotHandler != handler {
		t.Fatalf("WithGroup(\"\") got %v, want %v", gotHandler, handler)
	}
}

func TestHandlerWithGroupNonEmpty(t *testing.T) {
	delegate := &mockHandler{}
	handler := NewHandler(WithDelegate(delegate))

	derived := handler.WithGroup("group")
	if derived == handler {
		t.Fatalf("WithGroup() got %v, do NOT want %v", derived, handler)
	}

	derivedHandler, ok := derived.(*Handler)
	if !ok {
		t.Fatalf("derived handler is %T, want *Handler", derived)
	}

	rawDerivedDelegate := derivedHandler.delegate
	derivedDelegate, ok := rawDerivedDelegate.(*mockHandler)
	if !ok {
		t.Fatalf("derived delegate is %T, want *mockHandler", rawDerivedDelegate)
	}

	gotGroups := derivedDelegate.groups
	wantGroups := []string{"group"}
	if !slices.Equal(gotGroups, wantGroups) {
		t.Fatalf("derived delegate groups got %v, want %v", gotGroups, wantGroups)
	}
}

func TestNewHandlerDefaults(t *testing.T) {
	handler := NewHandler()
	state := handler.state

	if handler.delegate != defaultDelegate {
		t.Fatalf("delegate got %v, want %v", handler.delegate, defaultDelegate)
	}

	if !state.unlimitedBuffer {
		t.Fatalf("unlimitedBuffer got false, want true")
	}

	gotCapacity := cap(state.buffer)
	if gotCapacity != defaultInitialCapacity {
		t.Fatalf("buffer capacity got %d, want %d", gotCapacity, defaultInitialCapacity)
	}

	if state.flushLevel != defaultFlushLevel {
		t.Fatalf("flushLevel got %v, want %v", state.flushLevel, defaultFlushLevel)
	}
}
