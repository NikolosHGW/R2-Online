// Package handler contains login server packet handlers.
//
// Each handler receives the session (for sending responses and reading state)
// and a Reader over the packet payload. Handlers call the repository layer for
// database access and update session.State to drive the state machine.
//
// Add your game logic here — the network engine provides the plumbing.
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

// AuthDeps are the dependencies injected into the auth handler.
// Implement these interfaces in your repository layer.
type AuthDeps interface {
	// GetAccountByLogin returns the account id and hashed password for the given login.
	// Return (0, "", nil) when the account does not exist.
	GetAccountByLogin(loginName string) (accountID int32, hashedPassword string, err error)

	// ValidatePassword checks plain against the stored hash.
	ValidatePassword(plain, hash string) bool

	// IsAccountOnline returns true if the account is already logged in.
	IsAccountOnline(accountID int32) (bool, error)
}

// HandleAuthorizationLogin handles opcode 3100 — client sends credentials.
func HandleAuthorizationLogin(deps AuthDeps) login.HandlerFunc {
	return func(s *login.Session, r *network.Reader) error {
		var pkt recv.AuthorizationLogin
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleAuthorizationLogin: decode: %w", err)
		}

		slog.Info("login attempt", "login", pkt.Login, "remote", "TODO")

		// ── 1. Look up account ────────────────────────────────────────────────
		accountID, hash, err := deps.GetAccountByLogin(pkt.Login)
		if err != nil {
			return fmt.Errorf("HandleAuthorizationLogin: db: %w", err)
		}
		if accountID == 0 {
			return s.Send(opcode.LoginServerError, (&send.LoginServerError{
				ErrorCode: send.ErrInvalidCredentials,
			}).Encode())
		}

		// ── 2. Check password ─────────────────────────────────────────────────
		if !deps.ValidatePassword(pkt.Password, hash) {
			return s.Send(opcode.LoginServerError, (&send.LoginServerError{
				ErrorCode: send.ErrInvalidCredentials,
			}).Encode())
		}

		// ── 3. Check for duplicate login ──────────────────────────────────────
		online, err := deps.IsAccountOnline(accountID)
		if err != nil {
			return fmt.Errorf("HandleAuthorizationLogin: check online: %w", err)
		}
		if online {
			return s.Send(opcode.LoginServerError, (&send.LoginServerError{
				ErrorCode: send.ErrAlreadyOnline,
			}).Encode())
		}

		// ── 4. Mark session as authenticated ─────────────────────────────────
		s.AccountID = accountID
		s.Login = pkt.Login
		s.SetState(login.StateAuthed)

		// ── 5. Send server list ───────────────────────────────────────────────
		// Delegate to the lobby handler — it knows how to fetch servers.
		return sendServerList(s, deps)
	}
}

// sendServerList is a shared helper called after auth and on refresh.
func sendServerList(s *login.Session, deps AuthDeps) error {
	// TODO: replace with a real server list from DB/config
	// (implement GetServers in AuthDeps or add a separate ServerListDeps)
	servers := []send.GameServer{
		{
			ServerID:   1,
			Name:       "R2 Online",
			IP:         "127.0.0.1",
			Port:       5000,
			Type:       1,
			Status:     true,
			Congestion: 10,
		},
	}

	return s.Send(opcode.SendServers, (&send.SendServers{Servers: servers}).Encode())
}
