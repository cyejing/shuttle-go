package log

import (
	"io"
	"log"
	"os"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	os.Mkdir("logs", 0766)
	errFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	appFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	Debug = log.New(io.MultiWriter(os.Stdout, appFile), "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(io.MultiWriter(os.Stdout, appFile), "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(os.Stdout, appFile), "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, errFile), "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

}
