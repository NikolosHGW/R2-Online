// Package handler contains game server packet handlers.
package handler

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"r2server/internal/game"
	"r2server/internal/network"
	"r2server/internal/packet/game/recv"
	"r2server/internal/packet/game/send"
	"r2server/internal/packet/opcode"
	"r2server/internal/repository"
)

// EntryDeps are the dependencies for the game entry handler.
type EntryDeps interface {
	// ValidateToken checks the Redis session token and returns the accountID.
	ValidateToken(ctx context.Context, token int32) (accountID int32, err error)

	// GetCharactersByOwner returns all characters for the account.
	GetCharactersByOwner(ctx context.Context, ownerID int32) ([]*repository.Character, error)

	// GetStateByCharacter returns the mutable state (level, pos, stats).
	GetStateByCharacter(ctx context.Context, charID int32) (*repository.PcState, error)

	// GetEquippedByCharacter returns equipped items for visual preview.
	GetEquippedByCharacter(ctx context.Context, charID int32) ([]repository.EquippedItem, error)

	// MarkAccountOffline clears the online flag when the session ends.
	MarkAccountOffline(ctx context.Context, accountID int32) error
}

// HandleLoginUserReq handles opcode 5100 — first packet from the client on the game server.
// Validates the Redis session token, loads characters, and sends InformationCharacter (5101).
func HandleLoginUserReq(srv *game.Server, deps EntryDeps) game.HandlerFunc {
	return func(s *game.Session, r *network.Reader) error {
		var pkt recv.LoginUserReq
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleLoginUserReq: decode: %w", err)
		}

		ctx := context.Background()

		// ── 1. Validate session token from Redis ──────────────────────────────
		accountID, err := deps.ValidateToken(ctx, pkt.SessionID)
		if err != nil || accountID == 0 {
			s.Log().Warn("invalid session token",
				zap.Int32("token", pkt.SessionID),
				zap.Error(err),
			)
			return s.Send(opcode.GameServerError, (&send.GameServerError{
				ErrorCode: send.ErrNoUserNotLogin,
			}).Encode())
		}

		s.AccountID = accountID
		s.SetState(game.StateCharSelect)
		srv.AddSession(accountID, s)

		s.Log().Info("account logged in", zap.Int32("account_id", accountID))

		// ── 2. Load characters from DB ────────────────────────────────────────
		chars, err := deps.GetCharactersByOwner(ctx, accountID)
		if err != nil {
			return fmt.Errorf("HandleLoginUserReq: GetCharacters: %w", err)
		}

		// ── 3. Build InformationCharacter packet ──────────────────────────────
		infoChar := &send.InformationCharacter{}
		for _, c := range chars {
			if int(c.Slot) >= len(infoChar.Characters) {
				continue
			}

			state, _ := deps.GetStateByCharacter(ctx, c.ID)
			equipped, _ := deps.GetEquippedByCharacter(ctx, c.ID)

			slot := send.CharacterSlot{
				CharacterID: c.ID,
				Class:       uint8(c.Class),
				Gender:      uint8(c.Gender),
				Head:        uint8(c.Head),
				Face:        uint8(c.Face),
				Name:        c.Nickname,
			}

			if state != nil {
				slot.Level = state.Level
				slot.PosX = state.PosX
				slot.PosY = state.PosY
				slot.PosZ = state.PosZ
				slot.Strength = state.Strength
				slot.Intelligence = state.Intelligence
				slot.Dexterity = state.Dexterity
				slot.Chaotic = state.Chaotic
			}

			// Map equipped items to their visual slots
			for _, eq := range equipped {
				applyEquipSlot(&slot, eq)
			}

			infoChar.Characters[c.Slot] = slot
		}

		return s.Send(opcode.InformationCharacter, infoChar.Encode())
	}
}

// applyEquipSlot maps a DB equipment slot number to the CharacterSlot preview fields.
// Slot numbers match the R2 Online equipment slot definitions from FieldW.h.
func applyEquipSlot(slot *send.CharacterSlot, eq repository.EquippedItem) {
	switch eq.Slot {
	case 0:
		slot.WeaponItemID = eq.ItemNo
	case 1:
		slot.ShieldItemID = eq.ItemNo
	case 2:
		slot.ArmorItemID = eq.ItemNo
	case 3:
		slot.Ring1ItemID = eq.ItemNo
	case 4:
		slot.Ring2ItemID = eq.ItemNo
	case 5:
		slot.NecklaceItemID = eq.ItemNo
	case 6:
		slot.BootsItemID = eq.ItemNo
	case 7:
		slot.GlovesItemID = eq.ItemNo
	case 8:
		slot.HelmetItemID = eq.ItemNo
	case 9:
		slot.Belt1ItemID = eq.ItemNo
	case 10:
		slot.CloakItemID = eq.ItemNo
	}
}
