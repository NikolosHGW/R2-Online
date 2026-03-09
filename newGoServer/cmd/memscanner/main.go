//go:build windows

// memscanner scans a running R2 Online client process for RC4 key material.
//
// It answers three questions:
//  1. Is crypto.DecryptSbox hardcoded in the client binary?
//  2. Is there a live CRc4A::SState (S-box as int32[256]) in memory, and what are its values?
//  3. Where does the client keep the __mKey8 / WelcomeKey data?
//
// Usage:
//
//	go run ./cmd/memscanner -proc r2.exe
//	go run ./cmd/memscanner -proc r2.exe -dump 0x7FF612340000  (decode 1024 bytes at addr as SState)
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ── Patterns ──────────────────────────────────────────────────────────────────

// decryptSbox is the hardcoded S-box from crypto/sbox.go — 256 raw bytes.
var decryptSbox = [256]byte{
	0x90, 0x9A, 0xE2, 0xF4, 0x51, 0xBB, 0xB2, 0x13, 0xD6, 0x48, 0x0E, 0xE3, 0x59, 0x04, 0x07, 0x03,
	0xDA, 0x19, 0x47, 0xCF, 0x81, 0xA4, 0x41, 0x37, 0x40, 0xAB, 0xA6, 0xDC, 0xE1, 0x0A, 0x63, 0x4D,
	0x20, 0x53, 0xFD, 0x15, 0xFB, 0x11, 0xF3, 0x79, 0xA1, 0x10, 0xF5, 0x58, 0x38, 0x5C, 0x69, 0x0B,
	0xC6, 0x4A, 0x5A, 0x6E, 0x72, 0x9B, 0x87, 0x1C, 0x7E, 0x82, 0xF8, 0x71, 0x62, 0x14, 0x6A, 0x39,
	0xAF, 0x73, 0x30, 0x86, 0x61, 0x93, 0xB8, 0x05, 0x92, 0x9C, 0x77, 0xE9, 0x6C, 0x0F, 0x2B, 0x89,
	0xDB, 0x6D, 0xA8, 0xA3, 0x24, 0x12, 0xB5, 0x4C, 0x97, 0x02, 0xCE, 0x88, 0x57, 0xDD, 0xBE, 0x8A,
	0x50, 0x6F, 0x7A, 0x2D, 0x8C, 0x3C, 0x22, 0x9F, 0xFA, 0x3E, 0xD3, 0x52, 0xCC, 0x91, 0xC0, 0x31,
	0x08, 0xD0, 0x74, 0xB3, 0x43, 0x46, 0x2C, 0x4B, 0x95, 0x16, 0x9E, 0xB6, 0xB9, 0x00, 0x5F, 0xB0,
	0x1F, 0x8F, 0x25, 0xA5, 0xAC, 0xC7, 0xC4, 0xBC, 0x83, 0x45, 0x99, 0x5B, 0xA2, 0xFC, 0x34, 0xED,
	0x6B, 0x7C, 0xEA, 0xF1, 0xAD, 0x27, 0xFF, 0xB4, 0x26, 0x5D, 0xC5, 0x7B, 0x56, 0xB7, 0xE6, 0xD7,
	0x67, 0xA7, 0x1E, 0x60, 0xC8, 0xA0, 0x80, 0x3F, 0x4F, 0x98, 0x2E, 0x8B, 0x5E, 0x21, 0xEB, 0x49,
	0xCD, 0x0C, 0x3D, 0x1D, 0xBD, 0xD1, 0x64, 0xCA, 0x9D, 0xE8, 0x28, 0xC9, 0xD9, 0x01, 0xBF, 0xC3,
	0xE5, 0xE7, 0x06, 0x96, 0x3A, 0x29, 0x8E, 0x42, 0xF9, 0x8D, 0x94, 0x17, 0x32, 0xDF, 0x36, 0x1B,
	0xCB, 0x7F, 0x1A, 0x33, 0x84, 0x2A, 0x44, 0xF7, 0x0D, 0x7D, 0xE4, 0x35, 0xEC, 0x68, 0x4E, 0xF6,
	0xF0, 0x66, 0x3B, 0x70, 0xE0, 0xA9, 0xD4, 0x76, 0x18, 0xD5, 0x09, 0x2F, 0xD2, 0xC1, 0xDE, 0xC2,
	0x85, 0xB1, 0xF2, 0xEE, 0x54, 0xFE, 0xAE, 0xD8, 0x78, 0x55, 0xBA, 0x23, 0x65, 0xEF, 0x75, 0xAA,
}

// sboxAsInt32 encodes a 256-byte S-box as 256 little-endian int32 values
// (the in-memory format of CRc4A::SState::mM).
func sboxAsInt32(s [256]byte) []byte {
	out := make([]byte, 256*4)
	for i, b := range s {
		out[i*4] = b // low byte = the value; upper 3 bytes = 0
	}
	return out
}

// mKey8 — __mKey8 field from WelcomeKey[96:114] (the key that affects encryption per screenshot).
var mKey8 = []byte{
	0xc8, 0x07, 0x11, 0x4e, 0xa9, 0x35, 0xa9, 0x02,
	0xe3, 0x19, 0xb1, 0x7e, 0xb3, 0x59, 0xeb, 0x53,
	0xfe, 0x76,
}

