package wormhole

import (
	"fmt"
	"github.com/cyejing/shuttle/core/controller"
	"github.com/cyejing/shuttle/pkg/operate"
	"github.com/cyejing/shuttle/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func startWormhole(startFinish chan int) {
	go test.StartServer(startFinish, "../../example/shuttles.yaml")
	<-startFinish
	go test.StartSocksClient(startFinish, "../../example/shuttlec-wormhole.yaml")
	<-startFinish
}

func setup() {
	fmt.Println("wormhole test setup")
	startFinish := make(chan int, 3)
	startWormhole(startFinish)
	go test.StartEcho(startFinish)
	<-startFinish
	go controller.NewProxyCtl("unique-name", "test", "127.0.0.1:4081", "127.0.0.1:5010").Run()

	time.Sleep(1 * time.Second)
}

func TestConnectOP(t *testing.T) {
	sd := operate.GetSerDispatcher("unique-name")
	cd := operate.GetCliDispatcher("unique-name")
	assert.NotNil(t, sd)
	assert.NotNil(t, cd)

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
