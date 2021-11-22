package main

import (
	"fmt"
	"github.com/cyejing/shuttle/pkg/log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)

	http.ListenAndServe("127.0.0.1:8088", nil)
}

//!+handler
// handler echoes the HTTP request.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
	for k, v := range r.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.L.Error(err)
	}
	for k, v := range r.Form {
		fmt.Fprintf(w, "Form[%q] = %q\n", k, v)
	}
}

//!-handler
