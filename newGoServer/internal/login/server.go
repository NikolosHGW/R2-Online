// Package login implements the R2 Online Login Server.
//
// Protocol note: the Login Server sends a 198-byte ConnectionClient (1103) welcome
// challenge immediately on connect (plaintext). After that all traffic remains
// plaintext — no RC4 encryption is applied on the login server.
package login

import (
	"encoding/hex"
	"fmt"

	"go.uber.org/zap"

	"r2server/internal/crypto"
	"r2server/internal/network"
	"r2server/internal/packet/login/send"
	"r2server/internal/packet/opcode"
)

// HandlerFunc is the function signature for all login server packet handlers.
type HandlerFunc func(s *Session, r *network.Reader) error

// Server manages login server TCP connections and packet routing.
type Server struct {
	net      *network.Server
	handlers [65536]HandlerFunc
	log      *zap.Logger
}

// NewServer creates a login server listening on addr.
func NewServer(addr string, log *zap.Logger) *Server {
	srv := &Server{log: log}
	srv.net = network.NewServer(addr, srv.onAccept, log.Named("tcp"))
	return srv
}

// Handle registers a handler for the given opcode.
func (s *Server) Handle(op opcode.Opcode, h HandlerFunc) {
	s.handlers[op] = h
}

// ListenAndServe starts the server. Blocks until error.
func (s *Server) ListenAndServe() error {
	return s.net.ListenAndServe()
}

// onAccept is called in a goroutine for every new TCP connection.
func (s *Server) onAccept(conn *network.Conn) {
	// Server always speaks first: send the 198-byte welcome challenge.
	pkt := &send.ConnectionClient{}
	if err := conn.Send(opcode.ConnectionClient, pkt.Encode()); err != nil {
		s.log.Warn("failed to send ConnectionClient",
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Error(err),
		)
		conn.Close()
		return
	}

	// Client encrypts subsequent packets with RC4 using DecryptSbox.
	// The cipher is reset per-packet (fresh S-box state for each received packet).
	// Server responses are always plaintext.
	conn.SetRecvSbox(crypto.DecryptSbox)

	s.log.Info("client connected, welcome sent", zap.String("remote", conn.RemoteAddr().String()))
	sess := newSession(conn, s)
	sess.run()
}

// dispatch looks up and calls the handler for the given opcode.
func (s *Server) dispatch(sess *Session, op opcode.Opcode, r *network.Reader) error {
	h := s.handlers[op]
	if h == nil {
		if ce := s.log.Check(zap.DebugLevel, "unhandled opcode"); ce != nil {
			// At DEBUG level dump the raw bytes — helps reverse engineer unknown packets
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
