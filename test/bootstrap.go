package test

import (
	"fmt"
	"github.com/cyejing/shuttle/core/client"
	clientC "github.com/cyejing/shuttle/core/config/client"
	serverC "github.com/cyejing/shuttle/core/config/server"
	"github.com/cyejing/shuttle/core/server"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func StartServer(sf chan int, path string) {
	c, err := serverC.Load(path)
	if err != nil {
		return
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		sf <- 1
	}()
	server.Run(c)
}

func StartSocksClient(sf chan int, path string) {
	c, err := clientC.Load(path)
	if err != nil {
		return
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		sf <- 1
	}()
	client.Run(c)
}

func StartWeb(sf chan int) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	}))
	go func() {
		time.Sleep(100 * time.Millisecond)
		sf <- 1
	}()
	http.ListenAndServe("127.0.0.1:8088", mux)
}

func StartEcho(sf chan int) {
	server, err := net.Listen("tcp", "127.0.0.1:5010")
	if err != nil {
		return
	}
	sf <- 1
	for true {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("accept conn err %v \n", err)
		}

		go func() {
			defer conn.Close()
			_, err := io.Copy(conn, conn)
			if err != nil {
				return
			}
		}()
	}
	return
}
