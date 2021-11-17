package main

import (
	"fmt"
	"github.com/cyejing/shuttle/pkg/socks5"
)

func main() {
	fmt.Println("start socks5 port: 1220")
	socks5.New(1220).Listen()
}
