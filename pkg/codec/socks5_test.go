package codec

import (
	"bufio"
	"fmt"
	"testing"
)

func TestFmt(t *testing.T) {
	b := []byte{0x01, 0x02, 0x03}
	t.Logf("read: %X", b)
	t.Logf("read: %x", b)
	t.Logf("read: %v", b)
}

func TestDialRemote(t *testing.T) {
	s := &Socks5{}
	conn, err := s.DialRemote("tcp", "127.0.0.1:4842")
	if err != nil {
		return
	}
	scan := bufio.NewScanner(conn)
	for scan.Scan() {
		fmt.Println(scan.Text())
	}

}
