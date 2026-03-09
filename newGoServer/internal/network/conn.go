// Package network provides the low-level TCP framing layer for R2 Online.
//
// Wire format:
//   [SIZE:uint16 LE][ENC:uint8][SEQ:uint8][OP:uint16 LE][DATA...]
//
// SIZE includes all bytes starting from SIZE itself.
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

	maxPacketSize = 64 * 1024
)

type diagSbox struct {
	name string
	sbox [256]byte
}

// Conn is one framed TCP connection.
type Conn struct {
	conn net.Conn
	log  *zap.Logger

	mu sync.Mutex

	recvSboxSet bool
	recvSbox    [256]byte
	recvDropN   int
	recvStream  bool
	recvCipher  *crypto.RC4
	recvInit    bool

	// Candidate S-boxes for diagnostics.
	diagSboxes []diagSbox

	// Optional drop-N diagnostic.
	diagDropTarget uint16
	diagDropMax    int
	diagDropSeq    int
}

func NewConn(c net.Conn, log *zap.Logger) *Conn {
	return &Conn{
		conn: c,
		log:  log.With(zap.String("remote", c.RemoteAddr().String())),
	}
}

// AddDiagSbox adds a candidate sbox for trial decryption logs.
func (c *Conn) AddDiagSbox(name string, sbox [256]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.diagSboxes = append(c.diagSboxes, diagSbox{name: name, sbox: sbox})
}

// SetDiagDropSearch enables search for "drop N keystream bytes before decrypt".
func (c *Conn) SetDiagDropSearch(targetOpcode uint16, maxDrop int, expectedSeq int) {
	if maxDrop < 0 {
		maxDrop = 0
	}
	if expectedSeq < -1 {
		expectedSeq = -1
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.diagDropTarget = targetOpcode
	c.diagDropMax = maxDrop
	c.diagDropSeq = expectedSeq
}

// SetRecvSbox installs the incoming decryption sbox.
func (c *Conn) SetRecvSbox(sbox [256]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recvSbox = sbox
	c.recvSboxSet = true
	c.recvCipher = nil
	c.recvInit = false
}

// SetRecvDropN sets how many RC4 keystream bytes to discard before decrypting
// each packet (reset mode) or once before first packet (stream mode).
func (c *Conn) SetRecvDropN(n int) {
	if n < 0 {
		n = 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recvDropN = n
	c.recvCipher = nil
	c.recvInit = false
}

// SetRecvStreamMode enables/disables continuous RC4 state across packets.
func (c *Conn) SetRecvStreamMode(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recvStream = enabled
	c.recvCipher = nil
	c.recvInit = false
}

func (c *Conn) RemoteAddr() net.Addr { return c.conn.RemoteAddr() }
func (c *Conn) Close() error         { return c.conn.Close() }

// Send writes one plaintext packet.
func (c *Conn) Send(opcode uint16, payload []byte) error {
	inner := make([]byte, 4+len(payload))
	inner[0] = encPlain
	inner[1] = defaultSeq
	binary.LittleEndian.PutUint16(inner[2:], opcode)
	copy(inner[4:], payload)

	size := uint16(len(inner) + 2)
	frame := make([]byte, 2+len(inner))
	binary.LittleEndian.PutUint16(frame[:2], size)
	copy(frame[2:], inner)

	if ce := c.log.Check(zap.DebugLevel, "-> SEND"); ce != nil {
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

// Recv reads and decodes one packet.
func (c *Conn) Recv() (opcode uint16, payload []byte, err error) {
	var sizeBuf [2]byte
	if _, err = io.ReadFull(c.conn, sizeBuf[:]); err != nil {
		return 0, nil, fmt.Errorf("conn: read size: %w", err)
	}
	size := binary.LittleEndian.Uint16(sizeBuf[:])

	if ce := c.log.Check(zap.DebugLevel, "<- RECV size"); ce != nil {
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
		diags := c.diagSboxes
		dropTarget := c.diagDropTarget
		dropMax := c.diagDropMax
		dropSeq := c.diagDropSeq
		dropN := c.recvDropN
		stream := c.recvStream
		c.mu.Unlock()

		if !sboxSet {
			return 0, nil, fmt.Errorf("conn: encrypted packet but recv sbox not set")
		}

		if len(diags) > 0 {
			rawPreview := body
			if len(rawPreview) > 16 {
				rawPreview = rawPreview[:16]
			}
			fields := []zap.Field{
				zap.String("raw_hex", hex.EncodeToString(rawPreview)),
			}
			for _, d := range diags {
				buf := make([]byte, len(body))
				copy(buf, body)
				crypto.NewRC4(d.sbox).Crypt(buf)
				trialOp := uint16(0)
				if len(buf) >= 3 {
					trialOp = binary.LittleEndian.Uint16(buf[1:3])
				}
				fields = append(fields, zap.String("op_if_"+d.name, fmt.Sprintf("0x%04X(%d)", trialOp, trialOp)))
				if dropMax > 0 {
					if n, ok := findDropForOpcode(d.sbox, body, dropTarget, dropMax, dropSeq); ok {
						fields = append(fields, zap.Int("drop_if_"+d.name, n))
					}
				}
			}
			c.log.Info("<- ENCRYPTED packet - sbox trial", fields...)
		}

		var cipher *crypto.RC4
		if stream {
			if !c.recvInit || c.recvCipher == nil {
				c.recvCipher = crypto.NewRC4(sbox)
				if dropN > 0 {
					drop := make([]byte, dropN)
					c.recvCipher.Crypt(drop)
				}
				c.recvInit = true
			}
			cipher = c.recvCipher
		} else {
			cipher = crypto.NewRC4(sbox)
			if dropN > 0 {
				drop := make([]byte, dropN)
				cipher.Crypt(drop)
			}
		}
		cipher.Crypt(body)
	}

	if len(body) < 3 {
		return 0, nil, fmt.Errorf("conn: body too short: %d bytes", len(body))
	}
	opcode = binary.LittleEndian.Uint16(body[1:3])
	payload = body[3:]

	if ce := c.log.Check(zap.DebugLevel, "<- RECV"); ce != nil {
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

func findDropForOpcode(sbox [256]byte, cipherBody []byte, target uint16, maxDrop int, expectedSeq int) (int, bool) {
	if len(cipherBody) < 3 || maxDrop < 0 {
		return 0, false
	}
	ks := make([]byte, maxDrop+3)
	crypto.NewRC4(sbox).Crypt(ks) // RC4 keystream from zero input

	for drop := 0; drop <= maxDrop; drop++ {
		if expectedSeq >= 0 {
			seq := cipherBody[0] ^ ks[drop]
			if int(seq) != expectedSeq {
				continue
			}
		}
		op := uint16(cipherBody[1]^ks[drop+1]) | (uint16(cipherBody[2]^ks[drop+2]) << 8)
		if op == target {
			return drop, true
		}
	}
	return 0, false
}
