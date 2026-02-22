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
//
// Encryption model (discovered from reference implementation):
//   - Server → Client: always PLAINTEXT (enc_flag = 0x00).
//   - Client → Server: RC4-encrypted (enc_flag = 0x01) using a hardcoded S-box
//     (crypto.DecryptSbox). The cipher is RESET to the initial S-box state for
//     every incoming packet — state does NOT carry over between packets.
package network

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"

	"r2server/internal/crypto"
)

const (
	encPlain     = 0x00
	encEncrypted = 0x01
	defaultSeq   = 0x01

	// maxPacketSize guards against malformed/malicious oversized packets.
	maxPacketSize = 64 * 1024
)

// Conn is a single TCP connection with R2 Online packet framing.
// Sends are always plaintext. Receives are decrypted per-packet when enc_flag = 0x01.
// It is safe to call Send from multiple goroutines; reads must be from a single goroutine.
type Conn struct {
	conn net.Conn
	log  *zap.Logger

	mu sync.Mutex

	// recvSboxSet is true once SetRecvSbox has been called.
	// When set, incoming packets with enc_flag=0x01 are decrypted using a fresh
	// RC4 instance seeded from recvSbox for each packet.
	recvSboxSet bool
	recvSbox    [256]byte
}

func NewConn(c net.Conn, log *zap.Logger) *Conn {
	return &Conn{
		conn: c,
		log:  log.With(zap.String("remote", c.RemoteAddr().String())),
	}
}

// SetRecvSbox installs the S-box used to decrypt incoming (client→server) packets.
// The cipher is reset to this initial S-box state for every received packet.
// Server→client packets are always sent as plaintext.
func (c *Conn) SetRecvSbox(sbox [256]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recvSbox = sbox
	c.recvSboxSet = true
}

// RemoteAddr returns the remote address of the connection.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close closes the underlying TCP connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// Send serialises and writes a packet to the wire as plaintext (enc_flag = 0x00).
func (c *Conn) Send(opcode uint16, payload []byte) error {
	// Build frame: [SIZE][ENC=0x00][SEQ][OP_LO][OP_HI][DATA...]
	inner := make([]byte, 4+len(payload))
	inner[0] = encPlain
	inner[1] = defaultSeq
	binary.LittleEndian.PutUint16(inner[2:], opcode)
	copy(inner[4:], payload)

	size := uint16(len(inner) + 2)
	frame := make([]byte, 2+len(inner))
	binary.LittleEndian.PutUint16(frame[:2], size)
	copy(frame[2:], inner)

	if ce := c.log.Check(zap.DebugLevel, "→ SEND"); ce != nil {
		preview := payload
		if len(preview) > 64 {
			preview = preview[:64]
		}
		ce.Write(
			zap.String("op", fmt.Sprintf("0x%04X(%d)", opcode, opcode)),
			zap.String("enc", "plain"),
			zap.Int("payload_len", len(payload)),
			zap.String("payload_hex", hex.EncodeToString(preview)),
		)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.conn.Write(frame)
	return err
}

// Recv reads one packet from the wire and returns (opcode, payload, error).
// Blocks until a complete packet is available or an error occurs.
// If enc_flag == 0x01 and SetRecvSbox has been called, the body is decrypted
// using a fresh RC4 instance (per-packet reset).
func (c *Conn) Recv() (opcode uint16, payload []byte, err error) {
	// Read SIZE (2 bytes) — always plaintext
	var sizeBuf [2]byte
	if _, err = io.ReadFull(c.conn, sizeBuf[:]); err != nil {
		return 0, nil, fmt.Errorf("conn: read size: %w", err)
	}
	size := binary.LittleEndian.Uint16(sizeBuf[:])

	if ce := c.log.Check(zap.DebugLevel, "← RECV size"); ce != nil {
		ce.Write(
			zap.String("raw", hex.EncodeToString(sizeBuf[:])),
			zap.Uint16("size", size),
		)
	}

	if size < 6 {
		return 0, nil, fmt.Errorf("conn: packet too small: size=%d (raw: %s)", size, hex.EncodeToString(sizeBuf[:]))
	}
	if int(size) > maxPacketSize {
		return 0, nil, fmt.Errorf("conn: packet too large: size=%d", size)
	}

	// Read the rest: [ENC][SEQ][OP_LO][OP_HI][DATA...]
	inner := make([]byte, size-2)
	if _, err = io.ReadFull(c.conn, inner); err != nil {
		return 0, nil, fmt.Errorf("conn: read body: %w", err)
	}

	encFlag := inner[0]
	seq := inner[1]
	body := inner[1:] // [SEQ][OP_LO][OP_HI][DATA...]

	if encFlag == encEncrypted {
		c.mu.Lock()
		sboxSet := c.recvSboxSet
		sbox := c.recvSbox
		c.mu.Unlock()

		if !sboxSet {
			return 0, nil, fmt.Errorf("conn: encrypted packet but recv sbox not set")
		}
		// Fresh RC4 instance for this packet — state does not persist between packets.
		cipher := crypto.NewRC4(sbox)
		cipher.Crypt(body)
	}

	// body: [SEQ: 1][OP: 2][DATA: N]
	if len(body) < 3 {
		return 0, nil, fmt.Errorf("conn: body too short: %d bytes", len(body))
	}
	opcode = binary.LittleEndian.Uint16(body[1:3])
	payload = body[3:]

	if ce := c.log.Check(zap.DebugLevel, "← RECV"); ce != nil {
		encStr := "plain"
		if encFlag == encEncrypted {
			encStr = "rc4(fresh)"
		}
		preview := payload
		if len(preview) > 64 {
			preview = preview[:64]
		}
		ce.Write(
			zap.String("op", fmt.Sprintf("0x%04X(%d)", opcode, opcode)),
			zap.String("enc", encStr),
			zap.Uint8("seq", seq),
			zap.Int("payload_len", len(payload)),
			zap.String("payload_hex", hex.EncodeToString(preview)),
		)
	}

	return opcode, payload, nil
}
