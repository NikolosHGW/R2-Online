package send

import "r2server/internal/network"

// SelectedServer is opcode 3121 — confirms the server choice and provides the
// game server connection details. The client will disconnect from the login
// server and connect to the given IP:Port.
//
// Binary layout: [Pad: 4 bytes]
// The actual IP:Port is embedded in the preceding SelectedServer ACK at the
// transport level; the payload here is just the 4-byte confirmation padding
// as observed in the C# emulator.
type SelectedServer struct{}

func (p *SelectedServer) Encode() []byte {
	w := network.NewWriter()
	w.WriteZero(4)
	return w.Bytes()
}
