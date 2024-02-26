package log

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// JSONLogger struct in charge of logging tasks
type JSONLogger struct {
	context context.Context // Context to be used during logging
}

// GetNewJSONLogger Creates a new JSONLogger
//
// Parameters:
//
// Returns:
//   - JSONLogger: New JSONLogger
func GetNewJSONLogger() *JSONLogger {
	return &JSONLogger{}
}

// SetLoggingGlobalLevel Sets the global level for the globbing feature
//
// Parameters:
//   - level: logging Level to be configured
//
// Returns:
func (l *JSONLogger) SetLoggingGlobalLevel(level Level) {
	zerolog.SetGlobalLevel(zerolog.Level(level))
}

// GetLoggingGlobalLevel Gets the global level for the globbing feature
//
// Parameters:
//
// Returns:
//   - level: global logging level
func (l *JSONLogger) GetLoggingGlobalLevel() Level {
	return Level(zerolog.GlobalLevel())
}

// WithContext Sets the context for the logger
//
// Parameters:
//   - context: current context for the logger
//
// Returns:
//   - logger: Logger with the context assigned
func (l *JSONLogger) WithContext(context context.Context) Logger {
	l.context = context
	return l
}

// SetLoggingGlobalLevelFromString Sets the global level for the globbing feature based on a string,
// in the case the string is not recognized the default value ERROR will be used
//
// Parameters:
//   - level: logging Level string to be configured
//
// Returns:
func (l *JSONLogger) SetLoggingGlobalLevelFromString(level string) {
	switch level {
	case "DEBUG":
		l.SetLoggingGlobalLevel(DebugLevel)
	case "INFO":
		l.SetLoggingGlobalLevel(InfoLevel)
	case "WARNING":
		l.SetLoggingGlobalLevel(WarnLevel)
	case "ERROR":
		l.SetLoggingGlobalLevel(ErrorLevel)
	case "FATAL":
		l.SetLoggingGlobalLevel(FatalLevel)
	case "PANIC":
		l.SetLoggingGlobalLevel(PanicLevel)
	case "DISABLED":
		l.SetLoggingGlobalLevel(Disabled)
	case "TRACE":
		l.SetLoggingGlobalLevel(TraceLevel)
	default:
		l.SetLoggingGlobalLevel(ErrorLevel)
	}
}

// Trace writes a message to the TRACE level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Trace(message string, pack string, component string) {
	log.Trace().Str("package", pack).Str("component", component).Msg(message)
}

// Log Trace writes a message to the LOG level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Log(message string, pack string, component string) {
	log.Log().Str("package", pack).Str("component", component).Msg(message)
}

// Debug Trace writes a message to the DEBUG level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Debug(message string, pack string, component string) {
	log.Debug().Str("package", pack).Str("component", component).Msg(message)
}

// Info Trace writes a message to the INFO level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Info(message string, pack string, component string) {
	log.Info().Str("package", pack).Str("component", component).Msg(message)
}

// Warning Trace writes a message to the WARNING level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Warning(message string, pack string, component string) {
	log.Warn().Str("package", pack).Str("component", component).Msg(message)
}

// Error Trace writes a message to the ERROR level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Error(err error, message string, pack string, component string) {
	log.Error().Err(err).Str("package", pack).Str("component", component).Msg(message)
}

// Fatal Trace writes a message to the FATAL level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Fatal(err error, message string, pack string, component string) {
	log.Fatal().Err(err).Str("package", pack).Str("component", component).Msg(message)
}

// Panic Trace writes a message to the PANIC level
//
// Parameters:
//   - message: message to write
//   - pack: name of the package where the log is called
//   - component: name of the package where the log is called
//
// Returns:
func (l *JSONLogger) Panic(message string, pack string, component string) {
	log.Panic().Str("package", pack).Str("component", component).Msg(message)
}
