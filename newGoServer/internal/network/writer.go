package network

import (
	"encoding/binary"
	"math"

	"golang.org/x/text/encoding/charmap"
)

// Writer builds a binary packet payload in little-endian order.
// Mirrors the FormationPackage write methods from the C# emulator.
type Writer struct {
	buf []byte
}

func NewWriter() *Writer {
	return &Writer{buf: make([]byte, 0, 128)}
}

func (w *Writer) WriteUint8(v byte) {
	w.buf = append(w.buf, v)
}

func (w *Writer) WriteInt8(v int8) {
	w.buf = append(w.buf, byte(v))
}

func (w *Writer) WriteUint16(v uint16) {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *Writer) WriteInt16(v int16) {
	w.WriteUint16(uint16(v))
}

func (w *Writer) WriteUint16BE(v uint16) {
	b := [2]byte{}
	binary.BigEndian.PutUint16(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *Writer) WriteUint32(v uint32) {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *Writer) WriteInt32(v int32) {
	w.WriteUint32(uint32(v))
}

func (w *Writer) WriteUint64(v uint64) {
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *Writer) WriteInt64(v int64) {
	w.WriteUint64(uint64(v))
}

func (w *Writer) WriteFloat32(v float32) {
	w.WriteUint32(math.Float32bits(v))
}

func (w *Writer) WriteBytes(v []byte) {
	w.buf = append(w.buf, v...)
}

// WriteZero writes n zero bytes (padding).
func (w *Writer) WriteZero(n int) {
	w.buf = append(w.buf, make([]byte, n)...)
}

// WriteStringCP1251 encodes s to CP-1251 and writes it into a fixed-size field
// of maxLen bytes, zero-padded. If s is longer than maxLen-1 it is truncated.
func (w *Writer) WriteStringCP1251(s string, maxLen int) {
	encoded, err := charmap.Windows1251.NewEncoder().Bytes([]byte(s))
	if err != nil {
		encoded = []byte(s) // fallback: write raw bytes
	}
	field := make([]byte, maxLen)
	n := copy(field, encoded)
	if n < maxLen {
		field[n] = 0 // null terminator
	}
	w.buf = append(w.buf, field...)
}

// Bytes returns the accumulated payload bytes.
func (w *Writer) Bytes() []byte {
	return w.buf
}

// Len returns the number of bytes written so far.
func (w *Writer) Len() int {
	return len(w.buf)
}
