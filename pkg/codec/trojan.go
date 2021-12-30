package codec

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
)

var crlf = []byte{0x0d, 0x0a}

//Trojan struct
type Trojan struct {
	Hash     string
	Metadata *Metadata
}

//Encode write byte trojan
func (s *Trojan) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, maxPacketSize))
	buf.Write([]byte(s.Hash))
	buf.Write(crlf)
	err := s.Metadata.WriteTo(buf)
	if err != nil {
		return nil, utils.BaseErr("trojan encode write fail", err)
	}
	buf.Write(crlf)
	return buf.Bytes(), nil
}

//Decode read byte trojan
func (s *Trojan) Decode(reader io.Reader) error {
	hash := [56]byte{}
	n, err := reader.Read(hash[:])
	if err != nil || n != 56 {
		return utils.BaseErr("failed to read hash", err)
	}
	crlf := [2]byte{}
	_, err = io.ReadFull(reader, crlf[:])
	if err != nil {
		return utils.BaseErr("trojan decode read buf", err)
	}

	s.Metadata = &Metadata{}
	if err := s.Metadata.ReadFrom(reader); err != nil {
		return utils.BaseErr("trojan decode read buf", err)
	}

	_, err = io.ReadFull(reader, crlf[:])
	if err != nil {
		return utils.BaseErr("trojan decode read buf", err)
	}
	return nil
}

//DialTrojan dial trojan remote
func DialTrojan(metadata *Metadata) (outbound net.Conn, err error) {
	config := client.GetConfig()
	if config.SSLEnable {
		outbound, err = tls.Dial("tcp", config.RemoteAddr, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		outbound, err = net.Dial("tcp", config.RemoteAddr)
	}

	if err != nil {
		return nil, err
	}

	socks := &Trojan{
		Hash: config.GetHash(),
		Metadata: &Metadata{
			socksCommand: connect,
			address:      metadata.address,
		},
	}
	encode, err := socks.Encode()
	if err != nil {
		return nil, err
	}

	_, err = outbound.Write(encode)
	if err != nil {
		return nil, err
	}
	return outbound, err
}

//PeekTrojan peek trojan protocol
func PeekTrojan(bufr *bufio.Reader, conn net.Conn) (bool, error) {
	hash, err := bufr.Peek(56)
	if err != nil {
		return false, utils.BaseErr("peek bytes fail", err)
	}
	if pw := server.Passwords[string(hash)]; pw != nil {
		log.Infof("trojan %s authenticated as %s", conn.RemoteAddr(), pw.Raw)
		trojan := Trojan{}
		pr := &peekReader{r: bufr}
		err := trojan.Decode(pr)
		if err != nil {
			log.Warnf("trojan proto decode fail %v", err)
			return false, nil
		}

		_, err = bufr.Discard(pr.i)
		if err != nil {
			log.Warnf("Discard trojan proto fail %v", err)
			return false, nil
		}

		outbound, err := net.Dial("tcp", trojan.Metadata.address.String())
		if err != nil {
			return false, utils.BaseErrf("trojan dial addr fail %v", err, trojan.Metadata.address.String())
		}
		log.Infof("trojan %s requested connection to %s", conn.RemoteAddr(), trojan.Metadata.String())

		defer outbound.Close()
		return true, utils.ProxyStreamBuf(bufr, conn, outbound, outbound)
	}
	return false, nil
}
