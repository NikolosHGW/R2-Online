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

	"go.uber.org/zap"

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

	// CreateSessionToken stores a token in Redis linking accountID → gameServerID.
	// The token is sent to the client in SendServers and echoed back in LoginUserReq.
	CreateSessionToken(accountID int32, serverID int16) (token int32, err error)

	// GetGameServers returns the list of available game servers.
	GetGameServers() ([]send.GameServer, error)
}

// HandleAuthorizationLogin handles opcode 3100 — client sends credentials.
func HandleAuthorizationLogin(deps AuthDeps) login.HandlerFunc {
	return func(s *login.Session, r *network.Reader) error {
		var pkt recv.AuthorizationLogin
		if err := pkt.Decode(r); err != nil {
			return fmt.Errorf("HandleAuthorizationLogin: decode: %w", err)
		}

		s.Log().Info("login attempt",
			zap.String("login", pkt.Login),
			zap.String("remote", s.RemoteAddr()),
		)

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

		// ── 4. Create session token ───────────────────────────────────────────
		// Token is sent to the client in SendServers and echoed back by the client
		// in LoginUserReq (5100) to the game server.
		token, err := deps.CreateSessionToken(accountID, 0)
		if err != nil {
			return fmt.Errorf("HandleAuthorizationLogin: CreateSessionToken: %w", err)
		}

		// ── 5. Mark session as authenticated ─────────────────────────────────
		s.AccountID = accountID
		s.SessionID = token
		s.Login = pkt.Login
		s.SetState(login.StateAuthed)

		s.Log().Info("session token created",
			zap.Int32("account_id", accountID),
			zap.Int32("token", token),
		)

		// ── 6. Send server list ───────────────────────────────────────────────
		return sendServerList(s, deps)
	}
}

// sendServerList is a shared helper called after auth and on refresh.
func sendServerList(s *login.Session, deps AuthDeps) error {
	servers, err := deps.GetGameServers()
	if err != nil {
		return fmt.Errorf("sendServerList: GetGameServers: %w", err)
	}

	return s.Send(opcode.SendServers, (&send.SendServers{
		AccountID: s.AccountID,
		SessionID: s.SessionID,
		Servers:   servers,
	}).Encode())
}
