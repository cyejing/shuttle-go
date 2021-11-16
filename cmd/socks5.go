package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var logger *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

const (
	socks5Version = uint8(5)
)

func main() {
	fmt.Println("start socks5 port: 1220")
	Socks5Server(1220)
}

func Socks5Server(port int) {
	server, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	for true {
		conn, err := server.Accept()
		if err != nil {
			fmt.Errorf("accept conn err %v", err)
		}

		go handleSocks5Conn(conn)
	}
}

func handleSocks5Conn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	version := []byte{0}
	_, err := reader.Read(version)
	if err != nil {
		logger.Println("[ERR] socks: Failed to get version byte: %v", err)
		return
	}

	if version[0] != socks5Version {
		err := fmt.Errorf("unsupported SOCKS version: %v", version)
		logger.Printf("[ERR] socks: %v", err)
		return
	}

}
