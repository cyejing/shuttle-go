package main

import (
	"github.com/cyejing/shuttle/core"
	"github.com/cyejing/shuttle/pkg/log"
)

func main() {
	log.Info.Println("hello")
	core.NewEcho(8890)
	//server, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 8890})
	//if err != nil {
	//	return
	//}
	//for {
	//	conn, err := server.AcceptTCP()
	//	if err != nil {
	//		return
	//	}
	//	go handler(conn)
	//}
}

//
//func handler(conn *net.TCPConn) {
//	scanner := bufio.NewScanner(conn)
//	for scanner.Scan() {
//		fmt.Println(scanner.Text())
//		//fmt.Println([]byte(scanner.Text()))
//		//fmt.Println(scanner.Bytes())
//		conn.Write(scanner.Bytes())
//		//conn.Write( []byte("\n"))
//	}
//
//}
