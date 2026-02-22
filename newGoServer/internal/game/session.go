package game

import (
	"log/slog"

	"r2server/internal/network"
	"r2server/internal/packet/opcode"
)

// State represents the lifecycle stage of a game session.
type State int

const (
	StateConnected  State = iota // key sent, waiting for LoginUserReq
	StateCharSelect              // character list sent, waiting for character choice
	StateInWorld                 // character chosen, player is in the game world
)

// Session wraps a network.Conn with game-server-specific state.
type Session struct {
	conn   *network.Conn
	server *Server
	state  State

	// Populated after LoginUserReq is validated
	AccountID int32

	// Populated after ChoosePcReq
	CharacterID int32
}

func newSession(conn *network.Conn, srv *Server) *Session {
	return &Session{conn: conn, server: srv}
}

func (s *Session) SetState(st State) { s.state = st }

// Send encodes the packet payload and writes it to the wire.
func (s *Session) Send(op opcode.Opcode, payload []byte) error {
	return s.conn.Send(op, payload)
}

// SetCipher installs the RC4 cipher on the underlying connection.
func (s *Session) SetCipher(key [256]byte) {
	s.conn.SetCipher(key)
}

// Close closes the underlying TCP connection.
func (s *Session) Close() {
	s.conn.Close()
}

// run is the session read loop — blocks until the connection closes.
func (s *Session) run() {
	defer func() {
		s.conn.Close()
		slog.Info("game session closed",
			"remote", s.conn.RemoteAddr(),
			"account", s.AccountID,
			"character", s.CharacterID,
		)
		// Notify server so it can clean up world state
		s.server.onSessionEnd(s)
	}()

	for {
		op, data, err := s.conn.Recv()
		if err != nil {
			slog.Debug("game recv", "remote", s.conn.RemoteAddr(), "err", err)
			return
		}

		r := network.NewReader(data)
		if err := s.server.dispatch(s, op, r); err != nil {
			slog.Warn("game dispatch error",
				"remote", s.conn.RemoteAddr(),
				"opcode", op,
				"err", err,
			)
		}
	}
}
