// Package crypto implements the custom RC4-like stream cipher used by R2 Online.
//
// The algorithm is identical to standard RC4 but without the Key Scheduling
// Algorithm (KSA): the 256-byte S-box is used directly as provided, with the
// two state pointers (i, j) initialised to zero.
//
// This means the "key" sent to the client IS the S-box — the server generates
// 256 cryptographically random bytes, sends them in the ConnectionClient packet
// (opcode 1103), and both sides use those bytes as the initial RC4 S-box state.
package crypto

import "crypto/rand"

// RC4 holds the mutable cipher state for a single session.
// Each session must have its own RC4 instance; the state is not goroutine-safe.
type RC4 struct {
	s [256]byte
	i uint8 // pointerKey (called esi in the reversed binary)
	j uint8 // accumulator  (called edi in the reversed binary)
}

// NewRC4 creates a cipher seeded from the provided 256-byte S-box.
func NewRC4(sbox [256]byte) *RC4 {
	return &RC4{s: sbox}
}

// GenerateKey generates a cryptographically random 256-byte S-box suitable
// for seeding NewRC4 and for sending to the client in the ConnectionClient packet.
func GenerateKey() ([256]byte, error) {
	var key [256]byte
	_, err := rand.Read(key[:])
	return key, err
}

// Crypt applies the RC4 stream cipher to data in-place (encrypt == decrypt).
// The internal state is advanced; consecutive calls continue the stream.
func (c *RC4) Crypt(data []byte) {
	for k := range data {
		c.i++
		v7 := c.s[c.i]
		c.j += v7
		v8 := c.s[c.j]
		c.s[c.i] = v8
		c.s[c.j] = v7
		// uint8 addition wraps naturally in Go
		data[k] ^= c.s[uint8(v7+v8)]
	}
}
