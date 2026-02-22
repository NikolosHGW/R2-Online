// Package recv contains incoming (Client→Server) packet decoders for the Login Server.
package recv

import (
	"fmt"

	"r2server/internal/network"
)

// AuthorizationLogin is opcode 3100 — initial login credentials from the client.
//
// The client deliberately obscures where in the packet the login and password
// strings sit, using two "code" bytes as indices into lookup tables.
// This is a fixed obfuscation scheme (not per-session); both tables are
// hard-coded in the client binary.
type AuthorizationLogin struct {
	Login    string
	Password string
}

// loginOffsets maps (byte_at_256 / 8) → offset of the login string in the packet.
var loginOffsets = [8]int{151, 37, 87, 336, 129, 289, 172, 199}

// passwordOffsets maps (byte_at_81 / 2) → offset of the password string in the packet.
var passwordOffsets = [6]int{260, 60, 108, 4, 314, 357}

const passwordOffsetDefault = 390

func (p *AuthorizationLogin) Decode(r *network.Reader) error {
	// --- Determine password field offset ---
	codePwd, err := r.ByteAt(81)
	if err != nil {
		return fmt.Errorf("AuthorizationLogin: read password code: %w", err)
	}
	pwdIdx := int(codePwd) / 2
	var pwdOffset int
	if pwdIdx < len(passwordOffsets) {
		pwdOffset = passwordOffsets[pwdIdx]
	} else {
		pwdOffset = passwordOffsetDefault
	}

	// --- Determine login field offset ---
	codeLogin, err := r.ByteAt(256)
	if err != nil {
		return fmt.Errorf("AuthorizationLogin: read login code: %w", err)
	}
	loginIdx := int(codeLogin) / 8
	var loginOffset int
	if loginIdx < len(loginOffsets) {
		loginOffset = loginOffsets[loginIdx]
	} else {
		loginOffset = 220 // default from C# emulator
	}

	// --- Read password string (CP-1251, null-terminated) ---
	pwdSlice, err := r.Slice(pwdOffset, 64) // max password length
	if err != nil {
		return fmt.Errorf("AuthorizationLogin: slice password: %w", err)
	}
	pwdReader := network.NewReader(pwdSlice)
	p.Password, err = pwdReader.ReadStringCP1251(0)
	if err != nil {
		return fmt.Errorf("AuthorizationLogin: decode password: %w", err)
	}

	// --- Read login string (CP-1251, null-terminated) ---
	loginSlice, err := r.Slice(loginOffset, 64)
	if err != nil {
		return fmt.Errorf("AuthorizationLogin: slice login: %w", err)
	}
	loginReader := network.NewReader(loginSlice)
	p.Login, err = loginReader.ReadStringCP1251(0)
	if err != nil {
		return fmt.Errorf("AuthorizationLogin: decode login: %w", err)
	}

	return nil
}
