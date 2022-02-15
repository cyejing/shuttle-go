package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/operate"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

// TLSServer struct
type TLSServer struct {
	Cert    string
	Key     string
	Handler http.Handler
}

//ListenAndServeTLS serve tls
func (s *TLSServer) ListenAndServeTLS(addr string) error {
	cert, err := tls.LoadX509KeyPair(s.Cert, s.Key)

	if err != nil {
		return utils.BaseErr("start TLSServer fail, check cert and key", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return utils.BaseErr("start TLSServer fail", err)
	}
	defer ln.Close()

	return s.Server(ln)
}

//ListenAndServe listen and server addr
func (s *TLSServer) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return utils.BaseErr("start Server fail", err)
	}
	defer ln.Close()

	return s.Server(ln)
}

//Server server ln
func (s *TLSServer) Server(ln net.Listener) error {
	log.Infof("server listen at %s", ln.Addr())
	var tempDelay time.Duration
	for {
		rwc, err := ln.Accept()
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
				log.Warnf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			} else {
				log.Error("accept tls conn fail", err)
			}
		}
		c := &conn{
			rwc:     rwc,
			handler: s.Handler,
		}
		go func() {
			defer c.rwc.Close()
			err := c.handle()
			if err != nil {
				if err != io.EOF {
					log.Error("tls server handle conn fail :", err)
				}
				return
			}
		}()
	}
}

type response struct {
	resp    *http.Response
	bufBody *bytes.Buffer
}

func (r *response) Header() http.Header {
	return r.resp.Header
}

func (r *response) Write(bs []byte) (int, error) {
	return r.bufBody.Write(bs)
}

//WriteHeader write header
func (r *response) WriteHeader(statusCode int) {
	r.resp.StatusCode = statusCode
}

type conn struct {
	rwc     net.Conn
	handler http.Handler
	req     *http.Request
	resp    *http.Response
}

func (c *conn) handle() error {
	bufr := bufio.NewReader(c.rwc)

	err := c.handshakeCheck()
	if err != nil {
		return utils.BaseErr("handshake check fail", err)
	}

	ok, err := codec.PeekTrojan(bufr, c.rwc)
	if err != nil {
		return utils.BaseErr("peek trojan fail", err)
	}

	ok, err = operate.PeekWormhole(bufr, c.rwc)
	if err != nil {
		return utils.BaseErr("peek wormhole fail", err)
	}

	if !ok {
		err = c.handleHttp(err, bufr)
		if err != nil {
			return utils.BaseErr("handle http fail", err)
		}
	}

	return nil
}

func (c *conn) handleHttp(err error, bufr *bufio.Reader) error {
	req, err := http.ReadRequest(bufr)
	if err != nil {
		io.WriteString(c.rwc, "HTTP/1.0 400 Bad Request\r\n\r\nMalformed HTTP request\n")
		return utils.BaseErr("read request fail", err)
	}
	resp := newResponse(req)

	c.req = req
	c.resp = resp.resp

	c.handler.ServeHTTP(resp, req)

	err = c.finishRequest()
	if err != nil {
		return utils.BaseErr("finish request fail", err)
	}
	return nil
}

func (c *conn) handshakeCheck() error {
	ctx := context.Background()
	if tlsConn, ok := c.rwc.(*tls.Conn); ok {
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil && tlsRecordHeaderLooksLikeHTTP(re.RecordHeader) {
				io.WriteString(re.Conn, "HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")
				re.Conn.Close()
				return nil
			}
			log.Warnf("http: TLS handshake error from %s: %v", c.rwc.RemoteAddr(), err)
			return err
		}
	}
	return nil
}

func (c *conn) finishRequest() error {
	body, err := io.ReadAll(c.resp.Body)
	if err != nil {
		return utils.BaseErr("read body fail", err)
	}

	c.resp.ContentLength = int64(len(body))
	c.resp.Body = io.NopCloser(bytes.NewReader(body))

	if c.resp.Header.Get("Content-Type") == "" {
		c.resp.Header.Set("Content-Type", http.DetectContentType(body))
	}
	c.resp.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	err = c.resp.Write(c.rwc)
	if err != nil {
		return utils.BaseErr("response write fail", err)
	}
	return nil
}

func tlsRecordHeaderLooksLikeHTTP(hdr [5]byte) bool {
	switch string(hdr[:]) {
	case "GET /", "HEAD ", "POST ", "PUT /", "OPTIO":
		return true
	}
	return false
}

func newResponse(req *http.Request) *response {
	buf := bytes.NewBuffer([]byte{})
	resp := &http.Response{
		Body:    io.NopCloser(buf),
		Request: req,
		Header: http.Header{
			"Connection": {"keep-alive"},
		},
		TLS: req.TLS,
	}
	resp.ProtoMajor = req.ProtoMajor
	resp.ProtoMinor = req.ProtoMinor
	resp.StatusCode = 200
	return &response{
		resp:    resp,
		bufBody: buf,
	}
}

func logReqDump(req *http.Request) {
	respBytes, _ := httputil.DumpRequest(req, true)
	log.Debugf("\n%s", hex.Dump(respBytes))
}
func logRespDump(resp *http.Response) {
	respBytes, _ := httputil.DumpResponse(resp, true)
	log.Debugf("\n%s", hex.Dump(respBytes))
}
