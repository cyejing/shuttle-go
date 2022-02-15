package wormhole

import (
	"github.com/cyejing/shuttle/pkg/operate"
	"github.com/cyejing/shuttle/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConnectOP(t *testing.T) {

	startFinish := make(chan int, 2)

	go test.StartServer(startFinish, "../../example/shuttles.yaml")
	<-startFinish
	go test.StartSocksClient(startFinish, "../../example/shuttlec-wormhole.yaml")
	<-startFinish

	time.Sleep(1 * time.Second)

	sd := operate.GetSerDispatcher("unique-name")
	cd := operate.GetCliDispatcher("unique-name")
	assert.NotNil(t, sd)
	assert.NotNil(t, cd)
}
