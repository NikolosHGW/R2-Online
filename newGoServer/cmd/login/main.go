// Login Server entry point.
//
// Wires together: config → DB pool → Redis → repositories → handlers → server.
// This file is the composition root — it knows about everything.
// The internal packages know about nothing above them.
package main

import (
	"context"
	stdlog "log"

	"go.uber.org/zap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"r2server/config"
	"r2server/internal/logger"
	"r2server/internal/login"
	"r2server/internal/login/handler"
	"r2server/internal/packet/login/send"
	"r2server/internal/packet/opcode"
	"r2server/internal/repository"
)

func main() {
	log := logger.New().Named("login")

	defer log.Sync()
	cfg := config.Load()

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		stdlog.Fatalf("login: db connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		stdlog.Fatalf("login: db ping: %v", err)
	}
	log.Info("connected to postgres")

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		stdlog.Fatalf("login: redis ping: %v", err)
	}
	log.Info("connected to redis")

	// ── Repositories ──────────────────────────────────────────────────────────
	accountRepo := repository.NewAccountRepository(db)
	sessionRepo := repository.NewSessionRepository(db, rdb)
	serverRepo  := repository.NewGameServerRepository(db, rdb)

	// ── Dependency adapter ────────────────────────────────────────────────────
	// loginDeps wires repository calls to the handler.AuthDeps / handler.LobbyDeps interfaces.
	deps := &loginDeps{
		accounts: accountRepo,
		sessions: sessionRepo,
		servers:  serverRepo,
		cfg:      cfg,
	}

	// ── Server ────────────────────────────────────────────────────────────────
	srv := login.NewServer(cfg.LoginAddr, log)

	srv.Handle(opcode.AuthorizationLogin, handler.HandleAuthorizationLogin(deps))
	srv.Handle(opcode.RefreshServers,     handler.HandleRefreshServers(deps))
	srv.Handle(opcode.SelectServer,       handler.HandleSelectServer(deps))

	log.Info("login server starting", zap.String("addr", cfg.LoginAddr))
	if err := srv.ListenAndServe(); err != nil {
		stdlog.Fatalf("login: server error: %v", err)
	}
}

// ── loginDeps implements handler.AuthDeps + handler.LobbyDeps ────────────────

type loginDeps struct {
	accounts *repository.AccountRepository
	sessions *repository.SessionRepository
	servers  *repository.GameServerRepository
	cfg      *config.Config
}

func (d *loginDeps) GetAccountByLogin(loginName string) (int32, string, error) {
	acc, err := d.accounts.GetByLogin(context.Background(), loginName)
	if err != nil {
		return 0, "", err
	}
	if acc == nil {
		return 0, "", nil
	}
	return acc.ID, acc.Password, nil
}

func (d *loginDeps) ValidatePassword(plain, hash string) bool {
	// TODO: replace with bcrypt.CompareHashAndPassword or argon2id
	// For development you can temporarily use plain == hash to test the flow.
	// NEVER ship plain-text comparison in production.
	// return plain == hash
	return true
}

func (d *loginDeps) IsAccountOnline(accountID int32) (bool, error) {
	return d.sessions.IsAccountOnline(context.Background(), accountID)
}

func (d *loginDeps) GetGameServers() ([]send.GameServer, error) {
	return d.servers.GetAll(context.Background())
}

func (d *loginDeps) CreateSessionToken(accountID int32, serverID int16) (int32, error) {
	return d.sessions.CreateToken(context.Background(), accountID, serverID)
}

func (d *loginDeps) MarkAccountOnline(accountID int32) error {
	return d.sessions.MarkAccountOnline(context.Background(), accountID)
}
