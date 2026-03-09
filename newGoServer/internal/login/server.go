// Package login implements the R2 Online Login Server.
package login

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	// Optional diagnostics for key offset discovery:
	// - WELCOME_KEY_MUTATE_OFFSET=-1|96|153|155
	// - RECV_SBOX_MODE=static|welcome
	// - RECV_SBOX_WELCOME_OFFSET=96
	welcomeMutateOffset int
	recvSboxMode        string
	recvSboxOffset      int
	diagWelcomeSboxes   bool
	diagWelcomeBrute    bool
	diagDropSearch      bool
	diagDropMax         int
	diagDropExpectSeq   int

	recvDropN    int
	recvStreamRC4 bool
}

// NewServer creates a login server listening on addr.
func NewServer(addr string, log *zap.Logger) *Server {
	srv := &Server{
		log:                 log,
		welcomeMutateOffset: getEnvInt("WELCOME_KEY_MUTATE_OFFSET", -1),
		recvSboxMode:        strings.ToLower(strings.TrimSpace(getEnv("RECV_SBOX_MODE", "static"))),
		recvSboxOffset:      getEnvInt("RECV_SBOX_WELCOME_OFFSET", 96),
		diagWelcomeSboxes:   getEnv("DIAG_WELCOME_SBOXES", "") == "1",
		diagWelcomeBrute:    getEnv("DIAG_WELCOME_BRUTE", "") == "1",
		diagDropSearch:      getEnv("DIAG_DROP_SEARCH", "") == "1",
		diagDropMax:         getEnvInt("DIAG_DROP_MAX", 4096),
		diagDropExpectSeq:   getEnvInt("DIAG_DROP_EXPECT_SEQ", 1),
		recvDropN:           getEnvInt("RECV_RC4_DROP", 0),
		recvStreamRC4:       getEnv("RECV_RC4_STREAM", "") == "1",
	}
	if srv.recvSboxMode != "static" && srv.recvSboxMode != "welcome" {
		log.Warn("invalid RECV_SBOX_MODE, fallback to static", zap.String("mode", srv.recvSboxMode))
		srv.recvSboxMode = "static"
	}
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
	payload := send.CloneWelcomeKey()

	if s.welcomeMutateOffset >= 0 {
		key, err := crypto.GenerateSessionKey()
		if err != nil {
			s.log.Warn("failed to generate session key", zap.Error(err))
			conn.Close()
			return
		}
		if err := send.WriteSessionKey(payload, s.welcomeMutateOffset, key); err != nil {
			s.log.Warn("failed to write welcome key segment",
				zap.Int("offset", s.welcomeMutateOffset),
				zap.Error(err),
			)
			conn.Close()
			return
		}
		s.log.Info("welcome key segment mutated",
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Int("offset", s.welcomeMutateOffset),
			zap.String("key18_hex", hex.EncodeToString(key[:])),
		)
	}

	pkt := &send.ConnectionClient{Payload: payload}
	if err := conn.Send(opcode.ConnectionClient, pkt.Encode()); err != nil {
		s.log.Warn("failed to send ConnectionClient",
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Error(err),
		)
		conn.Close()
		return
	}

	switch s.recvSboxMode {
	case "welcome":
		key, err := send.ReadSessionKey(payload, s.recvSboxOffset)
		if err != nil {
			s.log.Warn("failed to read welcome key segment for recv sbox",
				zap.Int("offset", s.recvSboxOffset),
				zap.Error(err),
			)
			conn.Close()
			return
		}
		conn.SetRecvSbox(crypto.KSA(key[:]))
		s.log.Info("recv sbox set from welcome payload",
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Int("offset", s.recvSboxOffset),
			zap.String("key18_hex", hex.EncodeToString(key[:])),
		)
	default:
		conn.SetRecvSbox(crypto.DecryptSbox)
	}
	conn.SetRecvDropN(s.recvDropN)
	conn.SetRecvStreamMode(s.recvStreamRC4)
	if s.recvDropN > 0 || s.recvStreamRC4 {
		s.log.Info("recv rc4 mode configured",
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Int("drop", s.recvDropN),
			zap.Bool("stream", s.recvStreamRC4),
		)
	}

	if s.diagWelcomeSboxes {
		// Compare candidate key offsets by resulting opcode on first encrypted packet.
		conn.AddDiagSbox("static", crypto.DecryptSbox)
		addDiagOffset := func(off int) {
			key, err := send.ReadSessionKey(payload, off)
			if err != nil {
				return
			}
			conn.AddDiagSbox(fmt.Sprintf("wel_%d", off), crypto.KSA(key[:]))
			for name, kb := range deriveKeyVariants(key) {
				if len(kb) == 0 {
					continue
				}
				conn.AddDiagSbox(fmt.Sprintf("wel_%d_%s", off, name), crypto.KSA(kb))
			}
		}
		if s.diagWelcomeBrute {
			for off := 0; off <= len(payload)-18; off++ {
				addDiagOffset(off)
			}
		} else {
			for _, off := range []int{2, 22, 41, 59, 78, 96, 98, 116, 136, 153, 155, 173} {
				addDiagOffset(off)
			}
			// Raw variable-length windows around key candidates.
			for _, off := range []int{96, 153, 155} {
				for ln := 8; ln <= 40; ln++ {
					if off+ln > len(payload) {
						break
					}
					k := make([]byte, ln)
					copy(k, payload[off:off+ln])
					conn.AddDiagSbox(fmt.Sprintf("wel_%d_len_%d", off, ln), crypto.KSA(k))
				}
			}
		}
		for name, kb := range derivePayloadVariants(payload) {
			if len(kb) == 0 {
				continue
			}
			conn.AddDiagSbox(name, crypto.KSA(kb))
		}
		s.log.Info("diagnostic sbox trial enabled",
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Bool("brute", s.diagWelcomeBrute),
			zap.Bool("drop_search", s.diagDropSearch),
			zap.Int("drop_max", s.diagDropMax),
		)
	}
	if s.diagDropSearch {
		// login opcode after decrypt should be 0x0C1C
		conn.SetDiagDropSearch(0x0C1C, s.diagDropMax, s.diagDropExpectSeq)
		if !s.diagWelcomeSboxes {
			// Ensure at least one candidate is logged.
			conn.AddDiagSbox("static", crypto.DecryptSbox)
		}
	}

	s.log.Info("client connected, welcome sent", zap.String("remote", conn.RemoteAddr().String()))
	sess := newSession(conn, s)
	sess.run()
}

