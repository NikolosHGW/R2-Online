package handler

import (
	"fmt"
	"log/slog"

	"r2server/internal/login"
	"r2server/internal/network"
	"r2server/internal/packet/login/recv"
	"r2server/internal/packet/login/send"
	"r2server/internal/packet/opcode"
)

// LobbyDeps are the dependencies for lobby-phase handlers.
type LobbyDeps interface {
	AuthDeps

	// GetGameServers returns the list of available game servers.
	GetGameServers() ([]send.GameServer, error)

	// CreateSessionToken stores a token in Redis linking accountID → gameServerID.
	// The token is later validated by the game server.
	CreateSessionToken(accountID int32, serverID int16) (token int32, err error)

	// MarkAccountOnline records that the account is in the process of entering a game server.
	MarkAccountOnline(accountID int32) error
}

// HandleRefreshServers handles opcode 3115 — client wants an updated server list.
func HandleRefreshServers(deps LobbyDeps) login.HandlerFunc {
	return func(s *login.Session, r *network.Reader) error {
		var pkt recv.RefreshServers
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleRefreshServers: decode: %w", err)
		}

		servers, err := deps.GetGameServers()
		if err != nil {
			return fmt.Errorf("HandleRefreshServers: GetGameServers: %w", err)
		}

		return s.Send(opcode.RefreshedServers, (&send.RefreshedServers{Servers: servers}).Encode())
	}
}

// HandleSelectServer handles opcode 3120 — client picks a game server to play on.
func HandleSelectServer(deps LobbyDeps) login.HandlerFunc {
	return func(s *login.Session, r *network.Reader) error {
		var pkt recv.SelectServer
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleSelectServer: decode: %w", err)
		}

		if s.AccountID == 0 {
			return fmt.Errorf("HandleSelectServer: session not authenticated")
		}

		// ── 1. Create a session token in Redis (game server validates it) ─────
		token, err := deps.CreateSessionToken(s.AccountID, pkt.ServerID)
		if err != nil {
			return fmt.Errorf("HandleSelectServer: CreateSessionToken: %w", err)
		}

		// ── 2. Mark account as transitioning to game server ───────────────────
		if err := deps.MarkAccountOnline(s.AccountID); err != nil {
			slog.Warn("HandleSelectServer: MarkAccountOnline failed", "err", err)
		}

		slog.Info("client redirected to game server",
			"account", s.AccountID,
			"server", pkt.ServerID,
			"token", token,
		)

		// ── 3. Send SelectedServer (opcode 3121) — client will reconnect ──────
		s.SetState(login.StateRedirected)
		return s.Send(opcode.SelectedServer, (&send.SelectedServer{}).Encode())
	}
}