// welcomeKeyHeader — first 16 bytes of the 198-byte ConnectionClient payload.
var welcomeKeyHeader = []byte{
	0xd5, 0x49, 0x82, 0x55, 0x1d, 0x1a, 0x17, 0x2d,
	0xbb, 0x4a, 0x45, 0x43, 0xb7, 0x25, 0xe2, 0x18,
}

// ── CLI ───────────────────────────────────────────────────────────────────────

var (
	procName = flag.String("proc", "r2.exe", "target process name (e.g. r2.exe)")
	dumpAddr = flag.String("dump", "", "hex address to dump as CRc4A::SState (e.g. 0x7FF600010000)")
)

func main() {
	flag.Parse()

	pid, err := findPID(*procName)
	if err != nil {
		fatalf("process %q not found — is the client running?\n  %v", *procName, err)
	}
	printf("process %q → PID %d", *procName, pid)

	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ,
		false, pid,
	)
	if err != nil {
		fatalf("OpenProcess: %v\n(try running as Administrator)", err)
	}
	defer windows.CloseHandle(handle)

	// -dump mode: decode a specific address as CRc4A::SState
	if *dumpAddr != "" {
		var addr uintptr
		fmt.Sscanf(*dumpAddr, "0x%X", &addr)
		if addr == 0 {
			fmt.Sscanf(*dumpAddr, "%d", &addr)
		}
		if addr == 0 {
			fatalf("invalid address: %s", *dumpAddr)
		}
		dumpSState(handle, addr)
		return
	}

	// Normal scan mode
	type pattern struct {
		name string
		sig  []byte // signature to search for (first N bytes — must be unique enough)
		note string
	}

	int32sig := sboxAsInt32(decryptSbox)

	patterns := []pattern{
		{
			"DecryptSbox (raw bytes, 256 B)",
			decryptSbox[:],
			"hardcoded in binary/data segment — confirms static cipher key",
		},
		{
			"DecryptSbox as int32[256] — CRc4A::SState",
			int32sig,
			"live cipher state in memory using DecryptSbox — confirms what key is active NOW",
		},
		{
			"__mKey8 bytes (18 B)",
			mKey8,
			"where key-exchange field from WelcomeKey is stored in client memory",
		},
		{
			"WelcomeKey header (16 B)",
			welcomeKeyHeader,
			"where the 198-byte ConnectionClient payload is kept in memory",
		},
	}

	results := make(map[string][]uintptr, len(patterns))
	scanned := 0

	err = walkReadableMemory(handle, func(baseAddr uintptr, chunk []byte) {
		scanned++
		for _, p := range patterns {
			off := 0
			for {
				idx := bytes.Index(chunk[off:], p.sig)
				if idx < 0 {
					break
				}
				results[p.name] = append(results[p.name], baseAddr+uintptr(off+idx))
				off += idx + len(p.sig)
			}
		}
	})
	if err != nil {
		printf("warn: walk error: %v", err)
	}

	printf("scanned %d readable memory regions", scanned)
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Println("  RESULTS")
	fmt.Println("══════════════════════════════════════════════════════════")

	for _, p := range patterns {
		addrs := results[p.name]
		fmt.Printf("\n[%s]\n", p.name)
		fmt.Printf("  %s\n", p.note)
		if len(addrs) == 0 {
			fmt.Println("  → NOT FOUND")
		} else {
			fmt.Printf("  → FOUND at %d address(es):\n", len(addrs))
			for _, a := range addrs {
				fmt.Printf("      0x%016X\n", a)
				// For SState hits, also try to decode the surrounding context
				if strings.Contains(p.name, "SState") {
					// The SState is { int mX; int mY; int mM[256] }
					// mM starts at offset +8 (after mX and mY)
					// Our signature matched mM[0], so mX is at a-8, mY at a-4
					mxAddr := a - 8
					if mxAddr > a { // underflow guard
						continue
					}
					buf := make([]byte, 8+256*4)
					var n uintptr
					if err := windows.ReadProcessMemory(handle, mxAddr, &buf[0], uintptr(len(buf)), &n); err == nil && n == uintptr(len(buf)) {
						mX := int32(binary.LittleEndian.Uint32(buf[0:4]))
						mY := int32(binary.LittleEndian.Uint32(buf[4:8]))
						fmt.Printf("          mX=%d  mY=%d  (cipher position)\n", mX, mY)
						decoded := make([]byte, 256)
						for i := 0; i < 256; i++ {
							decoded[i] = buf[8+i*4]
						}
						fmt.Printf("          S-box first 32: %s\n", hex.EncodeToString(decoded[:32]))
						mismatches := 0
						for i := range decoded {
							if decoded[i] != decryptSbox[i] {
								mismatches++
							}
						}
						if mismatches == 0 {
							fmt.Println("          ✓ MATCHES DecryptSbox exactly")
						} else {
							fmt.Printf("          ✗ differs from DecryptSbox in %d/256 bytes\n", mismatches)
						}
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Println("INTERPRETATION:")
	sboxRaw := results["DecryptSbox (raw bytes, 256 B)"]
	sstateLive := results["DecryptSbox as int32[256] — CRc4A::SState"]
	switch {
	case len(sboxRaw) > 0 && len(sstateLive) > 0:
		fmt.Println("  DecryptSbox found in BOTH binary AND live memory.")
		fmt.Println("  → Cipher key is HARDCODED. Per-session encryption is NOT possible without patching the client.")
	case len(sboxRaw) > 0 && len(sstateLive) == 0:
		fmt.Println("  DecryptSbox found in binary (data/bss) but NOT in an active SState.")
		fmt.Println("  → Key is hardcoded; no active cipher session found at this moment.")
	case len(sboxRaw) == 0 && len(sstateLive) > 0:
		fmt.Println("  DecryptSbox found ONLY in live SState (not in binary).")
		fmt.Println("  → Possibly computed at runtime. Per-session MAY be possible.")
	case len(sboxRaw) == 0 && len(sstateLive) == 0:
		fmt.Println("  DecryptSbox NOT found anywhere.")
		fmt.Println("  → Either the client uses a different sbox, OR the sbox is computed dynamically.")
		fmt.Println("  → Run this tool WHILE the client is actively connected to the server.")
	}
	fmt.Println()
}

// dumpSState reads 8+256*4 bytes at addr, interprets them as CRc4A::SState,
// and prints the decoded S-box along with mX and mY.
func dumpSState(handle windows.Handle, addr uintptr) {
	// SState = { int mX; int mY; int mM[256]; }
	// Try with mX/mY prefix first, then without
	buf := make([]byte, 8+256*4)
	var n uintptr
	if err := windows.ReadProcessMemory(handle, addr, &buf[0], uintptr(len(buf)), &n); err != nil || n < uintptr(len(buf)) {
		fatalf("ReadProcessMemory at 0x%X: %v (read %d)", addr, err, n)
	}
	mX := int32(binary.LittleEndian.Uint32(buf[0:4]))
	mY := int32(binary.LittleEndian.Uint32(buf[4:8]))
	decoded := make([]byte, 256)
	for i := 0; i < 256; i++ {
		decoded[i] = buf[8+i*4]
	}

	printf("SState at 0x%016X: mX=%d  mY=%d", addr, mX, mY)
	fmt.Println("S-box (256 bytes):")
	for i := 0; i < 256; i += 16 {
		fmt.Printf("  [%3d] %s\n", i, hex.EncodeToString(decoded[i:i+16]))
	}

	matches := 0
	for i, b := range decoded {
		if b == decryptSbox[i] {
			matches++
		}
	}
	fmt.Printf("\nMatches DecryptSbox: %d/256\n", matches)
	if matches == 256 {
		fmt.Println("→ EXACT MATCH: this SState uses the hardcoded DecryptSbox")
	}
}

// ── Windows helpers ──────────────────────────────────────────────────────────

const (
	memCommit            = 0x1000
	pageReadOnly         = 0x02
	pageReadWrite        = 0x04
	pageWriteCopy        = 0x08
	pageExecuteRead      = 0x20
	pageExecuteReadWrite = 0x40
	pageExecuteWriteCopy = 0x80
)

func isReadable(protect uint32) bool {
	core := protect & 0xFF // strip guard/nocache modifiers
	return core&(pageReadOnly|pageReadWrite|pageWriteCopy|
		pageExecuteRead|pageExecuteReadWrite|pageExecuteWriteCopy) != 0
}

const readChunk = 4 << 20 // 4 MB

func walkReadableMemory(handle windows.Handle, fn func(base uintptr, chunk []byte)) error {
	var addr uintptr
	var mbi windows.MemoryBasicInformation
	for {
		if err := windows.VirtualQueryEx(handle, addr, &mbi, unsafe.Sizeof(mbi)); err != nil {
			break
		}
		if mbi.State == memCommit && isReadable(mbi.Protect) {
			for off := uintptr(0); off < mbi.RegionSize; off += readChunk {
				sz := mbi.RegionSize - off
				if sz > readChunk {
					sz = readChunk
				}
				buf := make([]byte, sz)
				var n uintptr
				if err := windows.ReadProcessMemory(handle, mbi.BaseAddress+off, &buf[0], sz, &n); err == nil && n > 0 {
					fn(mbi.BaseAddress+off, buf[:n])
				}
			}
		}
		addr = mbi.BaseAddress + mbi.RegionSize
	}
	return nil
}

func findPID(name string) (uint32, error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snap)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := windows.Process32First(snap, &pe); err != nil {
		return 0, err
	}
	for {
		if strings.EqualFold(windows.UTF16ToString(pe.ExeFile[:]), name) {
			return pe.ProcessID, nil
		}
		if err := windows.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return 0, fmt.Errorf("not found")
}

func printf(f string, a ...any) { fmt.Printf("[+] "+f+"\n", a...) }
func fatalf(f string, a ...any) { fmt.Fprintf(os.Stderr, "[-] "+f+"\n", a...); os.Exit(1) }
