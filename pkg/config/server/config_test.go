package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad(t *testing.T) {
	load, err := Load("../../../example/shuttles.yaml")
	if err != nil {
		t.FailNow()
		return
	}
	assert.Equal(t, "sQtfRnfhcNoZYZh1wY9u", load.Passwords[0])
}
