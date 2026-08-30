package elevator

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

type bufferedRecord struct {
	ctx     context.Context
	record  slog.Record
	handler slog.Handler
}

type bufferState struct {
	buffer          []bufferedRecord
	unlimitedBuffer bool
	head            int
	size            int
	locker          sync.Mutex
	flushLevel      slog.Level
}

func (state *bufferState) push(record bufferedRecord) {
	if state.unlimitedBuffer {
		state.buffer = append(state.buffer, record)
		state.size++
		return
	}

	capacity := len(state.buffer)
	head := state.head
	size := state.size

	if size < capacity {
		state.buffer[(head+size)%capacity] = record
		state.size++
		return
	}

	state.buffer[head] = record
	state.head = (head + 1) % capacity
}

func (state *bufferState) flush() error {
	var errs []error

	if state.unlimitedBuffer {
		for index := range state.size {
			record := state.buffer[index]
			// Drop references to allow GC
			state.buffer[index] = bufferedRecord{}

			handler := record.handler
			err := handler.Handle(record.ctx, record.record)
			if err != nil {
				errs = append(errs, err)
			}
		}

		state.buffer = state.buffer[:0]
		state.size = 0

		return errors.Join(errs...)
	}

	capacity := len(state.buffer)
	head := state.head
	size := state.size

	for index := range size {
		index = (head + index) % capacity
		record := state.buffer[index]
		// Drop references to allow GC
		state.buffer[index] = bufferedRecord{}

		handler := record.handler
		err := handler.Handle(record.ctx, record.record)
		if err != nil {
			errs = append(errs, err)
		}
	}

	state.head = 0
	state.size = 0

	return errors.Join(errs...)
}
