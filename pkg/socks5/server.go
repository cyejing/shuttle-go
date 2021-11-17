package socks5

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const (
	socks5Version = uint8(5)
)

const (
	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

const (
	successReply uint8 = iota
	serverFailure
	ruleFailure
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

var (
	unrecognizedAddrType = fmt.Errorf("Unrecognized address type")
)

type Server struct {
	port        int
	authMethods map[uint8]Authenticator
}

type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

type Request struct {
	// Protocol version
	Version uint8
	// Requested command
	Command uint8
	// AuthContext provided during negotiation
	AuthContext *AuthContext
	// AddrSpec of the the network that sent the request
	RemoteAddr *AddrSpec
	// AddrSpec of the desired destination
	DestAddr *AddrSpec
	// AddrSpec of the actual destination (might be affected by rewrite)
	realDestAddr *AddrSpec
	bufConn      io.Reader
}

func New(port int) *Server {
	s := &Server{
		port: port,
	}
	s.authMethods = map[uint8]Authenticator{
		NoAuth: &NoAuthAuthenticator{},
	}
	return s
}

func (s *Server) Listen() {
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		panic(err)
	}

	for true {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Errorf("[ERR] socks5 conn err %v", err)
		}
		fmt.Printf("new conn %v \n", conn)

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	version := []byte{0}
	_, err := reader.Read(version)
	if err != nil {
		fmt.Printf("[ERR] socks: Failed to get version byte: %v", err)
		return
	}

	if version[0] != socks5Version {
		err := fmt.Errorf("unsupported SOCKS version: %v", version)
		fmt.Printf("[ERR] socks: %v", err)
		return
	}
	_, err = s.authenticate(conn, reader)

	request, err := NewRequest(reader)

	fmt.Printf("new request %v \n", request.DestAddr)
	s.handleRequest(request, conn)
}

func NewRequest(bufConn io.Reader) (*Request, error) {
	// Read the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(bufConn, header, 3); err != nil {
		return nil, fmt.Errorf("Failed to get command version: %v", err)
	}

	// Ensure we are compatible
	if header[0] != socks5Version {
		return nil, fmt.Errorf("Unsupported command version: %v", header[0])
	}

	// Read in the destination address
	dest, err := readAddrSpec(bufConn)
	if err != nil {
		return nil, err
	}

	request := &Request{
		Version:  socks5Version,
		Command:  header[1],
		DestAddr: dest,
		bufConn:  bufConn,
	}

	return request, nil
}

func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	d := &AddrSpec{}

	// Get the address type
	addrType := []byte{0}
	if _, err := r.Read(addrType); err != nil {
		return nil, err
	}

	// Handle on a per type basis
	switch addrType[0] {
	case ipv4Address:
		addr := make([]byte, 4)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case ipv6Address:
		addr := make([]byte, 16)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case fqdnAddress:
		if _, err := r.Read(addrType); err != nil {
			return nil, err
		}
		addrLen := int(addrType[0])
		fqdn := make([]byte, addrLen)
		if _, err := io.ReadAtLeast(r, fqdn, addrLen); err != nil {
			return nil, err
		}
		d.FQDN = string(fqdn)

	default:
		return nil, unrecognizedAddrType
	}

	// Read the port
	port := []byte{0, 0}
	if _, err := io.ReadAtLeast(r, port, 2); err != nil {
		return nil, err
	}
	d.Port = (int(port[0]) << 8) | int(port[1])

	return d, nil
}

// handleRequest is used for request processing after authentication
func (s *Server) handleRequest(req *Request, conn net.Conn) error {
	ctx := context.Background()

	// Resolve the address if we have a FQDN
	dest := req.DestAddr
	if dest.FQDN != "" {
		ipAddr, err := net.ResolveIPAddr("ip", dest.FQDN)
		if err != nil {
			return fmt.Errorf("[ERR] failed to resolve destination '%v': %v", dest.FQDN, err)
		}
		dest.IP = ipAddr.IP
	}

	req.realDestAddr = req.DestAddr

	// Switch on the command
	switch req.Command {
	case ConnectCommand:
		return s.handleConnect(ctx, conn, req)

	default:
		if err := sendReply(conn, commandNotSupported, nil); err != nil {
			return fmt.Errorf("Failed to send reply: %v", err)
		}
		return fmt.Errorf("Unsupported command: %v", req.Command)
	}
}

func (a *AddrSpec) String() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func (a AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

// handleConnect is used to handle a connect command
func (s *Server) handleConnect(ctx context.Context, conn net.Conn, req *Request) error {

	// Attempt to connect
	dial := func(ctx context.Context, net_, addr string) (net.Conn, error) {
		return net.Dial(net_, addr)
	}

	target, err := dial(ctx, "tcp", req.realDestAddr.Address())
	if err != nil {
		msg := err.Error()
		resp := hostUnreachable
		if strings.Contains(msg, "refused") {
			resp = connectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			resp = networkUnreachable
		}
		if err := sendReply(conn, resp, nil); err != nil {
			return fmt.Errorf("Failed to send reply: %v", err)
		}
		return fmt.Errorf("Connect to %v failed: %v", req.DestAddr, err)
	}
	defer target.Close()

	// Send success
	local := target.LocalAddr().(*net.TCPAddr)
	bind := AddrSpec{IP: local.IP, Port: local.Port}
	if err := sendReply(conn, successReply, &bind); err != nil {
		return fmt.Errorf("Failed to send reply: %v", err)
	}

	// Start proxying
	errCh := make(chan error, 2)
	go proxy(target, req.bufConn, errCh)
	go proxy(conn, target, errCh)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

// handleBind is used to handle a connect command
func (s *Server) handleBind(ctx context.Context, conn net.Conn, req *Request) error {
	if err := sendReply(conn, commandNotSupported, nil); err != nil {
		return fmt.Errorf("Failed to send reply: %v", err)
	}
	return nil
}

// handleAssociate is used to handle a connect command
func (s *Server) handleAssociate(ctx context.Context, conn net.Conn, req *Request) error {

	if err := sendReply(conn, commandNotSupported, nil); err != nil {
		return fmt.Errorf("Failed to send reply: %v", err)
	}
	return nil
}

// sendReply is used to send a reply message
func sendReply(w io.Writer, resp uint8, addr *AddrSpec) error {
	// Format the address
	var addrType uint8
	var addrBody []byte
	var addrPort uint16
	switch {
	case addr == nil:
		addrType = ipv4Address
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.FQDN != "":
		addrType = fqdnAddress
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = ipv4Address
		addrBody = []byte(addr.IP.To4())
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = ipv6Address
		addrBody = []byte(addr.IP.To16())
		addrPort = uint16(addr.Port)

	default:
		return fmt.Errorf("Failed to format address: %v", addr)
	}

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
	_, err := w.Write(msg)
	return err
}

type closeWriter interface {
	CloseWrite() error
}

// proxy is used to suffle data from src to destination, and sends errors
// down a dedicated channel
func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	//if tcpConn, ok := dst.(closeWriter); ok {
	//	fmt.Printf("close writer %v \n",dst)
	//	tcpConn.CloseWrite()
	//}
	errCh <- err
}
