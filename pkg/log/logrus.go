package log

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

var L *logrus.Logger

func init() {
	os.Mkdir("logs", 0766)
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	L = logrus.New()
	L.SetOutput(io.MultiWriter(os.Stdout, file))
	L.SetLevel(logrus.DebugLevel)
	L.SetReportCaller(true)
	L.SetFormatter(&nested.Formatter{
		TimestampFormat: time.RFC3339,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funs := strings.Split(frame.Function, "/")
			return fmt.Sprintf(" [ %s:%d %s ]", frame.File, frame.Line, funs[len(funs)-1])
		},
	})
	//L.AddHook(&logrusLogger{})
}

type logrusLogger struct {
}

func (l logrusLogger) Levels() []logrus.Level {
	return logrus.AllLevels
}

var logrusPackage = "github.com/sirupsen/logrus"
var logPackage = "github.com/cyejing/shuttle/pkg/log"

func (l logrusLogger) Fire(entry *logrus.Entry) error {
	rpc := make([]uintptr, 10)
	n := runtime.Callers(8, rpc)
	if n < 1 {
		return nil
	}
	frames := runtime.CallersFrames(rpc)
	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		if pkg != logrusPackage && pkg != logPackage {
			entry.Caller = &f
			break
		}
	}
	return nil
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
