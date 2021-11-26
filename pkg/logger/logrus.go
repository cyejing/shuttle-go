package logger

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

var glog *logrus.Logger

func init() {
	os.Mkdir("logs", 0766)
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	glog = logrus.New()
	glog.SetOutput(io.MultiWriter(os.Stdout, file))
	glog.SetLevel(logrus.DebugLevel)
	glog.SetReportCaller(false)
	glog.SetFormatter(&nested.Formatter{
		TimestampFormat: time.RFC3339,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funs := strings.Split(frame.Function, "/")
			return fmt.Sprintf(" [ %s:%d %s ]", frame.File, frame.Line, funs[len(funs)-1])
		},
	})
}

// NewLog for packages
func NewLog() *logrus.Logger {
	return glog
}
