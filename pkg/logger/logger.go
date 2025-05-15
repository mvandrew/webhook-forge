package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// logger implements the Logger interface
type logger struct {
	level  Level
	format string
	output io.Writer
	fields map[string]interface{}
}

// New creates a new logger
func New(level string, format string, output io.Writer) Logger {
	lvl := parseLevel(level)
	return &logger{
		level:  lvl,
		format: format,
		output: output,
		fields: make(map[string]interface{}),
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
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
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
