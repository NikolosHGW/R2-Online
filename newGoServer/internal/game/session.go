package game

import (
	"fmt"

	"go.uber.org/zap"

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
	log    *zap.Logger

	// Populated after LoginUserReq is validated
	AccountID int32

	// Populated after ChoosePcReq
	CharacterID int32
}

func newSession(conn *network.Conn, srv *Server) *Session {
	return &Session{
		conn:   conn,
		server: srv,
		log:    srv.log.With(zap.String("remote", conn.RemoteAddr().String())),
	}
}

// Log returns the session-scoped logger (for use in handlers).
func (s *Session) Log() *zap.Logger { return s.log }

// RemoteAddr returns the remote address as a string.
func (s *Session) RemoteAddr() string { return s.conn.RemoteAddr().String() }

func (s *Session) SetState(st State) { s.state = st }

// Send encodes the packet payload and writes it to the wire.
func (s *Session) Send(op opcode.Opcode, payload []byte) error {
	return s.conn.Send(op, payload)
}

// Close closes the underlying TCP connection.
func (s *Session) Close() {
	s.conn.Close()
}

// run is the session read loop — blocks until the connection closes.
func (s *Session) run() {
	s.log.Info("game session started")
	defer func() {
		s.conn.Close()
		s.log.Info("game session closed",
			zap.Int32("account", s.AccountID),
			zap.Int32("character", s.CharacterID),
		)
		s.server.onSessionEnd(s)
	}()

	for {
		op, data, err := s.conn.Recv()
		if err != nil {
			s.log.Debug("recv error", zap.Error(err))
			return
		}

		s.log.Info("packet received",
			zap.String("op", fmt.Sprintf("0x%04X(%d)", op, op)),
			zap.Int("payload_len", len(data)),
		)

		r := network.NewReader(data)
		if err := s.server.dispatch(s, opcode.Opcode(op), r); err != nil {
			s.log.Warn("dispatch error",
				zap.String("op", fmt.Sprintf("0x%04X(%d)", op, op)),
				zap.Error(err),
			)
		}
	}
}
