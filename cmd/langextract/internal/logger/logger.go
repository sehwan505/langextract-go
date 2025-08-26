package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// Logger provides structured logging capabilities for the CLI
type Logger struct {
	*logrus.Logger
	colored bool
}

// NewLogger creates a new logger with the specified level and format
func NewLogger(level, format string) (*Logger, error) {
	log := logrus.New()
	
	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", level, err)
	}
	log.SetLevel(logLevel)
	
	// Configure formatter
	colored := shouldUseColor()
	switch strings.ToLower(format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case "text":
		log.SetFormatter(&TextFormatter{
			ForceColors:     colored,
			DisableColors:   !colored,
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
		})
	default:
		return nil, fmt.Errorf("invalid log format '%s', must be 'json' or 'text'", format)
	}
	
	// Set output to stderr for logs
	log.SetOutput(os.Stderr)
	
	return &Logger{
		Logger:  log,
		colored: colored,
	}, nil
}

// shouldUseColor determines if colored output should be used
func shouldUseColor() bool {
	// Check if NO_COLOR is set (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	// Check if FORCE_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	
	// Check if stderr is a terminal
	return isTerminal(os.Stderr)
}

// isTerminal checks if the given file is a terminal
func isTerminal(f *os.File) bool {
	// This is a simplified check - in production you might want to use
	// a more robust terminal detection library
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// TextFormatter is a custom logrus formatter for human-readable output
type TextFormatter struct {
	ForceColors     bool
	DisableColors   bool
	FullTimestamp   bool
	TimestampFormat string
}

// Format formats the log entry
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b strings.Builder
	
	useColors := f.ForceColors || (!f.DisableColors && isTerminal(os.Stderr))
	
	// Timestamp
	if f.FullTimestamp {
		timestamp := entry.Time.Format(f.TimestampFormat)
		if useColors {
			b.WriteString(color.New(color.FgHiBlack).Sprint(timestamp))
		} else {
			b.WriteString(timestamp)
		}
		b.WriteString(" ")
	}
	
	// Level
	level := strings.ToUpper(entry.Level.String())
	if useColors {
		levelColor := getLevelColor(entry.Level)
		b.WriteString(levelColor.Sprintf("%-5s", level))
	} else {
		b.WriteString(fmt.Sprintf("%-5s", level))
	}
	b.WriteString(" ")
	
	// Message
	b.WriteString(entry.Message)
	
	// Fields
	if len(entry.Data) > 0 {
		b.WriteString(" ")
		for key, value := range entry.Data {
			if useColors {
				b.WriteString(color.New(color.FgCyan).Sprint(key))
				b.WriteString("=")
				b.WriteString(fmt.Sprintf("%v", value))
			} else {
				b.WriteString(fmt.Sprintf("%s=%v", key, value))
			}
			b.WriteString(" ")
		}
	}
	
	b.WriteString("\n")
	return []byte(b.String()), nil
}

// getLevelColor returns the color for the log level
func getLevelColor(level logrus.Level) *color.Color {
	switch level {
	case logrus.DebugLevel:
		return color.New(color.FgMagenta)
	case logrus.InfoLevel:
		return color.New(color.FgBlue)
	case logrus.WarnLevel:
		return color.New(color.FgYellow)
	case logrus.ErrorLevel:
		return color.New(color.FgRed)
	case logrus.FatalLevel:
		return color.New(color.FgRed, color.Bold)
	case logrus.PanicLevel:
		return color.New(color.FgRed, color.Bold, color.BlinkSlow)
	default:
		return color.New(color.FgWhite)
	}
}

// LoggerEntry extends logrus.Entry with custom methods
type LoggerEntry struct {
	*logrus.Entry
	colored bool
}

// WithFile creates a logger entry with file context
func (le *LoggerEntry) WithFile(filename string) *LoggerEntry {
	return &LoggerEntry{Entry: le.Entry.WithField("file", filename), colored: le.colored}
}

