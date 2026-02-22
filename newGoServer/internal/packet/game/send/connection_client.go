// Package send contains outgoing (Server→Client) packet encoders for the Game Server.
package send

import (
	loginsend "r2server/internal/packet/login/send"
	"r2server/internal/network"
)

// ConnectionClient is opcode 1103 — same 198-byte welcome key as the login server.
// Reuses the hardcoded WelcomeKey from the login send package.
type ConnectionClient struct{}

func (p *ConnectionClient) Encode() []byte {
	w := network.NewWriter()
	w.WriteBytes(loginsend.WelcomeKey[:])
	return w.Bytes()
}
