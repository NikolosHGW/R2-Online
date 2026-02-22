package crypto

import (
	"crypto/rc4"
	"fmt"
	"testing"
)

// __mKey8 as raw bytes from WelcomeKey offset 96-113 (LE shorts in memory).
// CCryptKey.__mKey8 = short[9] = 9 × 2-byte LE values:
//   07c8  4e11  35a9  02a9  19e3  7eb1  59b3  53eb  76fe
var mKey8Raw = []byte{
	0xc8, 0x07, 0x11, 0x4e, 0xa9, 0x35, 0xa9, 0x02,
	0xe3, 0x19, 0xb1, 0x7e, 0xb3, 0x59, 0xeb, 0x53,
	0xfe, 0x76,
}

// standardRC4KSA runs the textbook RC4 KSA and returns the resulting 256-byte S-box.
func standardRC4KSA(key []byte) [256]byte {
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

// goRC4Keystream uses Go's stdlib rc4.NewCipher and returns the S-box after KSA
// by encrypting 256 zero bytes (the keystream IS the S-box output for zero input).
func goRC4Keystream256(key []byte) []byte {
	c, _ := rc4.NewCipher(key)
	buf := make([]byte, 256)
	c.XORKeyStream(buf, buf)
	return buf
}

func matchCount(a [256]byte, b [256]byte) int {
	n := 0
	for i := range a {
		if a[i] == b[i] {
			n++
		}
	}
	return n
}

func matchCountSlice(a [256]byte, b []byte) int {
	n := 0
	for i := range a {
		if i < len(b) && a[i] == b[i] {
			n++
		}
	}
	return n
}

// ksaWithShorts performs RC4 KSA using 9 uint16 key elements (from __mKey8).
// CRc4A::SState has int mM[256] — the cipher may use 16-bit j accumulator
// with short key values before masking to 255.
func ksaWithShorts(shorts [9]uint16) [256]byte {
	var s [256]byte
	for i := range s {
		s[i] = byte(i)
	}
	var j uint16
	for i := 0; i < 256; i++ {
		j = j + uint16(s[i]) + shorts[i%9]
		idx := j & 0xff
		s[i], s[idx] = s[idx], s[i]
	}
	return s
}

// ksaWithShortsHi16 does KSA where j stays as full uint16 but S-box index uses j%256.
func ksaWithShortsHi16(shorts [9]uint16) [256]byte {
	var s [256]byte
	for i := range s {
		s[i] = byte(i)
	}
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(shorts[i%9])) // no masking here!
		idx := j % 256
		if idx < 0 {
			idx += 256
		}
		s[i], s[idx] = s[idx], s[i]
	}
	return s
}

// TestKSAFromMKey8 tries every reasonable interpretation of __mKey8
// and reports how many bytes of the resulting S-box match DecryptSbox.
//
// A match of 256/256 means we found the derivation algorithm.
// A match of ~1-2/256 is chance.
func TestKSAFromMKey8(t *testing.T) {
	target := DecryptSbox

	type attempt struct {
		name string
		key  []byte
	}

	// Parse __mKey8 as 9 little-endian uint16 shorts
	var shorts [9]uint16
	for i := 0; i < 9; i++ {
		shorts[i] = uint16(mKey8Raw[i*2]) | uint16(mKey8Raw[i*2+1])<<8
	}

	// Shorts BE-swapped: swap each 2-byte pair
	beSwapped := make([]byte, 18)
	for i := 0; i < 9; i++ {
		beSwapped[i*2] = mKey8Raw[i*2+1]
		beSwapped[i*2+1] = mKey8Raw[i*2]
	}

	// Low bytes of each short (byte 0 of each pair)
	lowBytes := make([]byte, 9)
	for i := 0; i < 9; i++ {
		lowBytes[i] = mKey8Raw[i*2]
	}

	// High bytes of each short (byte 1 of each pair)
	highBytes := make([]byte, 9)
	for i := 0; i < 9; i++ {
		highBytes[i] = mKey8Raw[i*2+1]
	}

	// Maybe Windows CryptoAPI prepends the BLOBHEADER when doing KSA?
	blobHeader := []byte{0x08, 0x02, 0x00, 0x00, 0x01, 0x68, 0x00, 0x00, 0x12, 0x00, 0x00, 0x00}
	withBlobHeader := append(blobHeader, mKey8Raw...)

	attempts := []attempt{
		{"raw LE bytes (18)", mKey8Raw},
		{"BE-swapped shorts (18)", beSwapped},
		{"low bytes only (9)", lowBytes},
		{"high bytes only (9)", highBytes},
		{"with PLAINTEXTKEYBLOB header (30)", withBlobHeader},
	}

	for _, a := range attempts {
		sbox := standardRC4KSA(a.key)
		matches := matchCount(sbox, target)
		ks := goRC4Keystream256(a.key)
		ksMatches := matchCountSlice(target, ks)
		t.Logf("%-42s  KSA sbox: %3d/256  stdlib ks: %3d/256",
			a.name, matches, ksMatches)
		if matches == 256 {
			t.Logf("  *** EXACT MATCH (KSA sbox) for key: %x", a.key)
		}
	}

	// New: KSA using 16-bit j with short key values (CRc4A::SState uses int mM[256])
	{
		sbox := ksaWithShorts(shorts)
		m := matchCount(sbox, target)
		t.Logf("%-42s  KSA sbox: %3d/256", "KSA uint16-j, short key values (9 shorts)", m)
		if m == 256 {
			t.Logf("  *** EXACT MATCH: uint16 j with shorts!")
		}
	}
	{
		sbox := ksaWithShortsHi16(shorts)
		m := matchCount(sbox, target)
		t.Logf("%-42s  KSA sbox: %3d/256", "KSA unbounded j%256, short key (9 shorts)", m)
		if m == 256 {
			t.Logf("  *** EXACT MATCH: unbounded j with shorts!")
		}
	}

	// Also: check if DecryptSbox itself is the output of KSA(DecryptSbox as key)
	sboxAsKey := standardRC4KSA(target[:])
	t.Logf("%-42s  KSA sbox: %3d/256", "KSA(DecryptSbox as key)", matchCount(sboxAsKey, target))

	// Print first 16 bytes for manual inspection
	fmt.Println("\n--- First 16 bytes of KSA S-box per attempt ---")
	for _, a := range attempts {
		sbox := standardRC4KSA(a.key)
		fmt.Printf("%-42s  %02x\n", a.name, sbox[:16])
	}
	{
		sbox := ksaWithShorts(shorts)
		fmt.Printf("%-42s  %02x\n", "KSA uint16-j, short key values", sbox[:16])
	}
	fmt.Printf("%-42s  %02x\n", "Target (DecryptSbox)", target[:16])
}
