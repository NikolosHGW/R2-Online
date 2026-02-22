package send

import (
	"net"

	"r2server/internal/network"
)

// GameServer describes one entry in the server list sent to the client.
type GameServer struct {
	ServerID   int16
	Name       string
	IP         string // dotted-decimal, e.g. "192.168.1.1"
	Port       uint16
	Type       uint8 // 1 = normal, 2 = open beta
	Hidden     bool
	Status     bool   // true = online
	Congestion uint8  // 0-100
}

// SendServers is opcode 3101 — full server list.
//
// Binary layout per server (119 bytes):
//
//	[Status: uint8] [ServerId: int16] [Name: 101 bytes CP-1251]
//	[Congestion: uint8] [IP: 4 bytes] [Port: uint16 BE] [Type: uint8] [Hidden: uint8] [Pad: 6 bytes]
type SendServers struct {
	Servers []GameServer
}

func (p *SendServers) Encode() []byte {
	w := network.NewWriter()
	w.WriteUint8(uint8(len(p.Servers)))
	for _, s := range p.Servers {
		writeServerEntry(w, s)
	}
	return w.Bytes()
}

// RefreshedServers is opcode 3116 — same layout as SendServers.
type RefreshedServers = SendServers

func writeServerEntry(w *network.Writer, s GameServer) {
	status := uint8(0)
	if s.Status {
		status = 1
	}
	w.WriteUint8(status)
	w.WriteInt16(s.ServerID)
	w.WriteStringCP1251(s.Name, 101)
	w.WriteUint8(s.Congestion)

	ip := net.ParseIP(s.IP).To4()
	if ip == nil {
		ip = []byte{127, 0, 0, 1}
	}
	w.WriteBytes(ip)

	w.WriteUint16BE(s.Port) // big-endian per protocol
	w.WriteUint8(s.Type)

	hidden := uint8(0)
	if s.Hidden {
		hidden = 1
	}
	w.WriteUint8(hidden)
	w.WriteZero(6) // padding
}
