package utils

import (
	"github.com/cyejing/shuttle/pkg/logger"
	"github.com/google/gops/agent"
)
var log = logger.NewLog()

func OpenAgent() {
	log.Debug("open debug mode")
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}
}
