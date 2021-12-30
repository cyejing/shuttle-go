package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadSocks(t *testing.T) {
	load, err := Load("../../../example/shuttlec-socks.yaml")
	if err != nil {
		t.FailNow()
		return
	}
	assert.Equal(t, "sQtfRnfhcNoZYZh1wY9u", load.Password)
}

func TestLoadWormhole(t *testing.T) {
	load, err := Load("../../../example/shuttlec-wormhole.yaml")
	if err != nil {
		t.FailNow()
		return
	}
	assert.Equal(t, "58JCEmvcBkRAk1XkK1iH", load.Password)
}
