package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"strconv"
)

// socks5 const
const (
	Socks5Version = byte(0x05)
	NoAuth        = byte(0x00)
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
	if version[0] != Socks5Version {
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

	useMethod := NoAuth //默认不需要密码

	resp := []byte{Socks5Version, useMethod}
	s.Conn.Write(resp)
	return nil
}

//LSTRequest lst
func (s *Socks5) LSTRequest() (err error) {
	conn := s.Conn
	// Decode the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(conn, header, 3); err != nil {
		return utils.BaseErr("socks5 LSTRequest read header fail", err)
	}
	// Ensure we are compatible
	if header[0] != Socks5Version {
		return utils.NewErrf("unsupported SOCKS version: %v", string(header[:1]))
	}

	address := new(Address)
	err = address.ReadFrom(conn)
	if err != nil {
		return utils.BaseErr("socks5 LSTRequest fail", err)
	}

	s.Metadata = &Metadata{
		socksCommand: socksCommand(header[1]),
		Address:      address,
	}
	return nil
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
	// Format the Address
	addrType := ipv4Address
	addrBody := []byte{0, 0, 0, 0}
	addrPort := 0

	// Format the message
	msg := make([]byte, 6+len(addrBody))
	msg[0] = Socks5Version
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
	*Address
}

//ReadFrom metadata read byte
func (r *Metadata) ReadFrom(rr io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(rr, byteBuf[:])
	if err != nil {
		return err
	}
	r.socksCommand = socksCommand(byteBuf[0])
	r.Address = new(Address)
	err = r.Address.ReadFrom(rr)
	if err != nil {
		return utils.BaseErr("failed to marshal Address", err)
	}
	return nil
}

//WriteTo metadata write byte
func (r *Metadata) WriteTo(w io.Writer) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteByte(byte(r.socksCommand))
	if err := r.Address.WriteTo(buf); err != nil {
		return err
	}
	// use tcp by default
	r.Address.NetworkType = "tcp"
	_, err := w.Write(buf.Bytes())
	return err
}

//Network network string
func (r *Metadata) Network() string {
	return r.Address.Network()
}

//String Address string
func (r *Metadata) String() string {
	return r.Address.String()
}

type AddressType byte

// trojan AddressType
const (
	IPv4       AddressType = 1
	DomainName AddressType = 3
	IPv6       AddressType = 4
)

type Address struct {
	DomainName  string
	Port        int
	NetworkType string
	net.IP
	AddressType
}

func (a *Address) String() string {
	switch a.AddressType {
	case IPv4:
		return fmt.Sprintf("%s:%d", a.IP.String(), a.Port)
	case IPv6:
		return fmt.Sprintf("[%s]:%d", a.IP.String(), a.Port)
	case DomainName:
		return fmt.Sprintf("%s:%d", a.DomainName, a.Port)
	default:
		return "INVALID_ADDRESS_TYPE"
	}
}

func (a *Address) Network() string {
	if a.NetworkType == "" {
		return "tcp"
	}
	return a.NetworkType
}

func (a *Address) ResolveIP() (net.IP, error) {
	if a.AddressType == IPv4 || a.AddressType == IPv6 {
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

func (a *Address) ReadFrom(r io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(r, byteBuf[:])
	if err != nil {
		return utils.BaseErr("unable to read ATYP", err)
	}
	a.AddressType = AddressType(byteBuf[0])
	switch a.AddressType {
	case IPv4:
		var buf [6]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return utils.BaseErr("failed to read IPv4", err)
		}
		a.IP = buf[0:4]
		a.Port = int(binary.BigEndian.Uint16(buf[4:6]))
	case IPv6:
		var buf [18]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return utils.BaseErr("failed to read IPv6", err)
		}
		a.IP = buf[0:16]
		a.Port = int(binary.BigEndian.Uint16(buf[16:18]))
	case DomainName:
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
				a.AddressType = IPv4
			} else {
				a.AddressType = IPv6
			}
		} else {
			a.DomainName = string(host)
		}
		a.Port = int(binary.BigEndian.Uint16(buf[length : length+2]))
	default:
		return utils.NewErr("invalid ATYP " + strconv.FormatInt(int64(a.AddressType), 10))
	}
	return nil
}

func (a *Address) WriteTo(w io.Writer) error {
	_, err := w.Write([]byte{byte(a.AddressType)})
	if err != nil {
		return err
	}
	switch a.AddressType {
	case DomainName:
		w.Write([]byte{byte(len(a.DomainName))})
		_, err = w.Write([]byte(a.DomainName))
	case IPv4:
		_, err = w.Write(a.IP.To4())
	case IPv6:
		_, err = w.Write(a.IP.To16())
	default:
		return utils.NewErr("invalid ATYP " + strconv.FormatInt(int64(a.AddressType), 10))
	}
	if err != nil {
		return err
	}
	port := [2]byte{}
	binary.BigEndian.PutUint16(port[:], uint16(a.Port))
	_, err = w.Write(port[:])
	return err
}

func NewAddressFromAddr(network string, addr string) (*Address, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		panic(err)
	}
	return NewAddressFromHostPort(network, host, int(port)), nil
}

func NewAddressFromHostPort(network string, host string, port int) *Address {
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return &Address{
				IP:          ip,
				Port:        port,
				AddressType: IPv4,
				NetworkType: network,
			}
		}
		return &Address{
			IP:          ip,
			Port:        port,
			AddressType: IPv6,
			NetworkType: network,
		}
	}
	return &Address{
		DomainName:  host,
		Port:        port,
		AddressType: DomainName,
		NetworkType: network,
	}
}
