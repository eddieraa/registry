package registry

import (
	"fmt"
	golog "log"
	"os"
)

type Logger interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
}
type RegistryLogger interface {
	Logger
	SetLevel(l LogLevel)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

type LogLevel uint32

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel LogLevel = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// implement String() to LogLevel
func (l LogLevel) String() string {
	switch l {
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	case ErrorLevel:
		return "ERROR"
	case WarnLevel:
		return "WARN"
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBG"
	case TraceLevel:
		return "TRACE"
	default:
		return "INFO"

	}
}

type defaultLogger struct {
	level LogLevel
}

func NewDefaulLogger() RegistryLogger {
	return &defaultLogger{}
}

func (l *defaultLogger) log(level LogLevel, args ...any) {
	if l.level >= level {
		golog.Print(level.String(), args)
	}
}
func (l *defaultLogger) logf(level LogLevel, format string, args ...any) {
	if l.level >= level {
		l.log(level, fmt.Sprintf(format, args...))
	}
}

func (l *defaultLogger) Debug(args ...any) {
	l.log(DebugLevel, args...)
}
func (l *defaultLogger) Debugf(format string, args ...any) {
	l.logf(DebugLevel, format, args...)
}
func (l *defaultLogger) Error(args ...any) {
	l.log(ErrorLevel, args...)
}
func (l *defaultLogger) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, format, args...)
}
func (l *defaultLogger) Info(args ...any) {
	l.log(InfoLevel, args...)
}
func (l *defaultLogger) Infof(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}
func (l *defaultLogger) Warn(args ...any) {
	l.log(WarnLevel, args...)
}
func (l *defaultLogger) Fatal(args ...any) {
	l.log(FatalLevel, args...)
	os.Exit(1)
}
func (l *defaultLogger) Fatalf(format string, args ...any) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}
func (l *defaultLogger) Warnf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}
func (l *defaultLogger) SetLevel(level LogLevel) {
	l.level = level
}
