# Packet Examples in Go

This document shows how to read or send each packet defined in the repository. The structures correspond to C# packet models.

Helper types:
```go
type Vec3 struct { X, Y, Z float32 }
type UniqueId uint32
```

## LoginUserReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/5100_LoginUserReqModel.cs
```go
type LoginUserReq struct {
    AccountId uint32
    SessionId int32
    Password string
}

func ReadLoginUserReq(r io.Reader) (LoginUserReq, error) {
    var pkt LoginUserReq
    // TODO: read fields from r
    return pkt, nil
}
```

## LogoutPcReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/5115_LogoutPcReqModel.cs
```go
type LogoutPcReq struct {
}

func ReadLogoutPcReq(r io.Reader) (LogoutPcReq, error) {
    var pkt LogoutPcReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ChoosePcReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/5116_ChoosePcReqModel.cs
```go
type ChoosePcReq struct {
    PcNo uint32
}

func ReadChoosePcReq(r io.Reader) (ChoosePcReq, error) {
    var pkt ChoosePcReq
    // TODO: read fields from r
    return pkt, nil
}
```

## AbnormalRemoveReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Action/5163_AbnormalCancelReqModel.cs
```go
type AbnormalRemoveReq struct {
    Type int32
}

func ReadAbnormalRemoveReq(r io.Reader) (AbnormalRemoveReq, error) {
    var pkt AbnormalRemoveReq
    // TODO: read fields from r
    return pkt, nil
}
```

## UseSkillPackReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Action/5784_UseSkillPackReqModel.cs
```go
type UseSkillPackReq struct {
    SkillId int32
    TargetUniqueId uint32
    IsTeam byte
}

func ReadUseSkillPackReq(r io.Reader) (UseSkillPackReq, error) {
    var pkt UseSkillPackReq
    // TODO: read fields from r
    return pkt, nil
}
```

## AttackReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Attack/5133_AttackReqModel.cs
```go
type AttackReq struct {
    TargetSessionGameId uint32
    AttackType uint16
    AttackPosition Vec3
    AttackFlag byte
}

func ReadAttackReq(r io.Reader) (AttackReq, error) {
    var pkt AttackReq
    // TODO: read fields from r
    return pkt, nil
}
```

## CreatePcReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Character/5118_CreatePcReqModel.cs
```go
type CreatePcReq struct {
    Slot byte
    Class byte
    Sex byte
    Head byte
    Face byte
    TypeBody byte
    Name string
}

func ReadCreatePcReq(r io.Reader) (CreatePcReq, error) {
    var pkt CreatePcReq
    // TODO: read fields from r
    return pkt, nil
}
```

## DeletePcReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Character/5120_DeletePcReqModel.cs
```go
type DeletePcReq struct {
    PcNo uint32
    Slot byte
}

func ReadDeletePcReq(r io.Reader) (DeletePcReq, error) {
    var pkt DeletePcReq
    // TODO: read fields from r
    return pkt, nil
}
```

## RespawnReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Character/5141_RespawnReqModel.cs
```go
type RespawnReq struct {
    IsRespawn byte
}

func ReadRespawnReq(r io.Reader) (RespawnReq, error) {
    var pkt RespawnReq
    // TODO: read fields from r
    return pkt, nil
}
```

## DoMoveReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Character/5188_DoMoveReqModel.cs
```go
type DoMoveReq struct {
    Position Vec3
    Direction float32
    Action uint32
    Flag byte
}

func ReadDoMoveReq(r io.Reader) (DoMoveReq, error) {
    var pkt DoMoveReq
    // TODO: read fields from r
    return pkt, nil
}
```

## CharJumpReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Character/5192_CharJumpReqModel.cs
```go
type CharJumpReq struct {
    Action uint32
    MoveDirection float32
}

func ReadCharJumpReq(r io.Reader) (CharJumpReq, error) {
    var pkt CharJumpReq
    // TODO: read fields from r
    return pkt, nil
}
```

## CharDirReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Character/5194_CharDirReqModel.cs
```go
type CharDirReq struct {
    Direction float32
}

