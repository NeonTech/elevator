package elevator

import (
	"context"
	"log/slog"
	"time"
)

type mockHandler struct {
	records []slog.Record
	err     error
	attrs   []slog.Attr
	groups  []string
}

func (*mockHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (handler *mockHandler) Handle(ctx context.Context, record slog.Record) error {
	handler.records = append(handler.records, record)
	return handler.err
}

func (handler *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	attrsCopy := append([]slog.Attr{}, handler.attrs...)
	attrsCopy = append(attrsCopy, attrs...)
	return &mockHandler{
		records: handler.records,
		err:     handler.err,
		attrs:   attrsCopy,
		groups:  handler.groups,
	}
}

func (handler *mockHandler) WithGroup(name string) slog.Handler {
	groupsCopy := append([]string{}, handler.groups...)
	groupsCopy = append(groupsCopy, name)
	return &mockHandler{
		records: handler.records,
		err:     handler.err,
		attrs:   handler.attrs,
		groups:  groupsCopy,
	}
}

func (handler *mockHandler) messages() []string {
	messages := make([]string, len(handler.records))
	for index, record := range handler.records {
		messages[index] = record.Message
	}
	return messages
}

func (state *bufferState) messages() []string {
	if state.unlimitedBuffer {
		messages := make([]string, len(state.buffer))
		for index, bufferedRecord := range state.buffer {
			record := bufferedRecord.record
			messages[index] = record.Message
		}
		return messages
	}

	messages := make([]string, state.size)
	for index := range state.size {
		bufferIndex := (state.head + index) % len(state.buffer)
		bufferedRecord := state.buffer[bufferIndex]
		record := bufferedRecord.record
		messages[index] = record.Message
	}
	return messages
}

func newRecord(level slog.Level, message string) slog.Record {
	return slog.NewRecord(time.Now(), level, message, 0)
}
