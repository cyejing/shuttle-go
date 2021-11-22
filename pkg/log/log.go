package log

import (
	"fmt"
	"os"
	gdebug "runtime/debug"
)

type LogLevel int

const (
	AllLevel   LogLevel = 0
	InfoLevel  LogLevel = 1
	WarnLevel  LogLevel = 2
	ErrorLevel LogLevel = 3
	FatalLevel LogLevel = 4
	OffLevel   LogLevel = 5
)

type Logger interface {
	Panic(s string)
	Fatal(s string)
	Error(s string)
	Warn(s string)
	Info(s string)
	Debug(s string)
	Trace(s string)
}

var logger Logger = &goLogger{}
var logLevel = AllLevel

func SetLogLevel(level LogLevel) {
	logLevel = level
}

func Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	if logLevel <= FatalLevel {
		logger.Panic(s)
	}
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	panicl.Output(3, s)
	if logLevel <= FatalLevel {
		logger.Panic(s)
	}
	panic(s)
}

func Fatal(v ...interface{}) {
	if logLevel <= FatalLevel {
		v := append(v, "\n"+string(gdebug.Stack()))
		logger.Fatal(fmt.Sprintln(v...))
	}
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	if logLevel <= FatalLevel {
		v := append(v, string(gdebug.Stack()))
		logger.Fatal(fmt.Sprintf(format+"\n%s", v...))
	}
	os.Exit(1)
}

func Error(v ...interface{}) {
	if logLevel <= ErrorLevel {
		v := append(v, "\n"+string(gdebug.Stack()))
		logger.Error(fmt.Sprintln(v...))
	}
}

func Errorf(format string, v ...interface{}) {
	if logLevel <= ErrorLevel {
		v := append(v, string(gdebug.Stack()))
		logger.Error(fmt.Sprintf(format+"\n%s", v...))
	}
}

func Warn(v ...interface{}) {
	if logLevel <= WarnLevel {
		logger.Warn(fmt.Sprintln(v...))
	}
}

func Warnf(format string, v ...interface{}) {
	if logLevel <= WarnLevel {
		logger.Warn(fmt.Sprintf(format, v...))
	}
}

func Info(v ...interface{}) {
	if logLevel <= InfoLevel {
		logger.Info(fmt.Sprintln(v...))
	}
}

func Infof(format string, v ...interface{}) {
	if logLevel <= InfoLevel {
		logger.Info(fmt.Sprintf(format, v...))
	}
}

func Debug(v ...interface{}) {
	if logLevel <= AllLevel {
		logger.Debug(fmt.Sprintln(v...))
	}
}

func Debugf(format string, v ...interface{}) {
	if logLevel <= AllLevel {
		logger.Debug(fmt.Sprintf(format, v...))
	}
}

func Trace(v ...interface{}) {
	if logLevel <= AllLevel {
		logger.Trace(fmt.Sprintln(v...))
	}
}

func Tracef(format string, v ...interface{}) {
	if logLevel <= AllLevel {
		logger.Trace(fmt.Sprintf(format, v...))
	}
}
