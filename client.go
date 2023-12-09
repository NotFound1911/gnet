package gnet

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second*5)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()
	for i := 0; i < 10; i++ {
		_, err = conn.Write([]byte("test"))
		if err != nil {
			return err
		}
		res := make([]byte, 128)
		_, err = conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println("res:", string(res))
	}
	return nil
}

type Client struct {
	network string
	addr    string
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second*5)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()
	reqLen := len(data)
	// 构建响应数据
	req := make([]byte, reqLen+numOfLengthBytes)
	// step1.写入长度
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	// step2.写入数据
	copy(req[numOfLengthBytes:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}
	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}
	// 响应长度
	length := binary.BigEndian.Uint64(lenBs)
	respBs := make([]byte, length)
	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}
	return string(respBs), nil
}
