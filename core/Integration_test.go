package core

import (
	"fmt"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/codec"
	clientC "github.com/cyejing/shuttle/pkg/config/client"
	serverC "github.com/cyejing/shuttle/pkg/config/server"
	"log"
	"net/http"
	"testing"
	"time"
)

func startServer() {
	_, err := serverC.Load("../example/shuttles.yaml")
	if err != nil {
		return
	}

	srv := &server.TLSServer{
		Handler: server.NewRouteMux(),
	}
	srv.ListenAndServe("127.0.0.1:4880")
}

func startClient() {
	config, err := clientC.Load("../example/shuttlec.yaml")
	if err != nil {
		return
	}
	config.SSLEnable = false
	config.RemoteAddr = "127.0.0.1:4880"

	socks5 := &server.Socks5Server{
		DialFunc: codec.DialTrojan,
	}
	socks5.ListenAndServe("tcp", "127.0.0.1:4080")
}

func startWeb() {
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

	http.ListenAndServe("127.0.0.1:8088", nil)
}

func TestSocksRequest(t *testing.T) {
	go startWeb()
	go startServer()
	go startClient()

	time.Sleep(time.Second * 3)

	//request, err := http.NewRequest("GET","127.0.0.1:8088",nil)
	//if err != nil {
	//	return
	//}
	//
	//http.DefaultClient
}
