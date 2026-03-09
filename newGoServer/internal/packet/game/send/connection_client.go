// Package send contains outgoing (Server→Client) packet encoders for the Game Server.
package send

import (
	loginsend "r2server/internal/packet/login/send"
)

// ConnectionClient is opcode 1103 — same 198-byte GameGuard welcome key as the login server.
type ConnectionClient = loginsend.ConnectionClient
