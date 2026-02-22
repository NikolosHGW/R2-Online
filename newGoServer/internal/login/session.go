package login

import (
	"log/slog"

	"r2server/internal/network"
	"r2server/internal/packet/opcode"
)

// State represents the lifecycle stage of a login session.
type State int

const (
	StateConnected  State = iota // connected, waiting for AuthorizationLogin (plaintext)
	StateAuthed                  // authenticated, waiting for SelectServer
	StateRedirected              // SelectedServer sent, client will disconnect
)

// Session wraps a network.Conn with login-server-specific state.
type Session struct {
	conn      *network.Conn
	server    *Server
	state     State

	// Populated after successful authentication
	AccountID int32
	Login     string
}

func newSession(conn *network.Conn, srv *Server) *Session {
	return &Session{conn: conn, server: srv}
}

// Send encodes the packet and writes it to the wire.
func (s *Session) Send(op opcode.Opcode, payload []byte) error {
	return s.conn.Send(op, payload)
}

// SetState updates the session state machine.
func (s *Session) SetState(st State) { s.state = st }

// Close closes the underlying connection.
func (s *Session) Close() {
	s.conn.Close()
}

// run is the session read loop — runs in its own goroutine until the connection closes.
func (s *Session) run() {
	defer func() {
		s.conn.Close()
		slog.Info("login session closed", "remote", s.conn.RemoteAddr())
	}()

	for {
		op, data, err := s.conn.Recv()
		if err != nil {
			slog.Debug("login recv", "remote", s.conn.RemoteAddr(), "err", err)
			return
		}

		r := network.NewReader(data)
		if err := s.server.dispatch(s, op, r); err != nil {
			slog.Warn("login dispatch error",
				"remote", s.conn.RemoteAddr(),
				"opcode", op,
				"err", err,
			)
		}

		if s.state == StateRedirected {
			return // client will reconnect to game server
		}
	}
}
