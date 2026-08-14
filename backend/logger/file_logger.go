package logger

import (
	"os"
	"path/filepath"
)

// FileLogger owns both the logging facade and its file handle.
type FileLogger struct {
	*StandardLogger
	file *os.File
}

func NewFileLogger(path string, minimum Level) (*FileLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &FileLogger{StandardLogger: New(file, minimum), file: file}, nil
}

func (l *FileLogger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}