// WithCount creates a logger entry with count context
func (le *LoggerEntry) WithCount(count int) *LoggerEntry {
	return &LoggerEntry{Entry: le.Entry.WithField("count", count), colored: le.colored}
}

// WithProvider creates a logger entry with provider context
func (le *LoggerEntry) WithProvider(provider string) *LoggerEntry {
	return &LoggerEntry{Entry: le.Entry.WithField("provider", provider), colored: le.colored}
}

// WithDuration creates a logger entry with duration context
func (le *LoggerEntry) WithDuration(duration time.Duration) *LoggerEntry {
	return &LoggerEntry{Entry: le.Entry.WithField("duration", duration.String()), colored: le.colored}
}

// Success logs a success message with green color
func (le *LoggerEntry) Success(message string) {
	if le.colored {
		le.Info(color.GreenString("✓ " + message))
	} else {
		le.Info("✓ " + message)
	}
}

// WithOperation creates a logger with operation context
func (l *Logger) WithOperation(operation string) *LoggerEntry {
	return &LoggerEntry{Entry: l.WithField("operation", operation), colored: l.colored}
}

// WithProvider creates a logger with provider context
func (l *Logger) WithProvider(provider string) *LoggerEntry {
	return &LoggerEntry{Entry: l.WithField("provider", provider), colored: l.colored}
}

// WithFile creates a logger with file context
func (l *Logger) WithFile(filename string) *LoggerEntry {
	return &LoggerEntry{Entry: l.WithField("file", filename), colored: l.colored}
}

// WithRequest creates a logger with request context
func (l *Logger) WithRequest(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// WithDuration creates a logger with duration context
func (l *Logger) WithDuration(duration time.Duration) *logrus.Entry {
	return l.WithField("duration", duration.String())
}

// WithCount creates a logger with count context
func (l *Logger) WithCount(count int) *LoggerEntry {
	return &LoggerEntry{Entry: l.WithField("count", count), colored: l.colored}
}

// WithError creates a logger with error context
func (l *Logger) WithError(err error) *logrus.Entry {
	if err != nil {
		return l.WithField("error", err.Error())
	}
	return l.WithFields(logrus.Fields{})
}

// Progress logs a progress message
func (l *Logger) Progress(message string, current, total int) {
	percentage := float64(current) / float64(total) * 100
	l.WithFields(logrus.Fields{
		"current":    current,
		"total":      total,
		"percentage": fmt.Sprintf("%.1f%%", percentage),
	}).Info(message)
}

// Success logs a success message with green color
func (l *Logger) Success(message string) {
	if l.colored {
		l.Info(color.GreenString("✓ " + message))
	} else {
		l.Info("✓ " + message)
	}
}

// Warning logs a warning message with yellow color
func (l *Logger) Warning(message string) {
	if l.colored {
		l.Warn(color.YellowString("⚠ " + message))
	} else {
		l.Warn("⚠ " + message)
	}
}

// Error logs an error message with red color
func (l *Logger) Error(message string) {
	if l.colored {
		l.Logger.Error(color.RedString("✗ " + message))
	} else {
		l.Logger.Error("✗ " + message)
	}
}

// Fatal logs a fatal message with red color and exits
func (l *Logger) Fatal(message string) {
	if l.colored {
		l.Logger.Fatal(color.RedString("✗ " + message))
	} else {
		l.Logger.Fatal("✗ " + message)
	}
}

// SetOutput sets the output destination for logs
func (l *Logger) SetOutput(w io.Writer) {
	l.Logger.SetOutput(w)
}

// SetQuiet disables all log output except errors
func (l *Logger) SetQuiet(quiet bool) {
	if quiet {
		l.SetLevel(logrus.ErrorLevel)
	}
}

// SetVerbose enables verbose logging (debug level)
func (l *Logger) SetVerbose(verbose bool) {
	if verbose {
		l.SetLevel(logrus.DebugLevel)
	}
}