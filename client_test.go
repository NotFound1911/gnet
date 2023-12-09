package gnet

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	go func() {
		err := Serve("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 5)
	err := Connect("tcp", "localhost:8082")
	t.Log(err)
}
func TestClient_Send(t *testing.T) {
	server := &Server{}
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	client := &Client{
		network: "tcp",
		addr:    "localhost:8081",
	}
	resp, err := client.Send("test")
	require.NoError(t, err)
	assert.Equal(t, "testtest", resp)
}
