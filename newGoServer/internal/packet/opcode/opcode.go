// Package opcode contains all R2 Online packet opcodes as typed constants.
// Sources: ChannelW.h (EFnlAppCTr enum) + FieldW.h + C# emulator PacketType.cs.
package opcode

// Opcode is a 16-bit packet identifier.
type Opcode = uint16

const (
	// ─── Common (both servers) ────────────────────────────────────────────────
	ConnectionClient = Opcode(1103) // S→C  key exchange (1103 = 0x044F)
	GameServerError  = Opcode(1102) // S→C  error response

	// ─── Chat / Channel (EFnlAppCTr range 0x7D0-0x851 = 2000-2129) ──────────
	ChatReq           = Opcode(2033) // C→S  0x7F1 eCTrChatReq
	ChatAck           = Opcode(2034) // S→C  0x7F2 eCTrChatAck
	GameConfiguration = Opcode(2012) // S→C  0x7DC eCTrContentsAck

	// ─── Login Server ─────────────────────────────────────────────────────────
	AuthorizationLogin = Opcode(3100) // C→S  login + password (obfuscated offsets)
	SendServers        = Opcode(3101) // S→C  server list
	LoginServerError   = Opcode(3102) // S→C  login error
	RefreshServers     = Opcode(3115) // C→S  refresh server list
	RefreshedServers   = Opcode(3116) // S→C  updated server list
	SelectServer       = Opcode(3120) // C→S  pick a game server
	SelectedServer     = Opcode(3121) // S→C  confirmation + redirect IP:Port

	// ─── Game Server: session / character select ──────────────────────────────
	LoginUserReq            = Opcode(5100) // C→S  enter game (account id + session token)
	InformationCharacter    = Opcode(5101) // S→C  character list with equipment
	DisplayedCharacter      = Opcode(5103) // S→C  character appearance data
	EnteredMonAck           = Opcode(5104) // S→C  monster entered view
	EnteredItemAck          = Opcode(5105) // S→C  ground item entered view
	ExistedPcAck            = Opcode(5107) // S→C  player left view
	ExistedMonAck           = Opcode(5108) // S→C  monster left view
	ExistedItemAck          = Opcode(5110) // S→C  ground item left view
	ExitMapGbjAck           = Opcode(5114) // S→C  object exited map
	LogoutPcReq             = Opcode(5115) // C→S  logout request
	ChoosePcReq             = Opcode(5116) // C→S  select character slot
	CompleteEnterWorld      = Opcode(5117) // S→C  world entry complete
	CreatePcReq             = Opcode(5118) // C→S  create character
	CompleteCreateCharacter = Opcode(5119) // S→C  character created
	DeletePcReq             = Opcode(5120) // C→S  delete character
	CompleteDeleteCharacter = Opcode(5121) // S→C  character deleted

	// ─── Game Server: inventory / equipment ───────────────────────────────────
	EquipReq    = Opcode(5128) // C→S
	EquipAckAll = Opcode(5129) // S→C
	UnEquipReq  = Opcode(5130) // C→S
	UnEquipAckAll = Opcode(5131) // S→C
	ItemUseReq  = Opcode(5158) // C→S
	ItemDropReq = Opcode(5159) // C→S
	ItemPickupReq = Opcode(5177) // C→S
	ItemAddAck  = Opcode(5232) // S→C
	ItemRemoveAck = Opcode(5233) // S→C
	ItemChangeAck = Opcode(5237) // S→C
	ItemUseAck  = Opcode(5653) // S→C
	ItemCooldown = Opcode(5654) // S→C

	// ─── Game Server: combat ──────────────────────────────────────────────────
	AttackAck      = Opcode(5132) // S→C
	AttackReq      = Opcode(5133) // C→S
	AttackStopAck  = Opcode(5134) // S→C
	DeadAck        = Opcode(5137) // S→C
	AbnormalAck    = Opcode(5160) // S→C
	AbnormaleReleaseAck = Opcode(5161) // S→C
	AbnormalEndReq = Opcode(5162) // C→S
	AbnormalRemoveReq = Opcode(5163) // C→S
	TransformAck   = Opcode(5179) // S→C
	UseSkillPackReq = Opcode(5784) // C→S
	UseSkillPackAck = Opcode(5792) // S→C

	// ─── Game Server: progression ─────────────────────────────────────────────
	InfoExpAck   = Opcode(5139) // S→C
	LevelUpAck   = Opcode(5140) // S→C
	RespawnReq   = Opcode(5141) // C→S
	RespawnAck   = Opcode(5142) // S→C

	// ─── Game Server: stats / UI ──────────────────────────────────────────────
	InventoryCharacteristic  = Opcode(5145) // S→C
	HealthPointCharacteristic = Opcode(5146) // S→C
	SpeedCharacteristic      = Opcode(5147) // S→C
	InfoWeightAck            = Opcode(5149) // S→C
	InfoStomachAck           = Opcode(5173) // S→C
	GossipAck                = Opcode(5212) // S→C
	ServerTime               = Opcode(5652) // S→C
	CheckNeedMoney           = Opcode(5813) // S→C

	// ─── Game Server: movement ────────────────────────────────────────────────
	DoMoveReq       = Opcode(5188) // C→S
	MovedCharacter  = Opcode(5189) // S→C
	DoMoveToAck     = Opcode(5190) // S→C
	CharJumpReq     = Opcode(5192) // C→S
	JumpEndCharacter = Opcode(5193) // S→C
	CharDirReq      = Opcode(5194) // C→S
	CharDirAck      = Opcode(5195) // S→C
	StopMoveCharacter = Opcode(5326) // S→C

	// ─── Game Server: NPC / script ────────────────────────────────────────────
	ScriptReq      = Opcode(5151) // C→S
	ScriptProcReq  = Opcode(5152) // C→S
	MerchantListAck = Opcode(5271) // S→C
	MerchantBuyReq = Opcode(5273) // C→S
	ScrDialogNoMsg2Ack = Opcode(5902) // S→C

	// ─── Game Server: reinforcement ───────────────────────────────────────────
	ReinforceReq  = Opcode(5168) // C→S
	ReinforceAck  = Opcode(5169) // S→C
	ReinforceNak1 = Opcode(5170) // S→C

	// ─── Game Server: social / chat ───────────────────────────────────────────
	GlobalChatReq = Opcode(5225) // C→S
	GlobalChatAck = Opcode(5226) // S→C
	EmoticonReq   = Opcode(5834) // C→S
	EmoticonAck   = Opcode(5835) // S→C

	// ─── Game Server: misc ────────────────────────────────────────────────────
	ChaosBattleLogin      = Opcode(5662) // S→C
	TeleportCenterTownApply = Opcode(5929) // S→C
)
