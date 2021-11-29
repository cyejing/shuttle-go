package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad(t *testing.T) {
	load, err := Load("../../../example/shuttlec.yaml")
	if err != nil {
		t.FailNow()
		return
	}
	assert.Equal(t, "socks", load.RunType)
}
