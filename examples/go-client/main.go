package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// buildPacket creates a packet with standard header.
func buildPacket(packetType uint16, payload []byte) []byte {
	buf := new(bytes.Buffer)
	// header
	buf.WriteByte(0x00) // no encryption
	buf.WriteByte(0x01) // packet counter
	binary.Write(buf, binary.LittleEndian, packetType)
	buf.Write(payload)

	size := uint16(buf.Len() + 2)
	final := new(bytes.Buffer)
	binary.Write(final, binary.LittleEndian, size)
	final.Write(buf.Bytes())
	return final.Bytes()
}

// buildAuthorizationLogin crafts packet 3100 using the server's parsing logic.
func buildAuthorizationLogin(login, password string) []byte {
	data := make([]byte, 512)

	// login offset encoding
	codeLogin := byte(0)
	data[256] = codeLogin * 8
	copy(data[151:], []byte(login))

	// password offset encoding
	codePass := byte(0)
	data[81] = codePass * 2
	copy(data[260:], []byte(password))

	return buildPacket(3100, data)
}

// readPacket reads one packet from connection.
func readPacket(conn net.Conn) (uint16, []byte, error) {
	sizeBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, sizeBuf); err != nil {
		return 0, nil, err
	}
	size := binary.LittleEndian.Uint16(sizeBuf)
	raw := make([]byte, size-2)
	if _, err := io.ReadFull(conn, raw); err != nil {
		return 0, nil, err
	}
	packetType := binary.LittleEndian.Uint16(raw[2:4])
	return packetType, raw[4:], nil
}

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:11015")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// receive welcome packet with decrypt key
	packetType, payload, err := readPacket(conn)
	if err != nil {
		panic(err)
	}
	fmt.Printf("recv packet %d, key length %d\n", packetType, len(payload))

	// send authorization login
	loginPkt := buildAuthorizationLogin("admin", "test")
	if _, err := conn.Write(loginPkt); err != nil {
		panic(err)
	}
	fmt.Println("sent login request")
}
