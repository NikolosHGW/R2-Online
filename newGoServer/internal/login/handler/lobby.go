package handler

import (
	"fmt"

	"go.uber.org/zap"

	"r2server/internal/login"
	"r2server/internal/network"
	"r2server/internal/packet/login/recv"
	"r2server/internal/packet/login/send"
	"r2server/internal/packet/opcode"
)

// LobbyDeps are the dependencies for lobby-phase handlers.
type LobbyDeps interface {
	AuthDeps

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

		return s.Send(opcode.RefreshedServers, (&send.RefreshedServers{
			AccountID: s.AccountID,
			SessionID: s.SessionID,
			Servers:   servers,
		}).Encode())
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

		// ── 1. Mark account as transitioning to game server ───────────────────
		if err := deps.MarkAccountOnline(s.AccountID); err != nil {
			s.Log().Warn("MarkAccountOnline failed", zap.Error(err))
		}

		s.Log().Info("client redirected to game server",
			zap.Int32("account_id", s.AccountID),
			zap.Int16("server_id", pkt.ServerID),
			zap.Int32("token", s.SessionID),
		)

		// ── 2. Send SelectedServer (opcode 3121) — client will reconnect ──────
		s.SetState(login.StateRedirected)
		return s.Send(opcode.SelectedServer, (&send.SelectedServer{}).Encode())
	}
}
