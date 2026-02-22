// Package recv contains incoming (Client→Server) packet decoders for the Game Server.
package recv

import (
	"fmt"

	"r2server/internal/network"
)

// LoginUserReq is opcode 5100 — first packet sent to the game server after
// the client receives the redirect from the login server.
//
// Binary layout:
//
//	[AccountId: uint32] [SessionId: int32] [Pad: 4 bytes] [Password: 21 bytes CP-1251]
//
// SessionId here corresponds to the session token stored in Redis; use it to
// validate that this client was legitimately redirected by the login server.
type LoginUserReq struct {
	AccountID int32
	SessionID int32
	Password  string // 21 bytes max
}

func (p *LoginUserReq) Decode(r *network.Reader) error {
	accountID, err := r.ReadUint32()
	if err != nil {
		return fmt.Errorf("LoginUserReq: AccountId: %w", err)
	}
	p.AccountID = int32(accountID)

	p.SessionID, err = r.ReadInt32()
	if err != nil {
		return fmt.Errorf("LoginUserReq: SessionId: %w", err)
	}

	if err = r.Skip(4); err != nil { // padding
		return fmt.Errorf("LoginUserReq: pad: %w", err)
	}

	p.Password, err = r.ReadStringCP1251(21)
	if err != nil {
		return fmt.Errorf("LoginUserReq: Password: %w", err)
	}
	return nil
}
