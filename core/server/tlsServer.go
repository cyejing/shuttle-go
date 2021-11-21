package server

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/cyejing/shuttle/core/filter"
	config "github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/log"
	"net"
	"net/http"
	"time"
)

type Protocol string

const (
	HTTP   = Protocol("http")
	TROJAN = Protocol("trojan")
)

type conn struct {
	rwc        net.Conn
	remoteAddr string
	bufr       *bufio.Reader
	bufw       *bufio.Writer
	handler    http.Handler
}

type response struct {
	c          *conn
	proto      Protocol
	statusCode int
	header     http.Header
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	c := config.GetConfig()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panic("启动服务失败,端口监听异常", err)
	}

	if c.Ssl.Enable {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Panic("启动服务失败,请检查证书文件")
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		ln, err = tls.Listen("tcp", ":8890", config)
		if err != nil {
			log.Panic("启动服务失败,端口监听异常")
		}
	}
	defer ln.Close()

	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Error("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
		}
		go newConn(conn, handler).handle()
	}
}

func newConn(c net.Conn, h http.Handler) *conn {
	return &conn{
		rwc:     c,
		handler: h,
	}
}

func (c *conn) handle() {
	c.remoteAddr = c.rwc.RemoteAddr().String()
	c.bufr = bufio.NewReader(c.rwc)
	c.bufw = bufio.NewWriterSize(c.rwc, 4<<10)

	peek, err := c.bufr.Peek(56)
	if err != nil {
		log.Error("预览前置字节出错", err)

	}
	proto := HTTP
	if filter.PeekTrojanProto(peek) {
		proto = TROJAN
	}

	req, err := http.ReadRequest(c.bufr)
	if err != nil {
		log.Error("读取http请求错误", err)
	}
	resp := &response{
		c:          c,
		proto:      proto,
		statusCode: 200,
		header:     make(http.Header),
	}
	c.handler.ServeHTTP(resp, req)

	resp.finishRequest()
}

func (r *response) Header() http.Header {
	return r.header
}

func (r *response) Write(bytes []byte) (int, error) {
	switch r.proto {
	case TROJAN:
		return r.c.bufw.Write(bytes)
	case HTTP:
		r.writeHeader()
		return r.writeBody(bytes)
	default:
		return 0, errors.New("未知的协议")
	}
}

func (r *response) WriteHeader(statusCode int) {
	switch r.proto {
	case TROJAN:
	case HTTP:
		r.statusCode = statusCode
	}

}

func (r *response) writeHeader() {

}

func (r *response) writeBody(bytes []byte) (int, error) {
	return 0, nil
}

func (r *response) finishRequest() {
	r.c.bufw.Flush()
}
