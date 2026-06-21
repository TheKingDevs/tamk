package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

type Level slog.Level

const (
	LevelDebug   Level = Level(slog.LevelDebug)
	LevelInfo    Level = Level(slog.LevelInfo)
	LevelWarn    Level = Level(slog.LevelWarn)
	LevelError   Level = Level(slog.LevelError)
	LevelStep    Level = Level(slog.Level(2))
	LevelSuccess Level = Level(slog.Level(4))
)

var programLevel = new(slog.LevelVar)

func Init(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	programLevel.Set(level)

	handler := &consoleHandler{
		level: programLevel,
		w:     os.Stderr,
	}
	slog.SetDefault(slog.New(handler))
}

func SetOutput(w io.Writer) {
	handler := &consoleHandler{
		level: programLevel,
		w:     w,
	}
	slog.SetDefault(slog.New(handler))
}

func SetLevel(level string) {
	switch level {
	case "DEBUG":
		programLevel.Set(slog.LevelDebug)
	case "INFO":
		programLevel.Set(slog.LevelInfo)
	case "WARN":
		programLevel.Set(slog.LevelWarn)
	case "ERROR":
		programLevel.Set(slog.LevelError)
	}
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Step(msg string, args ...any) {
	slog.Log(context.Background(), slog.Level(2), msg, args...)
}

func Success(msg string, args ...any) {
	slog.Log(context.Background(), slog.Level(4), msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

const (
	AnsiReset  = "\033[0m"
	AnsiBold   = "\033[1m"
	AnsiDim    = "\033[2m"
	AnsiCyan   = "\033[36m"
	AnsiGreen  = "\033[32m"
	AnsiYellow = "\033[33m"
	AnsiRed    = "\033[31m"
	AnsiGray   = "\033[90m"
	AnsiBlue   = "\033[34m"
	AnsiPurple = "\033[35m"
)

type consoleHandler struct {
	level *slog.LevelVar
	w     io.Writer
	mu    sync.Mutex
}

func (h *consoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level
	var label string
	var color string

	switch {
	case level == slog.Level(4):
		label = " OK "
		color = AnsiGreen
	case level == slog.Level(2):
		label = "STEP"
		color = AnsiCyan
	case level >= slog.LevelError:
		label = "ERROR"
		color = AnsiRed
	case level >= slog.LevelWarn:
		label = "WARN"
		color = AnsiYellow
	case level >= slog.LevelInfo:
		label = "INFO"
		color = AnsiBlue
	default:
		label = "DEBUG"
		color = AnsiGray
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Fprintf(h.w, "%s[%s]%s %s", color, label, AnsiReset, r.Message)

	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() != slog.KindGroup {
			fmt.Fprintf(h.w, " %s=%v", a.Key, a.Value.Any())
		}
		return true
	})

	fmt.Fprintln(h.w)
	return nil
}

func (h *consoleHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *consoleHandler) WithGroup(_ string) slog.Handler {
	return h
}
