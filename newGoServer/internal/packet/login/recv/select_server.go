package recv

import (
	"fmt"

	"r2server/internal/network"
)

// SelectServer is opcode 3120 — client picks a game server from the list.
//
// Binary layout:
//
//	[AccountId: uint32] [Login: 20 bytes CP-1251] [ServerId: int16]
type SelectServer struct {
	AccountID int32
	Login     string
	ServerID  int16
}

func (p *SelectServer) Decode(r *network.Reader) error {
	var err error
	accountID, err := r.ReadUint32()
	if err != nil {
		return fmt.Errorf("SelectServer: AccountId: %w", err)
	}
	p.AccountID = int32(accountID)

	p.Login, err = r.ReadStringCP1251(20)
	if err != nil {
		return fmt.Errorf("SelectServer: Login: %w", err)
	}

	p.ServerID, err = r.ReadInt16()
	if err != nil {
		return fmt.Errorf("SelectServer: ServerId: %w", err)
	}

	return nil
}

// RefreshServers is opcode 3115 — client requests an updated server list.
// No payload.
type RefreshServers struct{}

func (p *RefreshServers) Decode(_ *network.Reader) error { return nil }
