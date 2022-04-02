package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad(t *testing.T) {
	load, err := LoadServer("../../../example/shuttles.yaml")
	if err != nil {
		t.FailNow()
		return
	}
	assert.Equal(t, "sQtfRnfhcNoZYZh1wY9u", load.Trojan.Passwords[0])
}
