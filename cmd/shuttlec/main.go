package main

import (
	socks5 "github.com/armon/go-socks5"
)

func main() {
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		panic(err)
	}

	panic(server.ListenAndServe("tcp", "127.0.0.1:1220"))
}
