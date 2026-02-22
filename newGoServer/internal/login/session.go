package login

import (
	"fmt"

	"go.uber.org/zap"

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
	conn   *network.Conn
	server *Server
	state  State
	log    *zap.Logger

	// Populated after successful authentication
	AccountID int32
	SessionID int32 // Redis token sent to client in SendServers, echoed back in LoginUserReq
	Login     string
}

func newSession(conn *network.Conn, srv *Server) *Session {
	return &Session{
		conn:   conn,
		server: srv,
		log:    srv.log.With(zap.String("remote", conn.RemoteAddr().String())),
	}
}

// Send encodes the packet and writes it to the wire.
func (s *Session) Send(op opcode.Opcode, payload []byte) error {
	return s.conn.Send(op, payload)
}

// Log returns the session-scoped logger (for use in handlers).
func (s *Session) Log() *zap.Logger { return s.log }

// RemoteAddr returns the remote address as a string.
func (s *Session) RemoteAddr() string { return s.conn.RemoteAddr().String() }

// SetState updates the session state machine.
func (s *Session) SetState(st State) {
	var name string
	switch st {
	case StateConnected:
		name = "connected"
	case StateAuthed:
		name = "authed"
	case StateRedirected:
		name = "redirected"
	default:
		name = fmt.Sprintf("unknown(%d)", st)
	}
	s.log.Info("session state →", zap.String("state", name))
	s.state = st
}

// Close closes the underlying connection.
func (s *Session) Close() {
	s.conn.Close()
}

// run is the session read loop — runs in its own goroutine until the connection closes.
func (s *Session) run() {
	s.log.Info("session started")
	defer func() {
		s.conn.Close()
		s.log.Info("session closed")
	}()

	for {
		op, data, err := s.conn.Recv()
		if err != nil {
			s.log.Warn("recv error", zap.Error(err))
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

		if s.state == StateRedirected {
			s.log.Info("client redirected to game server, closing login session")
			return
		}
	}
}
