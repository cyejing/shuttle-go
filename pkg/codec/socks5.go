package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cyejing/shuttle/pkg/log"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"strconv"
)

const (
	socks5Version = byte(0x05)
	NoAuth        = byte(0x00)
)

type Socks5 struct {
	Conn     net.Conn
	Metadata *Metadata
}

func (s *Socks5) HandleHandshake() error {
	bufConn := bufio.NewReader(s.Conn)
	version := []byte{0}
	if _, err := bufConn.Read(version); err != nil {
		log.Errorf("[ERR] socks: Failed to get version byte: %v", err)
		return err
	}
	// Ensure we are compatible
	if version[0] != socks5Version {
		err := fmt.Errorf("Unsupported SOCKS version: %v", version)
		log.Errorf("[ERR] socks: %v", err)
		return err
	}
	header := []byte{0}
	if _, err := bufConn.Read(header); err != nil {
		return err
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(bufConn, methods, numMethods)
	if err != nil {
		return err
	}

	useMethod := NoAuth //默认不需要密码

	resp := []byte{socks5Version, useMethod}
	s.Conn.Write(resp)
	return nil
}

func (s *Socks5) LSTRequest() (err error) {
	conn := s.Conn
	// Read the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(conn, header, 3); err != nil {
		return fmt.Errorf("Failed to get command version: %v", err)
	}
	// Ensure we are compatible
	if header[0] != socks5Version {
		return fmt.Errorf("Unsupported command version: %v", header[0])
	}

	address := new(Address)
	err = address.ReadFrom(conn)
	if err != nil {
		return err
	}

	s.Metadata = &Metadata{
		Command: Command(header[1]),
		Address: address,
	}
	return nil
}

func (s *Socks5) DialRemote(network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	socks := &Socks{
		Hash: utils.SHA224String("cyejing123"),
		Metadata: &Metadata{
			Command: Connect,
			Address: NewAddressFromHostPort("tcp", "127.0.0.1", 8088),
		},
	}
	encode, err := socks.Encode()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString("POST / HTTP/1.1")
	buf.Write(crlf)
	buf.WriteString("Host: localhost:4842")
	buf.Write(crlf)
	buf.WriteString("Content-Length: " + strconv.Itoa(len(encode)))
	buf.Write(crlf)
	buf.Write(crlf)
	buf.Write(encode)
	buf.Write(crlf)

	buf.WriteTo(conn)
	return conn, nil
}

const (
	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

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

// sendReply is used to send a reply message
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
