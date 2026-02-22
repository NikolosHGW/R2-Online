// Package game implements the R2 Online Game Server.
//
// Responsibilities:
//   - Accept TCP connections redirected from the login server
//   - Perform key exchange (ConnectionClient, opcode 1103)
//   - Validate session tokens (opcode 5100 LoginUserReq → Redis)
//   - Serve the character list and handle character select
//   - Route in-game packets (movement, combat, inventory, …) to handlers
//
// Inject your repository implementations via Register* methods.
// The server itself contains no game logic — only the routing plumbing.
package game

import (
	"encoding/hex"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"r2server/internal/network"
	"r2server/internal/packet/game/send"
	loginsend "r2server/internal/packet/login/send"
	"r2server/internal/packet/opcode"
)

// HandlerFunc is the function signature for all game server packet handlers.
type HandlerFunc func(s *Session, r *network.Reader) error

// Server manages game server TCP connections and packet routing.
type Server struct {
	net      *network.Server
	handlers [65536]HandlerFunc
	log      *zap.Logger

	// onEnd is called when a session disconnects (set via OnSessionEnd)
	onEnd func(s *Session)

	mu       sync.RWMutex
	sessions map[int32]*Session // accountID → Session
}

// NewServer creates a game server listening on addr.
func NewServer(addr string, log *zap.Logger) *Server {
	srv := &Server{
		sessions: make(map[int32]*Session),
		log:      log,
	}
	srv.net = network.NewServer(addr, srv.onAccept, log.Named("tcp"))
	return srv
}

// Handle registers a packet handler for the given opcode.
func (s *Server) Handle(op opcode.Opcode, h HandlerFunc) {
	s.handlers[op] = h
}

// OnSessionEnd registers a callback invoked when any session disconnects.
// Use this to remove the player from the game world.
func (s *Server) OnSessionEnd(fn func(sess *Session)) {
	s.onEnd = fn
}

// ListenAndServe starts the server. Blocks until error.
func (s *Server) ListenAndServe() error {
	return s.net.ListenAndServe()
}

// onAccept is called in a goroutine for every new TCP connection.
func (s *Server) onAccept(conn *network.Conn) {
	sess := newSession(conn, s)

	// Send the 198-byte GameGuard welcome challenge (hardcoded, same as login server).
	// The RC4 S-box is seeded from the same key padded to 256 bytes with zeros.
	// Both sides know this key statically, so they arrive at the same cipher state.
	pkt := &send.ConnectionClient{}
	if err := conn.Send(opcode.ConnectionClient, pkt.Encode()); err != nil {
		s.log.Warn("failed to send ConnectionClient", zap.Error(err))
		conn.Close()
		return
	}

	var sbox [256]byte
	copy(sbox[:], loginsend.WelcomeKey[:])
	conn.SetCipher(sbox)
	s.log.Info("key exchange complete", zap.String("remote", conn.RemoteAddr().String()))

	// Block until the session ends.
	sess.run()
}

// dispatch looks up and calls the handler for the given opcode.
func (s *Server) dispatch(sess *Session, op opcode.Opcode, r *network.Reader) error {
	h := s.handlers[op]
	if h == nil {
		if ce := s.log.Check(zap.DebugLevel, "unhandled opcode"); ce != nil {
			raw, _ := r.ReadBytes(r.Remaining())
			ce.Write(
				zap.String("opcode", fmt.Sprintf("0x%04X (%d)", op, op)),
				zap.Int("payload_len", len(raw)),
				zap.String("payload_hex", hex.EncodeToString(raw)),
			)
		} else {
			s.log.Warn("unhandled opcode",
				zap.String("opcode", fmt.Sprintf("0x%04X (%d)", op, op)),
			)
		}
		return nil
	}
	s.log.Debug("dispatch", zap.String("opcode", fmt.Sprintf("0x%04X", op)))
	return h(sess, r)
}

// AddSession registers a session by accountID (called after LoginUserReq validates).
func (s *Server) AddSession(accountID int32, sess *Session) {
	s.mu.Lock()
	s.sessions[accountID] = sess
	s.mu.Unlock()
}

// onSessionEnd is called by Session.run() when the connection closes.
func (s *Server) onSessionEnd(sess *Session) {
	if sess.AccountID != 0 {
		s.mu.Lock()
		delete(s.sessions, sess.AccountID)
		s.mu.Unlock()
	}
	if s.onEnd != nil {
		s.onEnd(sess)
	}
}

// GetSession returns the active session for an account (or nil).
func (s *Server) GetSession(accountID int32) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[accountID]
}
