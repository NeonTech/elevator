# elevator

Ascending logs at descending costs.

`elevator` is a [`slog.Handler`](https://pkg.go.dev/log/slog#Handler) that
enables applications to use as many debug/info/warn logs as desired to trivially
debug errors without "noise" from those logs during normal operations. Avoid
spending money on ingesting, and time parsing, debug/info/warn logs when you do
not need them while having them available when you do.

## Install

```sh
go get github.com/NeonTech/elevator
```

## Usage

```go
logger := slog.New(elevator.NewHandler())
```

Full documentation is available on
[pkg.go.dev](https://pkg.go.dev/github.com/NeonTech/elevator).
