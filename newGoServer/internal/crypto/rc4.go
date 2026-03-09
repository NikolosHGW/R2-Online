// Package crypto implements the stream cipher used by R2 Online.
//
// The server (FnlApiW.dll, CRc4A::SetKey) uses standard RC4 KSA to derive a
// 256-element S-box from an 18-byte key (__mKey8 from CTrCryptKeyAck / opcode 1103).
// The stream cipher itself is also standard RC4, but reset per-packet (fresh S-box
// state for every received packet, i=0 j=0).
//
// Per-session encryption flow:
//  1. Server generates a random 18-byte key (GenerateSessionKey).
//  2. Server derives the S-box via KSA(key) — identical to what CRc4A::SetKey does.
//  3. Server sends the key embedded in the __mKey8 field of ConnectionClient (1103).
//  4. Client runs the same KSA on the received __mKey8 and gets the same S-box.
//  5. All subsequent client→server packets are encrypted with that S-box.
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

// KSA runs the standard RC4 Key Scheduling Algorithm and returns the resulting
// 256-byte S-box. This matches CRc4A::SetKey in FnlApiW.dll exactly.
func KSA(key []byte) [256]byte {
	var s [256]byte
	for i := range s {
		s[i] = byte(i)
	}
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(key[i%len(key)])) & 0xff
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// GenerateSessionKey generates a cryptographically random 18-byte key (9 int16 values)
// for embedding in the __mKey8 field of ConnectionClient (opcode 1103).
// Derive the session S-box with KSA(key[:]).
func GenerateSessionKey() ([18]byte, error) {
	var key [18]byte
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
