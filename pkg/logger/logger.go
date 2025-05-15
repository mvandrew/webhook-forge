package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log levels
type Level int

const (
	// DebugLevel is for detailed debug information
	DebugLevel Level = iota
	// InfoLevel is for general information
	InfoLevel
	// WarnLevel is for warning information
	WarnLevel
	// ErrorLevel is for error information
	ErrorLevel
	// FatalLevel is for fatal errors - will exit the program
	FatalLevel
)

// Logger is the interface for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithField(key string, value interface{}) Logger
	Close() error
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// LogConfig represents configuration options for the logger
type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	FilePath   string `json:"file_path"`
	MaxSize    int64  `json:"max_size"`    // Max size in MB
	MaxBackups int    `json:"max_backups"` // Max number of rotated files to keep
}

// rotateWriter implements io.WriteCloser with log rotation capabilities
type rotateWriter struct {
	filePath   string
	maxSize    int64 // in bytes
	maxBackups int
	size       int64
	file       *os.File
	mu         sync.Mutex
}

// newRotateWriter creates a new rotate writer
func newRotateWriter(filePath string, maxSize int64, maxBackups int) (*rotateWriter, error) {
	// Convert maxSize from MB to bytes
	maxSize = maxSize * 1024 * 1024

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	return &rotateWriter{
		filePath:   filePath,
		maxSize:    maxSize,
		maxBackups: maxBackups,
		size:       info.Size(),
		file:       file,
	}, nil
}

// Write implements io.Writer
func (w *rotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Close implements io.Closer
func (w *rotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}

// rotate rotates the current log file
func (w *rotateWriter) rotate() error {
	// Close current file
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	// Rotate existing backup files
	for i := w.maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.filePath, i)
		newPath := fmt.Sprintf("%s.%d", w.filePath, i+1)

		// Remove the oldest backup if we're at max
		if i == w.maxBackups-1 {
			os.Remove(newPath)
		}

		// Rename the backups
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Rename current log file to .1
	if err := os.Rename(w.filePath, w.filePath+".1"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Create new log file
	file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	// Reset size and update file
	w.size = 0
	w.file = file

	return nil
}

// logger implements the Logger interface
type logger struct {
	level  Level
	format string
	output io.Writer
	fields map[string]interface{}
	closer io.Closer
}

// cleanupFileWriter gets a closer from a writer if it's a closable resource
func cleanupFileWriter(writer io.Writer) io.Closer {
	if closer, ok := writer.(io.Closer); ok {
		return closer
	}
	return nil
}

// NewWithConfig creates a new logger with a configuration
func NewWithConfig(config LogConfig) (Logger, error) {
	lvl := parseLevel(config.Level)

	var writer io.Writer
	var closer io.Closer

	// Use file if path is provided, otherwise use stdout
	if config.FilePath != "" {
		// Default values if not specified
		maxSize := config.MaxSize
		if maxSize <= 0 {
			maxSize = 100 // Default 100MB
		}

		maxBackups := config.MaxBackups
		if maxBackups <= 0 {
			maxBackups = 5 // Default 5 backups
		}

		fileWriter, err := newRotateWriter(config.FilePath, maxSize, maxBackups)
		if err != nil {
			return nil, fmt.Errorf("failed to create log writer: %w", err)
		}

		writer = fileWriter
		closer = fileWriter
	} else {
		writer = os.Stdout
	}

	return &logger{
		level:  lvl,
		format: config.Format,
		output: writer,
		fields: make(map[string]interface{}),
		closer: closer,
	}, nil
}

// New creates a new logger
func New(level string, format string, output io.Writer) Logger {
	lvl := parseLevel(level)
	return &logger{
		level:  lvl,
		format: format,
		output: output,
		fields: make(map[string]interface{}),
		closer: cleanupFileWriter(output),
	}
}

// Default creates a default logger
func Default() Logger {
	return New("info", "json", os.Stdout)
}

// parseLevel parses the level string
func parseLevel(level string) Level {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...Field) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, fields...)
	}
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...Field) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, fields...)
	}
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...Field) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, fields...)
	}
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...Field) {
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string, fields ...Field) {
	if l.level <= FatalLevel {
		l.log("FATAL", msg, fields...)
		os.Exit(1)
	}
}

// WithField returns a new logger with the field added
func (l *logger) WithField(key string, value interface{}) Logger {
	newLogger := &logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: make(map[string]interface{}, len(l.fields)+1),
		closer: l.closer,
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// Close implements io.Closer for cleaning up resources
func (l *logger) Close() error {
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// log logs a message with the specified level
func (l *logger) log(level string, msg string, fields ...Field) {
	entry := map[string]interface{}{
		"level":     level,
		"message":   msg,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Add default fields
	for k, v := range l.fields {
		entry[k] = v
	}

	// Add fields
	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	if l.format == "json" {
		jsonEntry, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
			return
		}
		fmt.Fprintln(l.output, string(jsonEntry))
	} else {
		// Simple text format
		fmt.Fprintf(l.output, "[%s] %s: %s\n", entry["timestamp"], level, msg)
		for k, v := range entry {
			if k != "timestamp" && k != "level" && k != "message" {
				fmt.Fprintf(l.output, "  %s: %v\n", k, v)
			}
		}
	}
}
