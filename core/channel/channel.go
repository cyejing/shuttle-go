package channel

import "github.com/cyejing/shuttle/pkg/logger"

var log = logger.NewLog()

type Channel interface {
	Run()
}
