// Game Server entry point.
package main

import (
	"context"
	stdlog "log"

	"go.uber.org/zap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"r2server/config"
	"r2server/internal/game"
	"r2server/internal/game/handler"
	"r2server/internal/logger"
	"r2server/internal/packet/opcode"
	"r2server/internal/repository"
)

func main() {
	log := logger.New().Named("game")
	defer log.Sync()

	cfg := config.Load()

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		stdlog.Fatalf("game: db connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		stdlog.Fatalf("game: db ping: %v", err)
	}
	log.Info("connected to postgres")

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		stdlog.Fatalf("game: redis ping: %v", err)
	}
	log.Info("connected to redis")

	// ── Repositories ──────────────────────────────────────────────────────────
	sessionRepo   := repository.NewSessionRepository(db, rdb)
	characterRepo := repository.NewCharacterRepository(db)

	// ── Server ────────────────────────────────────────────────────────────────
	srv := game.NewServer(cfg.GameAddr, log)

	// Entry deps — wraps repos for LoginUserReq handler
	entryDeps := &gameDeps{sessions: sessionRepo, characters: characterRepo}

	srv.Handle(opcode.LoginUserReq,        handler.HandleLoginUserReq(srv, entryDeps))
	srv.Handle(opcode.ChoosePcReq,         handler.HandleChoosePcReq(entryDeps))
	srv.Handle(opcode.CreatePcReq,         handler.HandleCreatePcReq(entryDeps))
	srv.Handle(opcode.DeletePcReq,         handler.HandleDeletePcReq(entryDeps))
	srv.Handle(opcode.LogoutPcReq,         handler.HandleLogoutPcReq())

	// Cleanup on disconnect
	srv.OnSessionEnd(func(s *game.Session) {
		if s.AccountID != 0 {
			_ = sessionRepo.MarkAccountOffline(context.Background(), s.AccountID)
		}
		// TODO: save character state, broadcast ExistedPcAck to nearby players
	})

	log.Info("game server starting", zap.String("addr", cfg.GameAddr))
	if err := srv.ListenAndServe(); err != nil {
		stdlog.Fatalf("game: server error: %v", err)
	}
}

// ── gameDeps implements handler.EntryDeps + handler.CharacterDeps ────────────

type gameDeps struct {
	sessions   *repository.SessionRepository
	characters *repository.CharacterRepository
}

// ── EntryDeps ─────────────────────────────────────────────────────────────────

func (d *gameDeps) ValidateToken(ctx context.Context, token int32) (int32, error) {
	return d.sessions.ValidateToken(ctx, token)
}

func (d *gameDeps) GetCharactersByOwner(ctx context.Context, ownerID int32) ([]*repository.Character, error) {
	return d.characters.GetByOwner(ctx, ownerID)
}

func (d *gameDeps) GetStateByCharacter(ctx context.Context, charID int32) (*repository.PcState, error) {
	return d.characters.GetState(ctx, charID)
}

func (d *gameDeps) GetEquippedByCharacter(ctx context.Context, charID int32) ([]repository.EquippedItem, error) {
	return d.characters.GetEquipped(ctx, charID)
}

func (d *gameDeps) MarkAccountOffline(ctx context.Context, accountID int32) error {
	return d.sessions.MarkAccountOffline(ctx, accountID)
}

// ── CharacterDeps ─────────────────────────────────────────────────────────────

func (d *gameDeps) GetCharacterByOwnerAndSlot(ctx context.Context, ownerID int32, slot uint8) (*repository.Character, error) {
	chars, err := d.characters.GetByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	for _, c := range chars {
		if c.Slot == int16(slot) {
			return c, nil
		}
	}
	return nil, nil
}

func (d *gameDeps) CreateCharacter(ctx context.Context, c *repository.Character, initialHP, initialMP int32) (int32, error) {
	return d.characters.Create(ctx, c, initialHP, initialMP)
}

func (d *gameDeps) DeleteCharacter(ctx context.Context, charID, ownerID int32) error {
	return d.characters.SoftDelete(ctx, charID, ownerID)
}

func (d *gameDeps) RecordLogin(ctx context.Context, charID int32) error {
	return d.characters.RecordLogin(ctx, charID)
}
