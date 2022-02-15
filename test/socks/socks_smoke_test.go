package socks

import (
	"github.com/cyejing/shuttle/test"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestSocksRequest(t *testing.T) {
	startFinish := make(chan int, 3)
	go test.StartWeb(startFinish)
	<-startFinish
	go test.StartServer(startFinish,"../../example/shuttles.yaml")
	<-startFinish
	go test.StartSocksClient(startFinish,"../../example/shuttlec-socks.yaml")
	<-startFinish

	request, err := http.NewRequest("GET", "http://127.0.0.1:8088", nil)
	if err != nil {
		t.Error("new request fail", err)
		return
	}

	cli := &http.Client{
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse("socks5://127.0.0.1:4080")
			},
		},
	}
	log.Println(cli)

	resp, err := cli.Do(request)
	if err != nil {
		t.Error("request do fail", err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode() = %v, want %v", resp.StatusCode, 22)
		return
	}
}