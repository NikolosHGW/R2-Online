package send

import "r2server/internal/network"

// CharacterSlot holds data for one character slot in the character select screen.
// Empty slots are represented by zero values (CharacterID == 0 means empty).
type CharacterSlot struct {
	CharacterID int32
	Class       uint8
	Gender      uint8
	Head        uint8
	Face        uint8
	Level       int16
	Name        string // max 15 chars

	// Position (used to show which map the character is on)
	PosX float32
	PosY float32
	PosZ float32

	// Basic stats shown in character select
	Strength     int32
	Intelligence int32
	Dexterity    int32
	Chaotic      int32

	// Equipped item IDs for visual preview (0 = empty slot)
	WeaponItemID int32
	ShieldItemID int32
	ArmorItemID  int32
	HelmetItemID int32
	GlovesItemID int32
	BootsItemID  int32
	Belt1ItemID  int32
	CloakItemID  int32
	Ring1ItemID  int32
	Ring2ItemID  int32
	NecklaceItemID int32
}

// InformationCharacter is opcode 5101 — the character list sent after LoginUserReq.
//
// The protocol supports exactly 3 character slots. Empty slots are zero-filled
// 144-byte blocks. The full structure is complex (1522 bytes total); see the
// C# emulator's InformationCharacter parser for the complete field map.
type InformationCharacter struct {
	Characters [3]CharacterSlot
}

// charBlockSize is the size of one character info block in the packet.
const charBlockSize = 144

// equipBlockSize is the size of one character's equipment block.
const equipBlockSize = 320

func (p *InformationCharacter) Encode() []byte {
	w := network.NewWriter()

	w.WriteZero(1) // leading pad

	// --- Character info blocks (3 × 144 bytes) ---
	for _, c := range p.Characters {
		writeCharBlock(w, c)
	}

	// --- Equipment blocks (3 × 320 bytes) ---
	for _, c := range p.Characters {
		writeEquipBlock(w, c)
	}

	// --- Padding (36 bytes = 12 per slot) ---
	w.WriteZero(36)

	// --- Stats (strength, int, dex, chaotic) interleaved per stat across 3 chars ---
	for _, c := range p.Characters {
		w.WriteInt32(c.Strength)
	}
	for _, c := range p.Characters {
		w.WriteInt32(c.Intelligence)
	}
	for _, c := range p.Characters {
		w.WriteInt32(c.Dexterity)
	}
	for _, c := range p.Characters {
		w.WriteInt32(c.Chaotic)
	}

	// --- Positions (X, Z, Y per char — note axis order) ---
	for _, c := range p.Characters {
		w.WriteFloat32(c.PosX)
		w.WriteFloat32(c.PosZ)
		w.WriteFloat32(c.PosY)
	}

	// --- Trailing padding ---
	w.WriteZero(9)

	return w.Bytes()
}

func writeCharBlock(w *network.Writer, c CharacterSlot) {
	if c.CharacterID == 0 {
		// Empty slot — zero-fill the entire block
		w.WriteZero(charBlockSize)
		return
	}

	w.WriteZero(1)  // GuildEmblem flag (reserved)
	w.WriteZero(3)  // align
	w.WriteInt32(c.CharacterID)
	w.WriteUint8(c.Class)
	w.WriteZero(3)  // align
	w.WriteUint8(c.Gender)
	w.WriteUint8(c.Head)
	w.WriteUint8(c.Face)
	w.WriteZero(1)  // Body (reserved)
	w.WriteZero(4)  // GuildNo
	w.WriteZero(4)  // GuildMarkSeq
	w.WriteZero(4)  // GuildGrade
	w.WriteZero(17) // GuildName (reserved)
	w.WriteZero(1)  // IsAtkTower
	w.WriteZero(2)  // DfnsBenefitLv
	w.WriteZero(4)  // DiscipleNo
	w.WriteZero(4)  // DiscipleMemberType
	w.WriteZero(4)  // Hp
	w.WriteZero(4)  // Mp
	w.WriteZero(2)  // Stomach
	w.WriteZero(1)  // StomachStatus
	w.WriteZero(5)  // align
	w.WriteZero(8)  // Exp (reserved)
	w.WriteInt16(c.Level)
	w.WriteStringCP1251(c.Name, 15)
	w.WriteZero(3)  // align
	w.WriteZero(4)  // ChaosBattleSide
	w.WriteZero(2)  // FieldSvrNo
	w.WriteZero(2)  // align
	w.WriteInt32(c.CharacterID) // duplicate
	w.WriteZero(2)  // FieldSvrSeq
	w.WriteZero(1)  // EmblemOfHonorSeq
	w.WriteZero(1)  // align
	w.WriteInt16(c.Level) // duplicate
	w.WriteZero(1)  // NationalFlagNo
	w.WriteZero(1)  // EmblemOfHonorEffectSeq
	w.WriteZero(1)  // TeamRankEffectSeq
	w.WriteZero(3)  // align
	w.WriteZero(4)  // UTGWMatchGroup
	w.WriteZero(4)  // align
	w.WriteZero(8)  // ExpToLevelUp (reserved)
	w.WriteZero(4)  // LastReceiptSection
	// Total so far: 1+3+4+1+3+1+1+1+1+4+4+4+17+1+2+4+4+4+4+2+1+5+8+2+15+3+4+2+4+2+2+1+1+2+1+1+1+3+4+4+8+4 = 144 ✓
}

func writeEquipBlock(w *network.Writer, c CharacterSlot) {
	if c.CharacterID == 0 {
		w.WriteZero(equipBlockSize)
		return
	}

	writeItemRef := func(serialNo uint64, itemID int32) {
		w.WriteUint64(serialNo) // SerialNo (0 = empty)
		w.WriteInt32(itemID)
		w.WriteZero(4) // padding
	}

	// 20 equipment slots × 16 bytes = 320 bytes
	writeItemRef(0, c.WeaponItemID)
	writeItemRef(0, c.ShieldItemID)
	writeItemRef(0, c.ArmorItemID)
	writeItemRef(0, c.Ring1ItemID)
	writeItemRef(0, c.Ring2ItemID)
	writeItemRef(0, c.NecklaceItemID)
	writeItemRef(0, c.BootsItemID)
	writeItemRef(0, c.GlovesItemID)
	writeItemRef(0, c.HelmetItemID)
	writeItemRef(0, c.Belt1ItemID)
	writeItemRef(0, c.CloakItemID)
	writeItemRef(0, 0) // SphereMastery
	writeItemRef(0, 0) // SphereSoul
	writeItemRef(0, 0) // SphereDefense
	writeItemRef(0, 0) // SphereDestruction
	writeItemRef(0, 0) // SphereLife
	writeItemRef(0, 0) // SphereLuck
	writeItemRef(0, 0) // SphereReincarnation
	writeItemRef(0, 0) // SphereCharacteristics
	w.WriteZero(16)    // Servant (reserved)
}
