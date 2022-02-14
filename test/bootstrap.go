package test

import (
	"fmt"
	"github.com/cyejing/shuttle/core/client"
	"github.com/cyejing/shuttle/core/server"
	"github.com/cyejing/shuttle/pkg/codec"
	clientC "github.com/cyejing/shuttle/pkg/config/client"
	serverC "github.com/cyejing/shuttle/pkg/config/server"
	"log"
	"net/http"
)

func StartServer(sf chan int, path string) {
	c, err := serverC.Load(path)
	if err != nil {
		return
	}

	srv := &server.TLSServer{
		Handler: server.NewRouteMux(c),
	}
	sf <- 1
	srv.ListenAndServe("127.0.0.1:4880")
}

func StartClient(sf chan int, path string) {
	config, err := clientC.Load(path)
	if err != nil {
		return
	}
	config.SSLEnable = false
	config.RemoteAddr = "127.0.0.1:4880"
	config.LocalAddr = "127.0.0.1:4080"

	socks5 := &client.Socks5Server{
		Config:   config,
		DialFunc: codec.DialTrojan,
	}

	sf <- 1
	socks5.ListenAndServe("tcp", config.LocalAddr)
}

func StartWeb(sf chan int) {
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
		fmt.Fprintf(w, "Query = %q\n", r.URL.Query())

	})

	sf <- 1
	http.ListenAndServe("127.0.0.1:8088", nil)
}
