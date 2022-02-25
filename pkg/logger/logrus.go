package logger

import (
	"fmt"
	"github.com/cyejing/shuttle/pkg/errors"
	"github.com/cyejing/shuttle/pkg/utils"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

var glog = initMultiLog(os.Stdout)

func initMultiLog(writers ...io.Writer) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.MultiWriter(writers...))
	log.SetLevel(logrus.DebugLevel)
	log.SetReportCaller(false)
	log.SetFormatter(&Formatter{
		NoFieldsColors:  true,
		TimestampFormat: time.RFC3339,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funs := strings.Split(frame.Function, "/")
			return fmt.Sprintf(" [ %s:%d %s ]", frame.File, frame.Line, funs[len(funs)-1])
		},
	})
	return log
}

//InitLog init log file
func InitLog(file string) error {
	path, _, err := utils.SplitPathAndFile(file)
	if err != nil {
		return errors.BaseErr("split file path fail", err)
	}
	err = os.MkdirAll(path, 0766)
	if err != nil {
		return errors.BaseErr("mkdir path fail", err)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	glog.Infof("log in file: %s", f.Name())
	glog.SetOutput(io.MultiWriter(os.Stdout, f))
	return nil
}

// NewLog for packages
func NewLog() *logrus.Logger {
	return glog
}
