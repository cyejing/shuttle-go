package main

import (
	"bytes"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
	"log"
	"net/http"
	"os"
)

func main() {
	socks := &codec.Socks{
		Hash: utils.SHA224String("cyejing123"),
		Metadata: &codec.Metadata{
			Command: codec.Connect,
			Address: codec.NewAddressFromHostPort("tcp", "127.0.0.1", 8088),
		},
	}
	encode, err := socks.Encode()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", encode)
	fmt.Printf("%x\n", encode)
	fmt.Printf("%v\n", encode)
	os.WriteFile("socks.file", encode, 0666)

	fb, err := os.ReadFile("socks.file")
	if err != nil {
		return
	}
	fsocks := new(codec.Socks)
	fsocks.Decode(bytes.NewReader(fb))
	fmt.Println(fsocks)
	fmt.Printf("%v\n", fsocks)

	http.HandleFunc("/", handler)

	http.ListenAndServe(":8088", nil)
}

//!+handler
// handler echoes the HTTP request.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
	for k, v := range r.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	for k, v := range r.Form {
		fmt.Fprintf(w, "Form[%q] = %q\n", k, v)
	}
}

//!-handler
