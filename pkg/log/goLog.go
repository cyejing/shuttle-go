package log

import (
	"io"
	golog "log"
	"os"
)

var (
	trace  *golog.Logger
	debug  *golog.Logger
	info   *golog.Logger
	warn   *golog.Logger
	error  *golog.Logger
	fatal  *golog.Logger
	panicl *golog.Logger
)

func init() {
	os.Mkdir("logs", 0766)
	errFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		golog.Fatalln("打开日志文件失败：", err)
	}
	appFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		golog.Fatalln("打开日志文件失败：", err)
	}

	trace = golog.New(io.MultiWriter(os.Stdout, appFile), "[TRACE] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
	debug = golog.New(io.MultiWriter(os.Stdout, appFile), "[DEBUG] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
	info = golog.New(io.MultiWriter(os.Stdout, appFile), "[INFO] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
	warn = golog.New(io.MultiWriter(os.Stdout, appFile), "[WARN] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
	error = golog.New(io.MultiWriter(os.Stderr, appFile, errFile), "[ERROR] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
	fatal = golog.New(io.MultiWriter(os.Stderr, appFile, errFile), "[FATAL] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
	panicl = golog.New(io.MultiWriter(os.Stderr, appFile, errFile), "[PANIC] ", golog.Ldate|golog.Ltime|golog.Lshortfile)
}

type goLogger struct {
}

func (l *goLogger) Panic(s string) {
	panicl.Output(3, s)
}

func (l *goLogger) Fatal(s string) {
	fatal.Output(3, s)
}

func (l *goLogger) Error(s string) {
	error.Output(3, s)
}

func (l *goLogger) Warn(s string) {
	warn.Output(3, s)
}

func (l *goLogger) Info(s string) {
	info.Output(3, s)
}

func (l *goLogger) Debug(s string) {
	debug.Output(3, s)
}

func (l *goLogger) Trace(s string) {
	trace.Output(3, s)
}
