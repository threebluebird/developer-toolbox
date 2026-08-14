// Package logger provides the small, dependency-free logging facade used by
// the toolbox core. Tool implementations can receive Logger without depending
// on a concrete logging backend.
package logger

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func ParseLevel(value string) Level {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "DEBUG":
		return DebugLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

type Logger interface {
	Debug(message string, args ...any)
	Info(message string, args ...any)
	Warn(message string, args ...any)
	Error(message string, args ...any)
}

// StandardLogger is a concurrency-safe leveled logger backed by an io.Writer.
type StandardLogger struct {
	mu       sync.Mutex
	minimum  Level
	delegate *log.Logger
}

func New(writer io.Writer, minimum Level) *StandardLogger {
	return &StandardLogger{
		minimum:  minimum,
		delegate: log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds),
	}
}

func (l *StandardLogger) Debug(message string, args ...any) {
	l.write(DebugLevel, "DEBUG", message, args...)
}
func (l *StandardLogger) Info(message string, args ...any) {
	l.write(InfoLevel, "INFO", message, args...)
}
func (l *StandardLogger) Warn(message string, args ...any) {
	l.write(WarnLevel, "WARN", message, args...)
}
func (l *StandardLogger) Error(message string, args ...any) {
	l.write(ErrorLevel, "ERROR", message, args...)
}

func (l *StandardLogger) write(level Level, label, message string, args ...any) {
	if l == nil || level < l.minimum {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Printf("[%s] %s", label, fmt.Sprintf(message, args...))
}
