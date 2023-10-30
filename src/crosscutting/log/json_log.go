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

// Func: GetNewJSONLogger Creates a new JSONLogger
// @author AB
// @param
// @return
// JSONLogger: JSONLogger
func GetNewJSONLogger() *JSONLogger {
	return &JSONLogger{}
}

// Func: SetLoggingGlobalLevel Sets the global level for the globbing feature
// @author AB
// @params
// level: logging Level to be configured
// @return
func (l *JSONLogger) SetLoggingGlobalLevel(level LogLevel) {
	zerolog.SetGlobalLevel(zerolog.Level(level))
}

// Func: GetLoggingGlobalLevel Gets the global level for the globbing feature
// @author AB
// @params
// @return
// LogLevel: Actual logging level
func (l *JSONLogger) GetLoggingGlobalLevel() LogLevel {
	return LogLevel(zerolog.GlobalLevel())
}

// Func: WithContext Sets the context for the logger
func (l *JSONLogger) WithContext(context context.Context) Logger {
	l.context = context
	return l
}

// Func: SetLoggingGlobalLevelFromString Sets the global level for the globbing feature based on a string,
// in the case the string is not recognized the default value ERROR will be used
// @author AB
// @params
// level: logging Level string to be configured
// @return
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

// Func: Trace writes a message to the TRACE level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Trace(message string, pack string, component string) {
	log.Trace().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the LOG level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Log(message string, pack string, component string) {
	log.Log().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the DEBUG level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Debug(message string, pack string, component string) {
	log.Debug().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the INFO level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Info(message string, pack string, component string) {
	log.Info().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the WARNING level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Warning(message string, pack string, component string) {
	log.Warn().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the ERROR level
// @author AB
// @params
// err: Error with details
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Error(err error, message string, pack string, component string) {
	log.Error().Err(err).Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the FATAL level
// @author AB
// @params
// err: Error with details
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Fatal(err error, message string, pack string, component string) {
	log.Fatal().Err(err).Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the PANIC level
// @author AB
// @params
// err: Error with details
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func (l *JSONLogger) Panic(message string, pack string, component string) {
	log.Panic().Str("package", pack).Str("component", component).Msg(message)
}