// dispatch looks up and calls the handler for the given opcode.
func (s *Server) dispatch(sess *Session, op opcode.Opcode, r *network.Reader) error {
	h := s.handlers[op]
	if h == nil {
		if ce := s.log.Check(zap.DebugLevel, "unhandled opcode"); ce != nil {
			// At DEBUG level dump raw payload for protocol reversing.
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func deriveKeyVariants(key [18]byte) map[string][]byte {
	out := map[string][]byte{}

	// swap byte pairs: [a b c d] -> [b a d c]
	{
		k := make([]byte, 18)
		for i := 0; i < 18; i += 2 {
			k[i], k[i+1] = key[i+1], key[i]
		}
		out["swap2"] = k
	}

	// low/high bytes of 9 uint16 words
	{
		lo := make([]byte, 9)
		hi := make([]byte, 9)
		for i := 0; i < 9; i++ {
			lo[i] = key[i*2]
			hi[i] = key[i*2+1]
		}
		out["u16_lo9"] = lo
		out["u16_hi9"] = hi
	}

	// each uint16 expanded to uint32 LE (36 bytes)
	{
		k := make([]byte, 0, 36)
		var tmp [4]byte
		for i := 0; i < 9; i++ {
			v := binary.LittleEndian.Uint16(key[i*2 : i*2+2])
			binary.LittleEndian.PutUint32(tmp[:], uint32(v))
			k = append(k, tmp[:]...)
		}
		out["u16_to_u32le_36"] = k
	}

	// each uint16 expanded to uint32 with swapped word bytes first.
	{
		k := make([]byte, 0, 36)
		var tmp [4]byte
		for i := 0; i < 9; i++ {
			v := uint16(key[i*2+1]) | (uint16(key[i*2]) << 8)
			binary.LittleEndian.PutUint32(tmp[:], uint32(v))
			k = append(k, tmp[:]...)
		}
		out["u16be_to_u32le_36"] = k
	}

	return out
}

func derivePayloadVariants(payload []byte) map[string][]byte {
	out := map[string][]byte{}
	if len(payload) == 0 {
		return out
	}
	cp := func(name string, b []byte) {
		k := make([]byte, len(b))
		copy(k, b)
		out[name] = k
	}

	cp("wel_full_198", payload)
	if len(payload) > 2 {
		cp("wel_full_skip2", payload[2:])
	}
	if len(payload) >= 192 {
		cp("wel_first_192", payload[:192])
		cp("wel_2_194", payload[2:194])
	}

	even := make([]byte, 0, (len(payload)+1)/2)
	odd := make([]byte, 0, len(payload)/2)
	for i := 0; i < len(payload); i++ {
		if i%2 == 0 {
			even = append(even, payload[i])
		} else {
			odd = append(odd, payload[i])
		}
	}
	cp("wel_even_bytes", even)
	cp("wel_odd_bytes", odd)

	rev := make([]byte, len(payload))
	for i := 0; i < len(payload); i++ {
		rev[i] = payload[len(payload)-1-i]
	}
	cp("wel_full_rev", rev)

	return out
}
