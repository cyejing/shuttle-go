package test

import (
	"fmt"
	"github.com/cyejing/shuttle/core/client"
	"github.com/cyejing/shuttle/core/server"
	clientC "github.com/cyejing/shuttle/pkg/config/client"
	serverC "github.com/cyejing/shuttle/pkg/config/server"
	"log"
	"net/http"
	"time"
)

func StartServer(sf chan int, path string) {
	c, err := serverC.Load(path)
	if err != nil {
		return
	}
	c.Addr = "127.0.0.1:4880"
	c.Cert = "../../example/s.cyejing.cn_chain.crt"
	c.Key = "../../example/s.cyejing.cn_key.key"

	time.Sleep(1 * time.Second)
	sf <- 1
	server.Run(c)
}

func StartSocksClient(sf chan int, path string) {
	c, err := clientC.Load(path)
	if err != nil {
		return
	}
	c.SSLEnable = false
	c.RemoteAddr = "127.0.0.1:4880"
	c.LocalAddr = "127.0.0.1:4080"

	sf <- 1
	client.Run(c)
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
