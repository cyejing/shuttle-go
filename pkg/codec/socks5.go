package codec

import (
	"bufio"
	config "github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
)

// socks5 const
const (
	socks5Version = byte(0x05)
	noAuth        = byte(0x00)
)

type Socks5 struct {
	Conn     net.Conn
	Metadata *Metadata
}

func (s *Socks5) HandleHandshake() error {
	bufConn := bufio.NewReader(s.Conn)
	version := []byte{0}
	if _, err := bufConn.Read(version); err != nil {
		return utils.BaseErr("Failed to get version byte", err)
	}
	// Ensure we are compatible
	if version[0] != socks5Version {
		return utils.NewErrf("unsupported SOCKS version: %v", string(version))
	}
	header := []byte{0}
	if _, err := bufConn.Read(header); err != nil {
		return utils.BaseErr("socks5 handshake fail", err)
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(bufConn, methods, numMethods)
	if err != nil {
		return utils.BaseErr("socks5 handshake fail", err)
	}

	useMethod := noAuth //默认不需要密码

	resp := []byte{socks5Version, useMethod}
	s.Conn.Write(resp)
	return nil
}

func (s *Socks5) LSTRequest() (err error) {
	conn := s.Conn
	// Read the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(conn, header, 3); err != nil {
		return utils.BaseErr("socks5 LSTRequest read header fail", err)
	}
	// Ensure we are compatible
	if header[0] != socks5Version {
		return utils.NewErrf("unsupported SOCKS version: %v", string(header[:1]))
	}

	address := new(address)
	err = address.ReadFrom(conn)
	if err != nil {
		return utils.BaseErr("socks5 LSTRequest fail", err)
	}

	s.Metadata = &Metadata{
		command: command(header[1]),
		address: address,
	}
	return nil
}

//DialSendTrojan send trojan protocol to remote
func (s *Socks5) DialSendTrojan(network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, utils.BaseErrf("socks5 dial remote fai %sl", err, addr)
	}
	c := config.GetConfig()

	remoteAddr, err := newAddressFromAddr("tcp", c.RemoteAddr)
	if err != nil {
		return nil, utils.BaseErrf("socks5 send trojan fail %v", err, c.RemoteAddr)
	}
	trojan := &Trojan{
		Hash: utils.SHA224String(c.Password),
		Metadata: &Metadata{
			command: connect,
			address: remoteAddr,
		},
	}
	encode, err := trojan.Encode()
	if err != nil {
		return nil, utils.BaseErrf("socks5 encode trojan fail %v", err, c.RemoteAddr)
	}

	conn.Write(encode)
	return conn, nil
}

const (
	connectCommand   = uint8(1)
	bindCommand      = uint8(2)
	associateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

//send reply byte
const (
	SuccessReply uint8 = iota
	serverFailure
	ruleFailure
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

//SendReply send reply byte
func (s *Socks5) SendReply(resp uint8) error {
	// Format the address
	addrType := ipv4Address
	addrBody := []byte{0, 0, 0, 0}
	addrPort := 0

	// Format the message
	msg := make([]byte, 6+len(addrBody))
	msg[0] = socks5Version
	msg[1] = resp
	msg[2] = 0 // Reserved
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)

	// Send the message
	_, err := s.Conn.Write(msg)
	return err
}
