// Package network provides the low-level TCP framing layer for R2 Online.
//
// Wire format of every packet:
//
//	[SIZE: uint16 LE]  — total bytes from this field onward (includes itself)
//	[ENC:  uint8]      — 0x00 = plaintext, 0x01 = RC4-encrypted from here
//	[SEQ:  uint8]      — packet sequence number
//	[OP:   uint16 LE]  — opcode
//	[DATA: bytes...]   — payload
//
// SIZE = 2 (ENC+SEQ) + 2 (OP) + len(DATA) + 2 (the SIZE field itself) = 6 + len(DATA)
package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"r2server/internal/crypto"
)

const (
	encPlain     = 0x00
	encEncrypted = 0x01
	defaultSeq   = 0x01

	// maxPacketSize guards against malformed/malicious oversized packets.
	maxPacketSize = 64 * 1024
)

// Conn is a single TCP connection with R2 Online packet framing and optional RC4 encryption.
// It is safe to call Send from multiple goroutines; reads must be called from a single goroutine.
type Conn struct {
	conn net.Conn

	mu         sync.Mutex
	sendCipher *crypto.RC4 // nil until SetCipher is called

	recvCipher *crypto.RC4 // nil until SetCipher is called
}

func NewConn(c net.Conn) *Conn {
	return &Conn{conn: c}
}

// SetCipher installs the RC4 cipher for both directions.
// Call this after sending ConnectionClient (1103) and receiving the client's first encrypted packet.
func (c *Conn) SetCipher(key [256]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sendCipher = crypto.NewRC4(key)
	c.recvCipher = crypto.NewRC4(key)
}

// RemoteAddr returns the remote address of the connection.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close closes the underlying TCP connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// Send serialises and writes a packet to the wire.
// If a cipher has been set, the [ENC+SEQ+OP+DATA] portion is encrypted.
func (c *Conn) Send(opcode uint16, payload []byte) error {
	// Build inner: [ENC][SEQ][OP_LO][OP_HI][DATA...]
	inner := make([]byte, 4+len(payload))
	inner[1] = defaultSeq
	binary.LittleEndian.PutUint16(inner[2:], opcode)
	copy(inner[4:], payload)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sendCipher != nil {
		inner[0] = encEncrypted
		c.sendCipher.Crypt(inner[1:]) // encrypt SEQ + OP + DATA (not the flag itself)
	} else {
		inner[0] = encPlain
	}

	// Prepend SIZE = len(inner) + 2 (the size field itself)
	size := uint16(len(inner) + 2)
	frame := make([]byte, 2+len(inner))
	binary.LittleEndian.PutUint16(frame[:2], size)
	copy(frame[2:], inner)

	_, err := c.conn.Write(frame)
	return err
}

// Recv reads one packet from the wire and returns (opcode, payload, error).
// Blocks until a complete packet is available or an error occurs.
func (c *Conn) Recv() (opcode uint16, payload []byte, err error) {
	// Read SIZE (2 bytes)
	var sizeBuf [2]byte
	if _, err = io.ReadFull(c.conn, sizeBuf[:]); err != nil {
		return 0, nil, fmt.Errorf("conn: read size: %w", err)
	}
	size := binary.LittleEndian.Uint16(sizeBuf[:])
	if size < 6 {
		return 0, nil, fmt.Errorf("conn: packet too small: size=%d", size)
	}
	if int(size) > maxPacketSize {
		return 0, nil, fmt.Errorf("conn: packet too large: size=%d", size)
	}

	// Read the rest: [ENC][SEQ][OP][DATA...]
	inner := make([]byte, size-2)
	if _, err = io.ReadFull(c.conn, inner); err != nil {
		return 0, nil, fmt.Errorf("conn: read body: %w", err)
	}

	encFlag := inner[0]
	body := inner[1:] // [SEQ][OP_LO][OP_HI][DATA...]

	if encFlag == encEncrypted {
		if c.recvCipher == nil {
			return 0, nil, fmt.Errorf("conn: encrypted packet but cipher not set")
		}
		c.recvCipher.Crypt(body) // decrypt in-place
	}

	// body: [SEQ: 1][OP: 2][DATA: N]
	if len(body) < 3 {
		return 0, nil, fmt.Errorf("conn: body too short after flag: %d bytes", len(body))
	}
	// seq := body[0]  // available if needed for anti-replay checks
	opcode = binary.LittleEndian.Uint16(body[1:3])
	payload = body[3:]
	return opcode, payload, nil
}
