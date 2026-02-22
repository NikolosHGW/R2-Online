package network

import (
	"net"

	"go.uber.org/zap"
)

// OnAcceptFunc is called in a new goroutine for every accepted TCP connection.
// Implementations should block until the session is done (i.e. run the read loop).
type OnAcceptFunc func(conn *Conn)

// Server is a generic TCP server. It knows nothing about packets or game logic —
// it only accepts connections and hands them to the provided callback.
type Server struct {
	addr     string
	onAccept OnAcceptFunc
	log      *zap.Logger
}

func NewServer(addr string, onAccept OnAcceptFunc, log *zap.Logger) *Server {
	return &Server{addr: addr, onAccept: onAccept, log: log}
}

// ListenAndServe starts listening and blocks until an unrecoverable error occurs.
func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.log.Info("server listening", zap.String("addr", l.Addr().String()))

	for {
		rawConn, err := l.Accept()
		if err != nil {
			s.log.Error("accept error", zap.Error(err))
			continue
		}
		s.log.Info("new connection", zap.String("remote", rawConn.RemoteAddr().String()))
		go s.onAccept(NewConn(rawConn))
	}
}
