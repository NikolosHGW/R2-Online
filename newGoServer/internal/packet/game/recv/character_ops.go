package recv

import (
	"fmt"

	"r2server/internal/network"
)

// ChoosePcReq is opcode 5116 — client selects a character slot to play.
//
// Binary layout: [Slot: uint8]
type ChoosePcReq struct {
	Slot uint8
}

func (p *ChoosePcReq) Decode(r *network.Reader) error {
	var err error
	p.Slot, err = r.ReadUint8()
	if err != nil {
		return fmt.Errorf("ChoosePcReq: Slot: %w", err)
	}
	return nil
}

// CreatePcReq is opcode 5118 — client requests character creation.
//
// Binary layout (approximate from C# emulator CreatePcReqModel):
//
//	[Slot: uint8] [Class: uint8] [Gender: uint8] [Head: uint8] [Face: uint8]
//	[Name: 15 bytes CP-1251]
type CreatePcReq struct {
	Slot   uint8
	Class  uint8
	Gender uint8
	Head   uint8
	Face   uint8
	Name   string // max 15 chars
}

func (p *CreatePcReq) Decode(r *network.Reader) error {
	var err error
	p.Slot, err = r.ReadUint8()
	if err != nil {
		return fmt.Errorf("CreatePcReq: Slot: %w", err)
	}
	p.Class, err = r.ReadUint8()
	if err != nil {
		return fmt.Errorf("CreatePcReq: Class: %w", err)
	}
	p.Gender, err = r.ReadUint8()
	if err != nil {
		return fmt.Errorf("CreatePcReq: Gender: %w", err)
	}
	p.Head, err = r.ReadUint8()
	if err != nil {
		return fmt.Errorf("CreatePcReq: Head: %w", err)
	}
	p.Face, err = r.ReadUint8()
	if err != nil {
		return fmt.Errorf("CreatePcReq: Face: %w", err)
	}
	p.Name, err = r.ReadStringCP1251(15)
	if err != nil {
		return fmt.Errorf("CreatePcReq: Name: %w", err)
	}
	return nil
}

// DeletePcReq is opcode 5120 — client requests character deletion.
//
// Binary layout: [CharacterId: int32]
type DeletePcReq struct {
	CharacterID int32
}

func (p *DeletePcReq) Decode(r *network.Reader) error {
	v, err := r.ReadUint32()
	if err != nil {
		return fmt.Errorf("DeletePcReq: CharacterId: %w", err)
	}
	p.CharacterID = int32(v)
	return nil
}

// LogoutPcReq is opcode 5115 — client requests logout. No payload.
type LogoutPcReq struct{}

func (p *LogoutPcReq) Decode(_ *network.Reader) error { return nil }
