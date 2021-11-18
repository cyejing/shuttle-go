package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

//
//func main() {
//	log.Info.Println("hello")
//	//core.NewEcho(8890)
//	server, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 8890})
//	if err != nil {
//		return
//	}
//	for {
//		conn, err := server.AcceptTCP()
//		if err != nil {
//			return
//		}
//		go handleConn(conn)
//	}
//}
//
func handleConn(conn net.Conn) {
	bs := make([]byte, 56)
	conn.Read(bs)
	fmt.Println(string(bs))
	fmt.Println("================")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		//conn.Write(scanner.Bytes())
		//conn.Write( []byte("\n"))
	}

}

func main() {
	cert, err := tls.LoadX509KeyPair(
		"/Users/cyejing/Project/born/shuttle/mini.cyejing.cn_chain.crt",
		"/Users/cyejing/Project/born/shuttle/mini.cyejing.cn_key.key")
	if err != nil {
		log.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", ":8890", config)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}
