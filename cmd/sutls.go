package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	fmt.Println("hello sutls")
	a := []int{1, 2, 3}
	fmt.Println(a)

	ln, err := net.Listen("tcp", "127.0.0.1:1220")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	_, err := io.Copy(conn, conn)
	if err != nil {
		return
	}
}
