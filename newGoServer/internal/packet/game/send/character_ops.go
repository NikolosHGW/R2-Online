package send

import "r2server/internal/network"

// CompleteEnterWorld is opcode 5117 — sent after the client selects a character.
// Signals that world entry is complete. Payload is empty in observed traffic.
type CompleteEnterWorld struct{}

func (p *CompleteEnterWorld) Encode() []byte { return nil }

// CompleteCreateCharacter is opcode 5119 — character creation acknowledged.
// No payload beyond the opcode itself.
type CompleteCreateCharacter struct{}

func (p *CompleteCreateCharacter) Encode() []byte { return nil }

// CompleteDeleteCharacter is opcode 5121 — character deletion acknowledged.
// No payload.
type CompleteDeleteCharacter struct{}

func (p *CompleteDeleteCharacter) Encode() []byte { return nil }

// GameServerError is opcode 1102 — generic error response on the game server.
//
// Binary layout: [ErrorCode: uint32]
type GameServerError struct {
	ErrorCode uint32
}

// Game server error codes (from C# GameServerErrorType enum).
const (
	ErrNoUserNotLogin         uint32 = 3461165659
	ErrNoUserChkAlreadyLogined uint32 = 4032203570
	ErrNoCharInvalidSlot      uint32 = 2056329710
)

func (p *GameServerError) Encode() []byte {
	w := network.NewWriter()
	w.WriteUint32(p.ErrorCode)
	return w.Bytes()
}
