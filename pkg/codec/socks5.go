package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	config "github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"strconv"
)

// socks5 const
const (
	socks5Version = byte(0x05)
	noAuth        = byte(0x00)
)

//Socks5 struct
type Socks5 struct {
	Conn     net.Conn
	Metadata *Metadata
}

//HandleHandshake handshake
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

//LSTRequest lst
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
		socksCommand: socksCommand(header[1]),
		address:      address,
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
			socksCommand: connect,
			address:      remoteAddr,
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

// socks metadata struct

type socksCommand byte

const (
	connect   socksCommand = 1
	associate socksCommand = 3
	mux       socksCommand = 0x7f
)

//Metadata struct
type Metadata struct {
	socksCommand
	*address
}

//ReadFrom metadata read byte
func (r *Metadata) ReadFrom(rr io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(rr, byteBuf[:])
	if err != nil {
		return err
	}
	r.socksCommand = socksCommand(byteBuf[0])
	r.address = new(address)
	err = r.address.ReadFrom(rr)
	if err != nil {
		return utils.BaseErr("failed to marshal address", err)
	}
	return nil
}

//WriteTo metadata write byte
func (r *Metadata) WriteTo(w io.Writer) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteByte(byte(r.socksCommand))
	if err := r.address.WriteTo(buf); err != nil {
		return err
	}
	// use tcp by default
	r.address.NetworkType = "tcp"
	_, err := w.Write(buf.Bytes())
	return err
}

//Network network string
func (r *Metadata) Network() string {
	return r.address.Network()
}

//String address string
func (r *Metadata) String() string {
	return r.address.String()
}

type addressType byte

// trojan AddressType
const (
	iPv4       addressType = 1
	domainName addressType = 3
	iPv6       addressType = 4
)

type address struct {
	DomainName  string
	Port        int
	NetworkType string
	net.IP
	addressType
}

func (a *address) String() string {
	switch a.addressType {
	case iPv4:
		return fmt.Sprintf("%s:%d", a.IP.String(), a.Port)
	case iPv6:
		return fmt.Sprintf("[%s]:%d", a.IP.String(), a.Port)
	case domainName:
		return fmt.Sprintf("%s:%d", a.DomainName, a.Port)
	default:
		return "INVALID_ADDRESS_TYPE"
	}
}

func (a *address) Network() string {
	return a.NetworkType
}

func (a *address) ResolveIP() (net.IP, error) {
	if a.addressType == iPv4 || a.addressType == iPv6 {
		return a.IP, nil
	}
	if a.IP != nil {
		return a.IP, nil
	}
	addr, err := net.ResolveIPAddr("ip", a.DomainName)
	if err != nil {
		return nil, utils.BaseErr("resolve ip fail", err)
	}
	a.IP = addr.IP
	return addr.IP, nil
}

func (a *address) ReadFrom(r io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(r, byteBuf[:])
	if err != nil {
		return utils.BaseErr("unable to read ATYP", err)
	}
	a.addressType = addressType(byteBuf[0])
	switch a.addressType {
	case iPv4:
		var buf [6]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return utils.BaseErr("failed to read iPv4", err)
		}
		a.IP = buf[0:4]
		a.Port = int(binary.BigEndian.Uint16(buf[4:6]))
	case iPv6:
		var buf [18]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return utils.BaseErr("failed to read iPv6", err)
		}
		a.IP = buf[0:16]
		a.Port = int(binary.BigEndian.Uint16(buf[16:18]))
	case domainName:
		_, err := io.ReadFull(r, byteBuf[:])
		length := byteBuf[0]
		if err != nil {
			return utils.NewErr("failed to read domain name length")
		}
		buf := make([]byte, length+2)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return utils.NewErr("failed to read domain name")
		}
		// the fucking browser uses IP as a domain name sometimes
		host := buf[0:length]
		if ip := net.ParseIP(string(host)); ip != nil {
			a.IP = ip
			if ip.To4() != nil {
				a.addressType = iPv4
			} else {
				a.addressType = iPv6
			}
		} else {
			a.DomainName = string(host)
		}
		a.Port = int(binary.BigEndian.Uint16(buf[length : length+2]))
	default:
		return utils.NewErr("invalid ATYP " + strconv.FormatInt(int64(a.addressType), 10))
	}
	return nil
}

func (a *address) WriteTo(w io.Writer) error {
	_, err := w.Write([]byte{byte(a.addressType)})
	if err != nil {
		return err
	}
	switch a.addressType {
	case domainName:
		w.Write([]byte{byte(len(a.DomainName))})
		_, err = w.Write([]byte(a.DomainName))
	case iPv4:
		_, err = w.Write(a.IP.To4())
	case iPv6:
		_, err = w.Write(a.IP.To16())
	default:
		return utils.NewErr("invalid ATYP " + strconv.FormatInt(int64(a.addressType), 10))
	}
	if err != nil {
		return err
	}
	port := [2]byte{}
	binary.BigEndian.PutUint16(port[:], uint16(a.Port))
	_, err = w.Write(port[:])
	return err
}

func newAddressFromAddr(network string, addr string) (*address, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		panic(err)
	}
	return newAddressFromHostPort(network, host, int(port)), nil
}

func newAddressFromHostPort(network string, host string, port int) *address {
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return &address{
				IP:          ip,
				Port:        port,
				addressType: iPv4,
				NetworkType: network,
			}
		}
		return &address{
			IP:          ip,
			Port:        port,
			addressType: iPv6,
			NetworkType: network,
		}
	}
	return &address{
		DomainName:  host,
		Port:        port,
		addressType: domainName,
		NetworkType: network,
	}
}
