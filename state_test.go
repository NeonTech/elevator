package elevator

import (
	"errors"
	"log/slog"
	"slices"
	"testing"
)

func TestBufferStatePushUnlimited(t *testing.T) {
	state := &bufferState{
		buffer:          make([]bufferedRecord, 0, 2),
		unlimitedBuffer: true,
	}

	state.push(bufferedRecord{record: newRecord(slog.LevelInfo, "a")})
	state.push(bufferedRecord{record: newRecord(slog.LevelInfo, "b")})
	state.push(bufferedRecord{record: newRecord(slog.LevelInfo, "c")})

	if state.size != 3 {
		t.Fatalf("buffer size got %d, want 3", state.size)
	}

	gotMessages := state.messages()
	wantMessages := []string{"a", "b", "c"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("buffer got %v, want %v", gotMessages, wantMessages)
	}
}

func TestBufferStatePush(t *testing.T) {
	state := &bufferState{
		buffer:          make([]bufferedRecord, 2),
		unlimitedBuffer: false,
	}

	state.push(bufferedRecord{record: newRecord(slog.LevelInfo, "a")})
	state.push(bufferedRecord{record: newRecord(slog.LevelInfo, "b")})

	if state.size != 2 {
		t.Fatalf("buffer size got %d, want 2", state.size)
	}
	if state.head != 0 {
		t.Fatalf("buffer head got %d, want 0", state.head)
	}

	gotMessages := state.messages()
	wantMessages := []string{"a", "b"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("buffer got %v, want %v", gotMessages, wantMessages)
	}

	// Pushing again must overwrite the oldest record ("a") and advance head
	state.push(bufferedRecord{record: newRecord(slog.LevelInfo, "c")})

	if state.size != 2 {
		t.Fatalf("buffer size got %d, want 2", state.size)
	}
	if state.head != 1 {
		t.Fatalf("buffer head got %d, want 1", state.head)
	}

	gotMessages = state.messages()
	wantMessages = []string{"b", "c"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("buffer got %v, want %v", gotMessages, wantMessages)
	}
}

func TestBufferStateFlushUnlimitedNoErrors(t *testing.T) {
	ok := &mockHandler{}
	state := &bufferState{
		buffer:          make([]bufferedRecord, 0, 2),
		unlimitedBuffer: true,
	}

	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "a"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "b"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "c"),
			handler: ok,
		},
	)

	gotErr := state.flush()
	if gotErr != nil {
		t.Fatalf("flush() got %v, want nil", gotErr)
	}

	gotMessages := ok.messages()
	wantMessages := []string{"a", "b", "c"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("ok delegate got %v, want %v", gotMessages, wantMessages)
	}

	if state.size != 0 {
		t.Fatalf("buffer size got %d, want 0", state.size)
	}

	gotLength := len(state.buffer)
	if gotLength != 0 {
		t.Fatalf("buffer length got %d, want 0", gotLength)
	}
}

func TestBufferStateFlushUnlimitedWithErrors(t *testing.T) {
	ok := &mockHandler{}
	failingErr := errors.New("boom")
	failing := &mockHandler{err: failingErr}

	state := &bufferState{
		buffer:          make([]bufferedRecord, 0, 2),
		unlimitedBuffer: true,
	}

	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "a"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "b"),
			handler: failing,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "c"),
			handler: failing,
		},
	)

	gotErr := state.flush()
	if !errors.Is(gotErr, failingErr) {
		t.Fatalf("flush() got %v, want %v", gotErr, failingErr)
	}

	gotMessages := ok.messages()
	wantMessages := []string{"a"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("ok delegate got %v, want %v", gotMessages, wantMessages)
	}

	gotMessages = failing.messages()
	wantMessages = []string{"b", "c"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("failing delegate got %v, want %v", gotMessages, wantMessages)
	}

	if state.size != 0 {
		t.Fatalf("buffer size got %d, want 0", state.size)
	}

	gotLength := len(state.buffer)
	if gotLength != 0 {
		t.Fatalf("buffer length got %d, want 0", gotLength)
	}
}

func TestBufferStateFlushNoErrors(t *testing.T) {
	ok := &mockHandler{}
	state := &bufferState{
		buffer:          make([]bufferedRecord, 2),
		unlimitedBuffer: false,
	}

	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "a"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "b"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "c"),
			handler: ok,
		},
	)

	gotErr := state.flush()
	if gotErr != nil {
		t.Fatalf("flush() got %v, want nil", gotErr)
	}

	gotMessages := ok.messages()
	wantMessages := []string{"b", "c"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("ok delegate got %v, want %v", gotMessages, wantMessages)
	}

	if state.size != 0 {
		t.Fatalf("buffer size got %d, want 0", state.size)
	}
	if state.head != 0 {
		t.Fatalf("buffer head got %d, want 0", state.head)
	}
}

func TestBufferStateFlushWithErrors(t *testing.T) {
	ok := &mockHandler{}
	failingErr := errors.New("boom")
	failing := &mockHandler{err: failingErr}

	state := &bufferState{
		buffer:          make([]bufferedRecord, 2),
		unlimitedBuffer: false,
	}

	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "a"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "b"),
			handler: ok,
		},
	)
	state.push(
		bufferedRecord{
			record:  newRecord(slog.LevelInfo, "c"),
			handler: failing,
		},
	)

	gotErr := state.flush()
	if !errors.Is(gotErr, failingErr) {
		t.Fatalf("flush() got %v, want %v", gotErr, failingErr)
	}

	gotMessages := ok.messages()
	wantMessages := []string{"b"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("ok delegate got %v, want %v", gotMessages, wantMessages)
	}

	gotMessages = failing.messages()
	wantMessages = []string{"c"}
	if !slices.Equal(gotMessages, wantMessages) {
		t.Fatalf("failing delegate got %v, want %v", gotMessages, wantMessages)
	}

	if state.size != 0 {
		t.Fatalf("buffer size got %d, want 0", state.size)
	}
	if state.head != 0 {
		t.Fatalf("buffer head got %d, want 0", state.head)
	}
}
