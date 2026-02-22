package handler

import (
	"context"
	"fmt"
	"log/slog"

	"r2server/internal/game"
	"r2server/internal/network"
	"r2server/internal/packet/game/recv"
	"r2server/internal/packet/game/send"
	"r2server/internal/packet/opcode"
	"r2server/internal/repository"
)

// CharacterDeps are the dependencies for character lifecycle handlers.
type CharacterDeps interface {
	// GetCharacterByOwnerAndSlot returns the character in the given slot for the account.
	GetCharacterByOwnerAndSlot(ctx context.Context, ownerID int32, slot uint8) (*repository.Character, error)

	// CreateCharacter inserts a new character and its initial state. Returns the new ID.
	CreateCharacter(ctx context.Context, c *repository.Character, initialHP, initialMP int32) (int32, error)

	// DeleteCharacter soft-deletes a character belonging to the account.
	DeleteCharacter(ctx context.Context, charID, ownerID int32) error

	// RecordLogin updates last_login timestamp.
	RecordLogin(ctx context.Context, charID int32) error
}

// HandleChoosePcReq handles opcode 5116 — client selects a character slot.
func HandleChoosePcReq(deps CharacterDeps) game.HandlerFunc {
	return func(s *game.Session, r *network.Reader) error {
		var pkt recv.ChoosePcReq
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleChoosePcReq: decode: %w", err)
		}

		if pkt.Slot > 2 {
			return s.Send(opcode.GameServerError, (&send.GameServerError{
				ErrorCode: send.ErrNoCharInvalidSlot,
			}).Encode())
		}

		ctx := context.Background()
		char, err := deps.GetCharacterByOwnerAndSlot(ctx, s.AccountID, pkt.Slot)
		if err != nil {
			return fmt.Errorf("HandleChoosePcReq: db: %w", err)
		}
		if char == nil {
			return s.Send(opcode.GameServerError, (&send.GameServerError{
				ErrorCode: send.ErrNoCharInvalidSlot,
			}).Encode())
		}

		s.CharacterID = char.ID
		s.SetState(game.StateInWorld)

		if err := deps.RecordLogin(ctx, char.ID); err != nil {
			slog.Warn("HandleChoosePcReq: RecordLogin failed", "err", err)
		}

		slog.Info("character entered world",
			"account", s.AccountID,
			"character", char.ID,
			"name", char.Nickname,
		)

		// ── Tell the client world entry is complete ───────────────────────────
		// TODO: here you would also broadcast EnteredPcAck to nearby players,
		//       send ServerTime, GameConfiguration, InventoryCharacteristic, etc.
		return s.Send(opcode.CompleteEnterWorld, (&send.CompleteEnterWorld{}).Encode())
	}
}

// HandleCreatePcReq handles opcode 5118 — client creates a new character.
func HandleCreatePcReq(deps CharacterDeps) game.HandlerFunc {
	return func(s *game.Session, r *network.Reader) error {
		var pkt recv.CreatePcReq
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleCreatePcReq: decode: %w", err)
		}

		ctx := context.Background()

		c := &repository.Character{
			OwnerID:  s.AccountID,
			Slot:     int16(pkt.Slot),
			Nickname: pkt.Name,
			Class:    int16(pkt.Class),
			Gender:   int16(pkt.Gender),
			Head:     int16(pkt.Head),
			Face:     int16(pkt.Face),
			HomeMapNo: 1, // TODO: map starting position from class defaults
		}

		// TODO: derive initialHP/MP from class defaults
		_, err := deps.CreateCharacter(ctx, c, 100, 100)
		if err != nil {
			return fmt.Errorf("HandleCreatePcReq: CreateCharacter: %w", err)
		}

		slog.Info("character created",
			"account", s.AccountID,
			"name", pkt.Name,
			"class", pkt.Class,
		)

		return s.Send(opcode.CompleteCreateCharacter, (&send.CompleteCreateCharacter{}).Encode())
	}
}

// HandleDeletePcReq handles opcode 5120 — client deletes a character.
func HandleDeletePcReq(deps CharacterDeps) game.HandlerFunc {
	return func(s *game.Session, r *network.Reader) error {
		var pkt recv.DeletePcReq
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleDeletePcReq: decode: %w", err)
		}

		ctx := context.Background()
		if err := deps.DeleteCharacter(ctx, pkt.CharacterID, s.AccountID); err != nil {
			return fmt.Errorf("HandleDeletePcReq: DeleteCharacter: %w", err)
		}

		slog.Info("character deleted",
			"account", s.AccountID,
			"character", pkt.CharacterID,
		)

		return s.Send(opcode.CompleteDeleteCharacter, (&send.CompleteDeleteCharacter{}).Encode())
	}
}

// HandleLogoutPcReq handles opcode 5115 — client requests logout.
func HandleLogoutPcReq() game.HandlerFunc {
	return func(s *game.Session, r *network.Reader) error {
		slog.Info("client logout",
			"account", s.AccountID,
			"character", s.CharacterID,
		)
		// TODO: save character state to DB, broadcast ExistedPcAck to nearby players
		s.Close()
		return nil
	}
}
