package network

import (
	"encoding/binary"
	"fmt"
	"math"

	"golang.org/x/text/encoding/charmap"
)

// Reader reads binary data from a byte slice in little-endian order.
// Mirrors the FormationPackage read methods from the C# emulator.
type Reader struct {
	buf []byte
	pos int
}

func NewReader(data []byte) *Reader {
	return &Reader{buf: data}
}

func (r *Reader) Remaining() int {
	return len(r.buf) - r.pos
}

func (r *Reader) ReadUint8() (byte, error) {
	if r.pos >= len(r.buf) {
		return 0, fmt.Errorf("reader: ReadUint8 out of bounds at %d", r.pos)
	}
	v := r.buf[r.pos]
	r.pos++
	return v, nil
}

func (r *Reader) ReadInt8() (int8, error) {
	v, err := r.ReadUint8()
	return int8(v), err
}

func (r *Reader) ReadUint16() (uint16, error) {
	if r.pos+2 > len(r.buf) {
		return 0, fmt.Errorf("reader: ReadUint16 out of bounds at %d", r.pos)
	}
	v := binary.LittleEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v, nil
}

func (r *Reader) ReadInt16() (int16, error) {
	v, err := r.ReadUint16()
	return int16(v), err
}

func (r *Reader) ReadUint16BE() (uint16, error) {
	if r.pos+2 > len(r.buf) {
		return 0, fmt.Errorf("reader: ReadUint16BE out of bounds at %d", r.pos)
	}
	v := binary.BigEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v, nil
}

func (r *Reader) ReadUint32() (uint32, error) {
	if r.pos+4 > len(r.buf) {
		return 0, fmt.Errorf("reader: ReadUint32 out of bounds at %d", r.pos)
	}
	v := binary.LittleEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *Reader) ReadInt32() (int32, error) {
	v, err := r.ReadUint32()
	return int32(v), err
}

func (r *Reader) ReadUint64() (uint64, error) {
	if r.pos+8 > len(r.buf) {
		return 0, fmt.Errorf("reader: ReadUint64 out of bounds at %d", r.pos)
	}
	v := binary.LittleEndian.Uint64(r.buf[r.pos:])
	r.pos += 8
	return v, nil
}

func (r *Reader) ReadInt64() (int64, error) {
	v, err := r.ReadUint64()
	return int64(v), err
}

func (r *Reader) ReadFloat32() (float32, error) {
	bits, err := r.ReadUint32()
	return math.Float32frombits(bits), err
}

// ReadBytes reads exactly n bytes.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.buf) {
		return nil, fmt.Errorf("reader: ReadBytes(%d) out of bounds at %d (len=%d)", n, r.pos, len(r.buf))
	}
	v := make([]byte, n)
	copy(v, r.buf[r.pos:r.pos+n])
	r.pos += n
	return v, nil
}

// Skip advances the position by n bytes without reading them.
func (r *Reader) Skip(n int) error {
	if r.pos+n > len(r.buf) {
		return fmt.Errorf("reader: Skip(%d) out of bounds at %d", n, r.pos)
	}
	r.pos += n
	return nil
}

// ReadStringCP1251 reads a null-terminated CP-1251 string from a fixed-size field
// of maxLen bytes, decoding it to UTF-8. If maxLen is 0, reads until end of buffer.
func (r *Reader) ReadStringCP1251(maxLen int) (string, error) {
	var raw []byte
	if maxLen > 0 {
		data, err := r.ReadBytes(maxLen)
		if err != nil {
			return "", err
		}
		// Trim at null terminator
		for i, b := range data {
			if b == 0 {
				raw = data[:i]
				break
			}
		}
		if raw == nil {
			raw = data
		}
	} else {
		// Read until null or end of buffer
		start := r.pos
		for r.pos < len(r.buf) && r.buf[r.pos] != 0 {
			r.pos++
		}
		raw = r.buf[start:r.pos]
		if r.pos < len(r.buf) {
			r.pos++ // consume null terminator
		}
	}

	decoded, err := charmap.Windows1251.NewDecoder().Bytes(raw)
	if err != nil {
		return string(raw), nil // fallback: treat as ASCII
	}
	return string(decoded), nil
}

// ByteAt reads a byte at an absolute offset without advancing the current position.
func (r *Reader) ByteAt(offset int) (byte, error) {
	if offset >= len(r.buf) {
		return 0, fmt.Errorf("reader: ByteAt(%d) out of bounds (len=%d)", offset, len(r.buf))
	}
	return r.buf[offset], nil
}

// Slice returns a sub-slice of the underlying buffer at an absolute range.
func (r *Reader) Slice(offset, length int) ([]byte, error) {
	if offset+length > len(r.buf) {
		return nil, fmt.Errorf("reader: Slice(%d, %d) out of bounds (len=%d)", offset, length, len(r.buf))
	}
	return r.buf[offset : offset+length], nil
}
