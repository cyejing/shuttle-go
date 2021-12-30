package codec

import (
	"testing"
)

func TestFmt(t *testing.T) {
	b := []byte{0x01, 0x02, 0x03}
	t.Logf("read: %X", b)
	t.Logf("read: %x", b)
	t.Logf("read: %v", b)
}
