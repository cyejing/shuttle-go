package test

import (
	"fmt"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/codec"
	clientC "github.com/cyejing/shuttle/pkg/config/client"
	serverC "github.com/cyejing/shuttle/pkg/config/server"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func startServer(sf chan int) {
	_, err := serverC.Load("../example/shuttles.yaml")
	if err != nil {
		return
	}

	srv := &server.TLSServer{
		Handler: server.NewRouteMux(),
	}
	sf <- 1
	srv.ListenAndServe("127.0.0.1:4880")
}

func startClient(sf chan int) {
	config, err := clientC.Load("../example/shuttlec-socks.yaml")
	if err != nil {
		return
	}
	config.SSLEnable = false
	config.RemoteAddr = "127.0.0.1:4880"
	config.LocalAddr = "127.0.0.1:4080"

	socks5 := &server.Socks5Server{
		DialFunc: codec.DialTrojan,
	}

	sf <- 1
	socks5.ListenAndServe("tcp", config.LocalAddr)
}

func startWeb(sf chan int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
		fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
		for k, v := range r.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
		fmt.Fprintf(w, "Host = %q\n", r.Host)
		fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr)
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		for k, v := range r.Form {
			fmt.Fprintf(w, "Form[%q] = %q\n", k, v)
		}
	})

	sf <- 1
	http.ListenAndServe("127.0.0.1:8088", nil)
}

func TestSocksRequest(t *testing.T) {
	startFinish := make(chan int, 3)
	go startWeb(startFinish)
	go startServer(startFinish)
	go startClient(startFinish)

	for i := 0; i < 3; i++ {
		<-startFinish
	}

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
