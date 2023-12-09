package gnet

import (
	"encoding/binary"
	"net"
)

const numOfLengthBytes = 8

func Serve(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		// 异步处理
		go func() {
			if errConn := handleConn(conn); errConn != nil {
				conn.Close()
			}
		}()
	}
}
func handleConn(conn net.Conn) error {
	for {
		bs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(bs)
		// err == net.ErrClosed || err == io.EOF || err == io.ErrUnexpectedEOF
		if err != nil {
			return err
		}
		res := handleMsg(bs)
		_, err = conn.Write(res)
		// err == net.ErrClosed || err == io.EOF || err == io.ErrUnexpectedEOF
		if err != nil {
			return err
		}
	}
}
func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}

type Server struct {
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if errConn := s.handleConn(conn); errConn != nil {
				conn.Close()
			}
		}()
	}
}

// 请求组成:
// part1. 长度字段，用固定字节表示
// part2. 请求数据
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		// lenBs 长度字段的字节表示
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}
		// 获取消息长度
		length := binary.BigEndian.Uint64(lenBs)

		reqBs := make([]byte, length)
		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}

		respData := handleMsg(reqBs)
		respLen := len(respData)

		// 构建响应数据
		res := make([]byte, respLen+numOfLengthBytes)
		// step1.写入长度
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
		// step2.写入数据
		copy(res[numOfLengthBytes:], respData)

		if _, err = conn.Write(res); err != nil {
			return err
		}
	}
	return nil
}
