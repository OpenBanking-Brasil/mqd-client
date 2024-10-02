package log

import "context"

// Level - Custom type to hold value for weekday ranging from 1-7
type Level int8

// Declare related constants for each LogLevel starting with index 0
const (
	DebugLevel Level = iota // DebugLevel defines debug log level.
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
)

// Logger - Interface to log
type Logger interface {
	WithContext(context context.Context) Logger                     // indicates wich context to use
	SetLoggingGlobalLevel(level Level)                              // Sets the global level for the globbing feature
	SetLoggingGlobalLevelFromString(level string)                   // Sets the global level for the globbing feature based on a string,
	GetLoggingGlobalLevel() Level                                   // Gets the global level for the globbing feature
	Trace(message string, pack string, component string)            // Trace logs a message at trace level.
	Log(message string, pack string, component string)              // Trace writes a message to the LOG level
	Debug(message string, pack string, component string)            // Trace writes a message to the DEBUG level
	Info(message string, pack string, component string)             // Trace writes a message to the INFO level
	Warning(message string, pack string, component string)          // Trace writes a message to the WARNING level
	Error(err error, message string, pack string, component string) // Trace writes a message to the ERROR level
	Fatal(err error, message string, pack string, component string) // Trace writes a message to the FATAL level
	Panic(message string, pack string, component string)            // Trace writes a message to the PANIC level
}
