package codec

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/cyejing/shuttle/pkg/config/client"
	"github.com/cyejing/shuttle/pkg/config/server"
	"github.com/cyejing/shuttle/pkg/utils"
	"io"
	"net"
	"strconv"
)

var crlf = []byte{0x0d, 0x0a}

//Trojan struct
type Trojan struct {
	Hash     string
	Metadata *Metadata
}

func exitHash(hash []byte) (*server.Password, bool) {
	pw := server.Passwords[string(hash)]
	return pw, pw != nil
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
		Hash: utils.SHA224String(config.Password),
		Metadata: &Metadata{
			command: connect,
			address: metadata.address,
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
func PeekTrojan(bufr *bufio.Reader, conn net.Conn) error {
	peek, err := bufr.Peek(56)
	if err != nil {
		return utils.BaseErr("peek bytes fail", err)
	}
	if pw, ok := exitHash(peek); ok {
		log.Infof("%s authenticated as %s", conn.RemoteAddr(), pw.Raw)
		trojan := Trojan{}
		pr := &peekReader{r: bufr}
		err := trojan.Decode(pr)
		if err != nil {
			log.Warnf("trojan proto decode fail %v", err)
			return nil
		}

		_, err = bufr.Discard(pr.i)
		if err != nil {
			log.Warnf("Discard trojan proto fail %v", err)
			return nil
		}

		outbound, err := net.Dial("tcp", trojan.Metadata.address.String())
		if err != nil {
			return utils.BaseErrf("trojan dial addr fail %v", err, trojan.Metadata.address.String())
		}
		log.Infof("trojan %s requested connection to %s", conn.RemoteAddr(), trojan.Metadata.String())

		defer outbound.Close()
		return utils.ProxyStreamBuf(bufr, conn, outbound, outbound)
	}
	return nil
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

type command byte

const (
	connect   command = 1
	associate command = 3
	mux       command = 0x7f
)

//Metadata struct
type Metadata struct {
	command
	*address
}

//ReadFrom metadata read byte
func (r *Metadata) ReadFrom(rr io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(rr, byteBuf[:])
	if err != nil {
		return err
	}
	r.command = command(byteBuf[0])
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
	buf.WriteByte(byte(r.command))
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