func ReadCharDirReq(r io.Reader) (CharDirReq, error) {
    var pkt CharDirReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ChatReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Chat/2033_ChatReqModel.cs
```go
type ChatReq struct {
    Type byte
    Message string
    Name string
}

func ReadChatReq(r io.Reader) (ChatReq, error) {
    var pkt ChatReq
    // TODO: read fields from r
    return pkt, nil
}
```

## GlobalChatReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Chat/5225_GlobalChatReqModel.cs
```go
type GlobalChatReq struct {
    Message string
}

func ReadGlobalChatReq(r io.Reader) (GlobalChatReq, error) {
    var pkt GlobalChatReq
    // TODO: read fields from r
    return pkt, nil
}
```

## EmoticonReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Chat/5834_EmoticonReqModel.cs
```go
type EmoticonReq struct {
    Type int32
}

func ReadEmoticonReq(r io.Reader) (EmoticonReq, error) {
    var pkt EmoticonReq
    // TODO: read fields from r
    return pkt, nil
}
```

## EquipReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Inventory/5128_EquipReqModel.cs
```go
type EquipReq struct {
    SerialNumber uint64
    Index uint32
}

func ReadEquipReq(r io.Reader) (EquipReq, error) {
    var pkt EquipReq
    // TODO: read fields from r
    return pkt, nil
}
```

## UnEquipReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Inventory/5130_UnEquipReqModel.cs
```go
type UnEquipReq struct {
    Position ItemPositionType
}

func ReadUnEquipReq(r io.Reader) (UnEquipReq, error) {
    var pkt UnEquipReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ItemUseReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Inventory/5158_ItemUseReqModel.cs
```go
type ItemUseReq struct {
    SerialNumber uint64
    ItemId int32
    UniqueIdentifier uint32
    TargetSerialNumber uint64
    IsTeam byte
}

func ReadItemUseReq(r io.Reader) (ItemUseReq, error) {
    var pkt ItemUseReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ItemDropReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Inventory/5159_ItemDropReqModel.cs
```go
type ItemDropReq struct {
    SerialNumber uint64
    Stack uint32
}

func ReadItemDropReq(r io.Reader) (ItemDropReq, error) {
    var pkt ItemDropReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ReinforceReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Inventory/5168_ReinforceReqModel.cs
```go
type ReinforceReq struct {
    SerialNumber uint64
    SerialNumber0 uint64
    SerialNumber1 uint64
    SerialNumber2 uint64
    Count int32
}

func ReadReinforceReq(r io.Reader) (ReinforceReq, error) {
    var pkt ReinforceReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ItemPickupReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Inventory/5177_ItemPickupReqModel.cs
```go
type ItemPickupReq struct {
    UniqueIdentifierItem uint32
}

func ReadItemPickupReq(r io.Reader) (ItemPickupReq, error) {
    var pkt ItemPickupReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ScriptReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Npc/5151_ScriptReqModel.cs
```go
type ScriptReq struct {
    UniqueIdentifier uint32
    Param int32
}

func ReadScriptReq(r io.Reader) (ScriptReq, error) {
    var pkt ScriptReq
    // TODO: read fields from r
    return pkt, nil
}
```

## ScriptProcReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Npc/5152_ScriptProcReqModel.cs
```go
type ScriptProcReq struct {
    UniqueIdentifier uint32
    ScriptAction ScriptAction
    Param int32
}

func ReadScriptProcReq(r io.Reader) (ScriptProcReq, error) {
    var pkt ScriptProcReq
    // TODO: read fields from r
    return pkt, nil
}
```

## MerchantBuyReq (C->S)
// Packets/Packets.Server.Game/Models/Receive/Npc/5273_MerchantBuyReqModel.cs
```go
type MerchantBuyReq struct {
    UniqueIdentifier uint32
    ItemId int32
    Count int32
    ParmA int32
    ParmB int32
    SortKey int32
}

func ReadMerchantBuyReq(r io.Reader) (MerchantBuyReq, error) {
    var pkt MerchantBuyReq
    // TODO: read fields from r
    return pkt, nil
}
```

## GameServerError (S->C)
// Packets/Packets.Server.Game/Models/Send/1102_GameServerErrorModel.cs
```go
type GameServerError struct {
    PacketType PacketType
    ErrorType GameServerErrorType
    IsMsgBox bool
}

func SendGameServerError(w io.Writer, pkt GameServerError) error {
    // TODO: marshal fields and write
    return nil
}
```

## ConnectionClient (S->C)
// Packets/Packets.Server.Game/Models/Send/1103_ConnectionClientModel.cs
```go
type ConnectionClient struct {
    DecryptKey []byte
}

func SendConnectionClient(w io.Writer, pkt ConnectionClient) error {
    // TODO: marshal fields and write
    return nil
}
```

## CompleteEnterWorld (S->C)
// Packets/Packets.Server.Game/Models/Send/5117_CompleteEnterWorldModel.cs
```go
type CompleteEnterWorld struct {
    SessionGameId uint32
    Position Vec3
    Reputation int16
    MoveRate int16
    AttackRate int16
}

func SendCompleteEnterWorld(w io.Writer, pkt CompleteEnterWorld) error {
    // TODO: marshal fields and write
    return nil
}
```

## AbnormalAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Action/5160_AbnormalAckModel.cs
```go
type AbnormalAck struct {
    UniqueIdentifier uint32
    BuffId int32
    EndTick uint32
    Position Vec3
}

func SendAbnormalAck(w io.Writer, pkt AbnormalAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## AbnormaleReleaseAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Action/5161_AbnormalReleaseAckModel.cs
```go
type AbnormaleReleaseAck struct {
    UniqueIdentifier uint32
    Type int32
}

func SendAbnormaleReleaseAck(w io.Writer, pkt AbnormaleReleaseAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## TransformAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Action/5179_TransformAckModel.cs
```go
type TransformAck struct {
    UniqueIdentifier uint32
    MonsterId int32
}

func SendTransformAck(w io.Writer, pkt TransformAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ItemCooldown (S->C)
// Packets/Packets.Server.Game/Models/Send/Action/5654_ItemCooldownAckModel.cs
```go
type ItemCooldown struct {
    ItemId int32
}

func SendItemCooldown(w io.Writer, pkt ItemCooldown) error {
    // TODO: marshal fields and write
    return nil
}
```

## UseSkillPackAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Action/5792_UseSkillPackAckModel.cs
```go
type UseSkillPackAck struct {
    SkillId int32
}

func SendUseSkillPackAck(w io.Writer, pkt UseSkillPackAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## AttackAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Attack/5132_AttackAckModel.cs
```go
type AttackAck struct {
    OffenseSessionGameId uint32
    DefenseSessionGameId uint32
    TypeHit TypeHit
    OffensePosition Vec3
    HpAttacked int16
}

func SendAttackAck(w io.Writer, pkt AttackAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## AttackStopAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Attack/5134_AttackStopAckModel.cs
```go
type AttackStopAck struct {
    OffenceSesionGameId uint32
}

func SendAttackStopAck(w io.Writer, pkt AttackStopAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## DeadAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Attack/5137_DeadAckModel.cs
```go
type DeadAck struct {
    DefenseSessionGameId uint32
    OffenseSessionGameId uint32
    Chaotic int32
    ChaoticStatus ChaoticStatusType
}

func SendDeadAck(w io.Writer, pkt DeadAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## InformationCharacter (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5101_InformationCharacterModel.cs
```go
type InformationCharacter struct {
    Id int32
    Class byte
    Gender byte
    Head byte
    Face byte
    Level int16
    Name string
    Str int32
    Dex int32
    Int int32
    Chaotic int32
    Position Vec3
    Weapon Item
    Shield Item
    Armor Item
    FirstRing Item
    SecondRing Item
    Necklace Item
    Boots Item
    Gloves Item
    Helmet Item
    Belt Item
    Cloak Item
    SphereMastery Item
    SphereSoul Item
    SphereDefense Item
    SphereDestruction Item
    SphereLife Item
    SphereLuck Item
    SphereReincarnation Item
    SphereCharacteristics Item
    Id uint64
    ItemId int32
}

func SendInformationCharacter(w io.Writer, pkt InformationCharacter) error {
    // TODO: marshal fields and write
    return nil
}
```

## DisplayedCharacter (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5103_DisplayedCharacterModel.cs
```go
type DisplayedCharacter struct {
    Character PublicPc
    IsTeleport bool
}

func SendDisplayedCharacter(w io.Writer, pkt DisplayedCharacter) error {
    // TODO: marshal fields and write
    return nil
}
```

## ExistedPcAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5107_ExistedPcAckModel.cs
```go
type ExistedPcAck struct {
}

func SendExistedPcAck(w io.Writer, pkt ExistedPcAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## CompleteCreateCharacter (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5119_CompleteCreateCharacterModel.cs
```go
type CompleteCreateCharacter struct {
    CharacterId int32
    Str int32
    Dex int32
    Int int32
}

func SendCompleteCreateCharacter(w io.Writer, pkt CompleteCreateCharacter) error {
    // TODO: marshal fields and write
    return nil
}
```

## CompleteDeleteCharacter (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5121_CompleteDeleteCharacterModel.cs
```go
type CompleteDeleteCharacter struct {
}

func SendCompleteDeleteCharacter(w io.Writer, pkt CompleteDeleteCharacter) error {
    // TODO: marshal fields and write
    return nil
}
```

## RespawnAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5142_RespawnAckModel.cs
```go
type RespawnAck struct {
    SessionGameId uint32
    Position Vec3
}

func SendRespawnAck(w io.Writer, pkt RespawnAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## InfoStomachAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5173_InfoStomachModel.cs
```go
type InfoStomachAck struct {
    Stomach int32
    StomachStatus byte
}

func SendInfoStomachAck(w io.Writer, pkt InfoStomachAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## MovedCharacter (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5189_MovedCharacterModel.cs
```go
type MovedCharacter struct {
    SessionGameId uint32
    Position Vec3
    MoveRate int16
    Flag byte
    DirectionSight float32
    Action uint32
}

func SendMovedCharacter(w io.Writer, pkt MovedCharacter) error {
    // TODO: marshal fields and write
    return nil
}
```

## JumpEndCharacter (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5193_JumpEndCharacterModel.cs
```go
type JumpEndCharacter struct {
    SessionGameId uint32
    Action uint32
    DirectionSight float32
}

func SendJumpEndCharacter(w io.Writer, pkt JumpEndCharacter) error {
    // TODO: marshal fields and write
    return nil
}
```

## CharDirAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/5195_CharDirModel.cs
```go
type CharDirAck struct {
    SessionGameId uint32
    DirectionSight float32
}

func SendCharDirAck(w io.Writer, pkt CharDirAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## InventoryCharacteristic (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/Characteristics/5145_InventoryCharacteristicModel.cs
```go
type InventoryCharacteristic struct {
    DDv int16
    MDv int16
    RDv int16
    DPv int16
    MPv int16
    RPv int16
    DDD int16
    DHit int16
    RDD int16
    RHit int16
    MDD int16
    MHit int16
    Str int16
    Dex int16
    Int int16
    CriticalHit int16
    HpMax int32
    MpMax int32
}

func SendInventoryCharacteristic(w io.Writer, pkt InventoryCharacteristic) error {
    // TODO: marshal fields and write
    return nil
}
```

## HealthPointCharacteristic (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/Characteristics/5146_HealthPointCharacteristicModel.cs
```go
type HealthPointCharacteristic struct {
    Hp int32
    Mp int32
}

func SendHealthPointCharacteristic(w io.Writer, pkt HealthPointCharacteristic) error {
    // TODO: marshal fields and write
    return nil
}
```

## SpeedCharacteristic (S->C)
// Packets/Packets.Server.Game/Models/Send/Character/Characteristics/5147_SpeedCharacteristicModel.cs
```go
type SpeedCharacteristic struct {
    AttackRate int16
    MoveRate int16
    SessionGameId uint32
}

func SendSpeedCharacteristic(w io.Writer, pkt SpeedCharacteristic) error {
    // TODO: marshal fields and write
    return nil
}
```

## ChatAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Chat/2034_ChatAckModel.cs
```go
type ChatAck struct {
    Type byte
    SessionGameId uint32
    Name string
    Message string
}

func SendChatAck(w io.Writer, pkt ChatAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## GossipAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Chat/5212_GossipAckModel.cs
```go
type GossipAck struct {
    UniqueIdentifier uint32
    FromName string
    ToName string
    Message string
}

func SendGossipAck(w io.Writer, pkt GossipAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## GlobalChatAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Chat/5226_GlobalChatAckModel.cs
```go
type GlobalChatAck struct {
    SessionGameId uint32
    Name string
    Message string
}

func SendGlobalChatAck(w io.Writer, pkt GlobalChatAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## EmoticonAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Chat/5835_EmoticonAckModel.cs
```go
type EmoticonAck struct {
    Type int32
    SessionGameId uint32
    Name string
}

func SendEmoticonAck(w io.Writer, pkt EmoticonAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## EnteredItemAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5105_EnteredItemAckModel.cs
```go
type EnteredItemAck struct {
    Item PublicItem
}

func SendEnteredItemAck(w io.Writer, pkt EnteredItemAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ExistedItemAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5110_ExistedItemAckModel.cs
```go
type ExistedItemAck struct {
}

func SendExistedItemAck(w io.Writer, pkt ExistedItemAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ExitMapGbjAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5114_ExitMapGbjAckModel.cs
```go
type ExitMapGbjAck struct {
    UniqueItemDrop uint32
    Why ExitMapWhy
}

func SendExitMapGbjAck(w io.Writer, pkt ExitMapGbjAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## EquipAckAll (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5129_EquipAckAllModel.cs
```go
type EquipAckAll struct {
    SessionGameId uint32
    ItemId int32
    SerialNumber uint64
    Position ItemPositionType
}

func SendEquipAckAll(w io.Writer, pkt EquipAckAll) error {
    // TODO: marshal fields and write
    return nil
}
```

## UnEquipAckAll (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5131_UnEquipAckAllModel.cs
```go
type UnEquipAckAll struct {
    Position ItemPositionType
    SessionGameId uint32
}

func SendUnEquipAckAll(w io.Writer, pkt UnEquipAckAll) error {
    // TODO: marshal fields and write
    return nil
}
```

## InfoWeightAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5149_InfoWeightAckModel.cs
```go
type InfoWeightAck struct {
    MaxWeight int32
    Weight int32
    WeightStatus byte
}

func SendInfoWeightAck(w io.Writer, pkt InfoWeightAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ReinforceAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5169_ReinforceAckModel.cs
```go
type ReinforceAck struct {
    SerialNumber uint64
    ItemId int32
    IsConfirm byte
}

func SendReinforceAck(w io.Writer, pkt ReinforceAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ReinforceNak1 (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5170_ReinforceNak1Model.cs
```go
type ReinforceNak1 struct {
    IsDestroy byte
    SerialNumber uint64
}

func SendReinforceNak1(w io.Writer, pkt ReinforceNak1) error {
    // TODO: marshal fields and write
    return nil
}
```

## ItemAddAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5232_ItemAddAckModel.cs
```go
type ItemAddAck struct {
    Item ItemApiModel
    SessionGameId uint32
    Reason byte
}

func SendItemAddAck(w io.Writer, pkt ItemAddAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ItemRemoveAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5233_ItemRemoveAckModel.cs
```go
type ItemRemoveAck struct {
    SerialNumber uint64
    Count int32
    SessionGameId uint32
    Reason byte
}

func SendItemRemoveAck(w io.Writer, pkt ItemRemoveAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ItemChangeAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5237_ItemChangeAckModel.cs
```go
type ItemChangeAck struct {
    SerialNumber uint64
    ItemId int32
    Reason byte
    IsCreate int32
}

func SendItemChangeAck(w io.Writer, pkt ItemChangeAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ItemUseAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Inventory/5653_ItemUseAckModel.cs
```go
type ItemUseAck struct {
    ItemId int32
}

func SendItemUseAck(w io.Writer, pkt ItemUseAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## InfoExpAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Level/5139_InfoExpAckModel.cs
```go
type InfoExpAck struct {
    Level int16
    Exp int64
    ExpAim uint64
}

func SendInfoExpAck(w io.Writer, pkt InfoExpAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## LevelUpAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Level/5140_LevelUpAckModel.cs
```go
type LevelUpAck struct {
    SessionGameId uint32
    Hp int16
    Mp int16
}

func SendLevelUpAck(w io.Writer, pkt LevelUpAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## EnteredMonAck (S->C)
// Packets/Packets.Server.Game/Models/Send/MonsterNpc/5104_EnteredMonAckModel.cs
```go
type EnteredMonAck struct {
    Monster MonsterApiModel
    IsTeleport bool
    CntAbn byte
}

func SendEnteredMonAck(w io.Writer, pkt EnteredMonAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ExistedMonAck (S->C)
// Packets/Packets.Server.Game/Models/Send/MonsterNpc/5108_ExistedMonAckModel.cs
```go
type ExistedMonAck struct {
}

func SendExistedMonAck(w io.Writer, pkt ExistedMonAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## DoMoveToAck (S->C)
// Packets/Packets.Server.Game/Models/Send/MonsterNpc/5190_DoMoveToAckModel.cs
```go
type DoMoveToAck struct {
    SessionGameId uint32
    Position Vec3
    PointPosition Vec3
    Flag byte
    Velocity float32
}

func SendDoMoveToAck(w io.Writer, pkt DoMoveToAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## MerchantListAck (S->C)
// Packets/Packets.Server.Game/Models/Send/Npc/5271_MerchantListAckModel.cs
```go
type MerchantListAck struct {
    UniqueIdentifier uint32
    MerchantId int32
    ParmA int32
    ParmB int32
    CountBuy int32
    CountSell int32
    CountCharge int32
    PaymentType PaymentType
    ItemId int32
    Price int32
    Flag int32
    SortKey int32
}

func SendMerchantListAck(w io.Writer, pkt MerchantListAck) error {
    // TODO: marshal fields and write
    return nil
}
```

## ScrDialogNoMsg2Ack (S->C)
// Packets/Packets.Server.Game/Models/Send/Npc/5902_ScrDialogNoMsg2AckModel.cs
```go
type ScrDialogNoMsg2Ack struct {
    ScriptId int32
    UniqueIdentifier uint32
    Param int32
}

func SendScrDialogNoMsg2Ack(w io.Writer, pkt ScrDialogNoMsg2Ack) error {
    // TODO: marshal fields and write
    return nil
}
```

## ServerTime (S->C)
// Packets/Packets.Server.Game/Models/Send/Settings/5651_ServerTimeModel.cs
```go
type ServerTime struct {
    ServerTick int32
    Year int16
    Month int16
    DayOfWeek int16
    Day int16
    Hour int16
    Minute int16
    Second int16
    Millisecond int16
}

func SendServerTime(w io.Writer, pkt ServerTime) error {
    // TODO: marshal fields and write
    return nil
}
```

## CheckNeedMoney (S->C)
// Packets/Packets.Server.Game/Models/Send/Settings/5813_CheckNeedMoneyModel.cs
```go
type CheckNeedMoney struct {
}

func SendCheckNeedMoney(w io.Writer, pkt CheckNeedMoney) error {
    // TODO: marshal fields and write
    return nil
}
```

## AuthorizationLogin (C->S)
// Packets/Packets.Server.Login/Models/Receive/3100_AuthorizationLoginModel.cs
```go
type AuthorizationLogin struct {
    Login string
    Password string
}

func ReadAuthorizationLogin(r io.Reader) (AuthorizationLogin, error) {
    var pkt AuthorizationLogin
    // TODO: read fields from r
    return pkt, nil
}
```

## RefreshServers (C->S)
// Packets/Packets.Server.Login/Models/Receive/3115_RefreshServersModel.cs
```go
type RefreshServers struct {
}

func ReadRefreshServers(r io.Reader) (RefreshServers, error) {
    var pkt RefreshServers
    // TODO: read fields from r
    return pkt, nil
}
```

## SelectServer (C->S)
// Packets/Packets.Server.Login/Models/Receive/3120_SelectServerModel.cs
```go
type SelectServer struct {
    AccountId int32
    Login string
    ServerId int16
}

func ReadSelectServer(r io.Reader) (SelectServer, error) {
    var pkt SelectServer
    // TODO: read fields from r
    return pkt, nil
}
```

## ConnectionClient (S->C)
// Packets/Packets.Server.Login/Models/Send/1103_ConnectionClientModel.cs
```go
type ConnectionClient struct {
    DecryptKey []byte
}

func SendConnectionClient(w io.Writer, pkt ConnectionClient) error {
    // TODO: marshal fields and write
    return nil
}
```

## SendServers (S->C)
// Packets/Packets.Server.Login/Models/Send/3101_SendServersModel.cs
```go
type SendServers struct {
    AccountId int32
    SessionId int32
}

func SendSendServers(w io.Writer, pkt SendServers) error {
    // TODO: marshal fields and write
    return nil
}
```

## LoginServerError (S->C)
// Packets/Packets.Server.Login/Models/Send/3102_LoginServerErrorModel.cs
```go
type LoginServerError struct {
    ErrorType ServerErrorType
}

func SendLoginServerError(w io.Writer, pkt LoginServerError) error {
    // TODO: marshal fields and write
    return nil
}
```

## RefreshedServers (S->C)
// Packets/Packets.Server.Login/Models/Send/3116_RefreshedServersModel.cs
```go
type RefreshedServers struct {
}

func SendRefreshedServers(w io.Writer, pkt RefreshedServers) error {
    // TODO: marshal fields and write
    return nil
}
```

## SelectedServer (S->C)
// Packets/Packets.Server.Login/Models/Send/3121_SelectedServerModel.cs
```go
type SelectedServer struct {
}

func SendSelectedServer(w io.Writer, pkt SelectedServer) error {
    // TODO: marshal fields and write
    return nil
}
```

