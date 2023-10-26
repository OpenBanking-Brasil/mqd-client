package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel - Custom type to hold value for weekday ranging from 1-7
type LogLevel int

// Declare related constants for each LogLevel starting with index 0
const (
	DebugLevel LogLevel = iota // DebugLevel defines debug log level.
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
	TraceLevel LogLevel = -1
)

// Func: SetLoggingGlobalLevel Sets the global level for the globbing feature
// @author AB
// @params
// level: logging Level to be configured
// @return
func SetLoggingGlobalLevel(level LogLevel) {
	zerolog.SetGlobalLevel(zerolog.Level(level))
}

// Func: GetLoggingGlobalLevel Gets the global level for the globbing feature
// @author AB
// @params
// @return
// LogLevel: Actual logging level
func GetLoggingGlobalLevel() LogLevel {
	return LogLevel(zerolog.GlobalLevel())
}

// Func: SetLoggingGlobalLevelFromString Sets the global level for the globbing feature based on a string,
// in the case the string is not recognized the default value ERROR will be used
// @author AB
// @params
// level: logging Level string to be configured
// @return
func SetLoggingGlobalLevelFromString(level string) {
	switch level {
	case "DEBUG":
		SetLoggingGlobalLevel(DebugLevel)
	case "INFO":
		SetLoggingGlobalLevel(InfoLevel)
	case "WARNING":
		SetLoggingGlobalLevel(WarnLevel)
	case "ERROR":
		SetLoggingGlobalLevel(ErrorLevel)
	case "FATAL":
		SetLoggingGlobalLevel(FatalLevel)
	case "PANIC":
		SetLoggingGlobalLevel(PanicLevel)
	case "DISABLED":
		SetLoggingGlobalLevel(Disabled)
	case "TRACE":
		SetLoggingGlobalLevel(TraceLevel)
	default:
		SetLoggingGlobalLevel(ErrorLevel)
	}
}

// Func: Trace writes a message to the TRACE level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func Trace(message string, pack string, component string) {
	log.Trace().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the LOG level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func Log(message string, pack string, component string) {
	log.Log().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the DEBUG level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func Debug(message string, pack string, component string) {
	log.Debug().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the INFO level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func Info(message string, pack string, component string) {
	log.Info().Str("package", pack).Str("component", component).Msg(message)
}

// Func: Trace writes a message to the WARNING level
// @author AB
// @params
// message: message to write
// pack: name of the package where the log is called
// component: name of the package where the log is called
// @return
func Warning(message string, pack string, component string) {
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
func Error(err error, message string, pack string, component string) {
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
func Fatal(err error, message string, pack string, component string) {
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
func Panic(message string, pack string, component string) {
	log.Panic().Str("package", pack).Str("component", component).Msg(message)
}
