package send

import "r2server/internal/network"

// LoginServerError is opcode 3102 — notifies the client of an error during login.
//
// Binary layout: [ErrorCode: uint32]
type LoginServerError struct {
	ErrorCode uint32
}

// Known error codes (extend as needed).
const (
	ErrInvalidCredentials uint32 = 1
	ErrAccountBanned      uint32 = 2
	ErrAlreadyOnline      uint32 = 3
	ErrServerFull         uint32 = 4
)

func (p *LoginServerError) Encode() []byte {
	w := network.NewWriter()
	w.WriteUint32(p.ErrorCode)
	return w.Bytes()
}
