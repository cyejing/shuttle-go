package main

import (
	"fmt"
	"github.com/cyejing/shuttle/pkg/log"
	"net/http"
)

func main() {
	http.HandleFunc("/", writeSelf)
	http.ListenAndServe("127.0.0.1:8890", nil)
}

func writeSelf(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("request %s", req.RequestURI)
	fmt.Fprintf(resp, "%s %s %s\n", req.Method, req.URL, req.Proto)
	for k, v := range req.Header {
		fmt.Fprintf(resp, "Header[%q] = %q\n", k, v)
	}
	fmt.Fprintf(resp, "Host = %q\n", req.Host)
	fmt.Fprintf(resp, "RemoteAddr = %q\n", req.RemoteAddr)
	if err := req.ParseForm(); err != nil {
	}
	for k, v := range req.Form {
		fmt.Fprintf(resp, "Form[%q] = %q\n", k, v)
	}
}
