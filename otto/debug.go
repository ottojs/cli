package otto

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// Log levels from most to least severe
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
)

// String representation of log levels
func (l LogLevel) String() string {
	switch l {
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// Logger represents our logging configuration
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	timeFormat string
}

// Global logger instance
var globalLogger = &Logger{
	level:      INFO, // Default to INFO level
	logger:     log.New(os.Stderr, "", 0),
	timeFormat: "2006-01-02 15:04:05",
}

// SetLogLevel sets the minimum log level that will be displayed
// ERROR will only show ERROR messages
// DEBUG will show ERROR, WARN, INFO, and DEBUG messages
func SetLogLevel(level LogLevel) {
	globalLogger.level = level
}

// SetLogLevelFromString sets the log level from a string representation
func SetLogLevelFromString(level string) error {
	switch strings.ToUpper(level) {
	case "ERROR":
		globalLogger.level = ERROR
	case "WARN":
		globalLogger.level = WARN
	case "INFO":
		globalLogger.level = INFO
	case "DEBUG":
		globalLogger.level = DEBUG
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	return globalLogger.level
}

// logMessage logs a message if it meets the current log level threshold
func logMessage(level LogLevel, format string, args ...interface{}) {
	// Only log if the message level is <= current log level (lower number = higher severity)
	if level <= globalLogger.level {
		timestamp := time.Now().Format(globalLogger.timeFormat)
		message := fmt.Sprintf(format, args...)
		globalLogger.logger.Printf("[%s] [%s] %s", timestamp, level.String(), message)
	}
}

// LogError logs an error message (always displayed unless log level is set higher than ERROR)
func LogError(format string, args ...interface{}) {
	logMessage(ERROR, format, args...)
}

// LogWarn logs a warning message
func LogWarn(format string, args ...interface{}) {
	logMessage(WARN, format, args...)
}

// LogInfo logs an informational message
func LogInfo(format string, args ...interface{}) {
	logMessage(INFO, format, args...)
}

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	logMessage(DEBUG, format, args...)
}

// Log maintains backward compatibility with existing code
// It now logs at DEBUG level
func Log(v ...any) {
	LogDebug("%v", fmt.Sprint(v...))
}

// SetDebug maintains backward compatibility
// Setting debug to true sets log level to DEBUG, false sets it to INFO
func SetDebug(b bool) {
	if b {
		SetLogLevel(DEBUG)
	} else {
		SetLogLevel(INFO)
	}
}

// SetLogOutput allows changing where logs are written (default is os.Stderr)
func SetLogOutput(file *os.File) {
	globalLogger.logger = log.New(file, "", 0)
}

// SetTimeFormat allows customizing the timestamp format
func SetTimeFormat(format string) {
	globalLogger.timeFormat = format
}
