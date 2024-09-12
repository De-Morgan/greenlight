package logger

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

const (
	LevelTrace Level = iota
	LevelInfo
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelOff:
		return "OFF"
	default:
		return ""
	}
}

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// Return a new Logger instance which writes log entries at or above a minimum severity
// level to a specific output destination.

func New(out io.Writer, min Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: min,
	}
}

func (l *Logger) PrintTrace(message string, properties map[string]any) {
	l.print(LevelTrace, message, properties)

}
func (l *Logger) PrintInfo(message string, properties map[string]any) {
	l.print(LevelInfo, message, properties)

}

func (l *Logger) PrintError(message string, properties map[string]any) {
	l.print(LevelError, message, properties)

}

func (l *Logger) PrintFatal(message string, properties map[string]any) {
	l.print(LevelFatal, message, properties)
	os.Exit(1)
}

func (l *Logger) print(level Level, message string, properties map[string]any) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string         `json:"level"`
		Time       string         `json:"time"`
		Message    string         `json:"message"`
		Properties map[string]any `json:"properties"`
		Trace      string         `json:"trace"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}
	var line []byte
	line, err := json.MarshalIndent(aux, "", "\t")

	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(line)
}

func (l *Logger) Write(p []byte) (int, error) {
	return l.print(LevelError, string(p), nil)
}
