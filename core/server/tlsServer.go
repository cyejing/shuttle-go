package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/cyejing/shuttle/pkg/codec"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type Protocol string

const (
	HTTP   = Protocol("http")
	TROJAN = Protocol("trojan")
)

type TLSServer struct {
	Addr    string
	Cert    string
	Key     string
	Handler http.Handler
}

type peekReader struct {
	r *bufio.Reader
	i int
}

func (p *peekReader) Read(b []byte) (n int, err error) {
	peek, err := p.r.Peek(p.i + len(b))
	if err != nil {
		return 0, err
	}
	ci := copy(b, peek[p.i:])
	p.i += ci
	return ci, nil
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

func (r *response) WriteHeader(statusCode int) {
	r.resp.StatusCode = statusCode
}

func (s *TLSServer) ListenAndServeTLS() error {
	cert, err := tls.LoadX509KeyPair(s.Cert, s.Key)
	if err != nil {
		return utils.NewError("start TLSServer fail, check cert and key").Base(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", s.Addr, config)
	//ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return utils.NewError("start TLSServer fail").Base(err)
	}
	defer ln.Close()
	log.Infof("server listen at %s", s.Addr)

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
				log.Error("accept tls conn fail |", err)
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
					log.Error("tls server handle conn fail |", err)
				}
				return
			}
		}()
	}
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
		return err
	}

	err = peekTrojan(bufr, c.rwc)
	if err != nil {
		return err
	}

	req, err := http.ReadRequest(bufr)
	if err != nil {
		return err
	}
	resp := newResponse(req)

	c.req = req
	c.resp = resp.resp

	c.handler.ServeHTTP(resp, req)

	err = c.finishRequest()
	if err != nil {
		return err
	}

	return nil
}

func (c *conn) handshakeCheck() error {
	ctx := context.Background()
	if tlsConn, ok := c.rwc.(*tls.Conn); ok {
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn's underlying net.Conn.
			if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil && tlsRecordHeaderLooksLikeHTTP(re.RecordHeader) {
				io.WriteString(re.Conn, "HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")
				re.Conn.Close()
				return err
			}
			log.Warnf("http: TLS handshake error from %s: %v", c.rwc.RemoteAddr(), err)
			return err
		}
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

func (c *conn) finishRequest() error {
	body, err := io.ReadAll(c.resp.Body)
	if err != nil {
		return err
	}

	c.resp.ContentLength = int64(len(body))
	c.resp.Body = io.NopCloser(bytes.NewReader(body))

	if c.resp.Header.Get("Content-Type") == "" {
		c.resp.Header.Set("Content-Type", http.DetectContentType(body))
	}
	c.resp.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	err = c.resp.Write(c.rwc)
	if err != nil {
		return err
	}
	return nil
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

func peekTrojan(bufr *bufio.Reader, conn net.Conn) error {
	peek, err := bufr.Peek(56)
	if err != nil {
		return err
	}
	if pw, ok := codec.ExitHash(peek); ok {
		log.Infof("%s authenticated as %s", conn.RemoteAddr(), pw.Raw)
		trojan := codec.Trojan{}
		pr := &peekReader{r: bufr}
		err := trojan.Decode(pr)
		if err != nil {
			log.Warnf("trojan proto decode fail %v", err)
			return nil
		} else {
			_, err := bufr.Discard(pr.i)
			if err != nil {
				log.Warnf("Discard trojan proto fail %v", err)
				return nil
			}
			outbound, err := net.Dial("tcp", trojan.Metadata.Address.String())
			if err != nil {
				return utils.NewError(fmt.Sprintf("trojan dial addr fail %v", trojan.Metadata.Address.String())).Base(err)
			}
			log.Infof("trojan %s requested connection to %s", conn.RemoteAddr(), trojan.Metadata.String())

			defer outbound.Close()
			return utils.ProxyStreamBuf(bufr, conn, outbound, outbound)
		}
	}
	return nil
}

func logReqDump(req *http.Request) {
	respBytes, _ := httputil.DumpRequest(req, true)
	log.Debugf("\n%s", hex.Dump(respBytes))
}
func logRespDump(resp *http.Response) {
	respBytes, _ := httputil.DumpResponse(resp, true)
	log.Debugf("\n%s", hex.Dump(respBytes))
}
