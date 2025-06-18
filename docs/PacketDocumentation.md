# R2 Online Packet Documentation

Generated automatically by `Scripts/generate_packet_docs.py`.

## 1102 - GameServerError

**Direction:** Server -> Client

Model serevr error

| Field | Type |
| --- | --- |
| PacketType | PacketType |
| ErrorType | GameServerErrorType |
| IsMsgBox | bool |


## 1103 - ConnectionClient

**Direction:** Server -> Client

Connection client model send DecryptKey

| Field | Type |
| --- | --- |
| DecryptKey | byte[] |


## 1103 - ConnectionClient

**Direction:** Server -> Client

Connection client model send DecryptKey

| Field | Type |
| --- | --- |
| DecryptKey | byte[] |


## 2033 - ChatReq

**Direction:** Client -> Server

Model for chat req

| Field | Type |
| --- | --- |
| Type | byte |
| Message | string |
| Name | string |


## 2034 - ChatAck

**Direction:** Server -> Client

Model for chat ack

| Field | Type |
| --- | --- |
| Type | byte |
| SessionGameId | UniqueId |
| Name | string |
| Message | string |


## 3100 - AuthorizationLogin

**Direction:** Client -> Server

Model authorization client by login and password

| Field | Type |
| --- | --- |
| Login | string |
| Password | string |


## 3101 - SendServers

**Direction:** Server -> Client

Send server model

| Field | Type |
| --- | --- |
| AccountId | int |
| SessionId | int |
| Servers | List<ServerModel> |


## 3102 - LoginServerError

**Direction:** Server -> Client

Error type server model

| Field | Type |
| --- | --- |
| ErrorType | ServerErrorType |


## 3115 - RefreshServers

**Direction:** Client -> Server

Model refresh servers


## 3116 - RefreshedServers

**Direction:** Server -> Client

Refreshed server model

| Field | Type |
| --- | --- |
| Servers | List<ServerModel> |


## 3120 - SelectServer

**Direction:** Client -> Server

Select server model

| Field | Type |
| --- | --- |
| AccountId | int |
| Login | string |
| ServerId | short |


## 3121 - SelectedServer

**Direction:** Server -> Client

Select server model


## 5100 - LoginUserReq

**Direction:** Client -> Server

Model for login user req

| Field | Type |
| --- | --- |
| AccountId | uint |
| SessionId | int |
| Password | string |


## 5101 - InformationCharacter

**Direction:** Server -> Client

Model for send character

| Field | Type |
| --- | --- |
| Characters | List<Character> |
| Equipments | List<Equipment> |
| Id | int |
| Class | byte |
| Gender | byte |
| Head | byte |
| Face | byte |
| Level | short |
| Name | string |
| Str | int |
| Dex | int |
| Int | int |
| Chaotic | int |
| Position | Vector3 |
| Weapon | Item |
| Shield | Item |
| Armor | Item |
| FirstRing | Item |
| SecondRing | Item |
| Necklace | Item |
| Boots | Item |
| Gloves | Item |
| Helmet | Item |
| Belt | Item |
| Cloak | Item |
| SphereMastery | Item |
| SphereSoul | Item |
| SphereDefense | Item |
| SphereDestruction | Item |
| SphereLife | Item |
| SphereLuck | Item |
| SphereReincarnation | Item |
| SphereCharacteristics | Item |
| Id | ulong |
| ItemId | int |


## 5103 - DisplayedCharacter

**Direction:** Server -> Client

Model displayed character

| Field | Type |
| --- | --- |
| Character | PublicPc |
| IsTeleport | bool |


## 5104 - EnteredMonAck

**Direction:** Server -> Client

Model for entered mon ack

| Field | Type |
| --- | --- |
| Monster | MonsterApiModel |
| IsTeleport | bool |
| CntAbn | byte |


## 5105 - EnteredItemAck

**Direction:** Server -> Client

Model for entered item ack

| Field | Type |
| --- | --- |
| Item | PublicItem |


## 5107 - ExistedPcAck

**Direction:** Server -> Client

Model existed pc

| Field | Type |
| --- | --- |
| Character | List<PublicPc> |


## 5108 - ExistedMonAck

**Direction:** Server -> Client

Model for existed mon ack

| Field | Type |
| --- | --- |
| NpcMonsters | List<MonsterApiModel> |


## 5110 - ExistedItemAck

**Direction:** Server -> Client

Model for existed item ack

| Field | Type |
| --- | --- |
| Items | List<PublicItem> |


## 5114 - ExitMapGbjAck

**Direction:** Server -> Client

Model for exit map gbj ack model

| Field | Type |
| --- | --- |
| UniqueItemDrop | UniqueId |
| Why | ExitMapWhy |


## 5115 - LogoutPcReq

**Direction:** Client -> Server

Model for logout pc req


## 5116 - ChoosePcReq

**Direction:** Client -> Server

Model for choose pc req

| Field | Type |
| --- | --- |
| PcNo | uint |


## 5117 - CompleteEnterWorld

**Direction:** Server -> Client

Model complete enter world and items to inventory

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Position | Vector3 |
| Reputation | short |
| Items | List<ItemApiModel> |
| MoveRate | short |
| AttackRate | short |


## 5118 - CreatePcReq

**Direction:** Client -> Server

Model for creating pc

| Field | Type |
| --- | --- |
| Slot | byte |
| Class | byte |
| Sex | byte |
| Head | byte |
| Face | byte |
| TypeBody | byte |
| Name | string |


## 5119 - CompleteCreateCharacter

**Direction:** Server -> Client

Model complete create character

| Field | Type |
| --- | --- |
| CharacterId | int |
| Str | int |
| Dex | int |
| Int | int |


## 5120 - DeletePcReq

**Direction:** Client -> Server

Model for delete pc

| Field | Type |
| --- | --- |
| PcNo | uint |
| Slot | byte |


## 5121 - CompleteDeleteCharacter

**Direction:** Server -> Client

Model complete create character


## 5128 - EquipReq

**Direction:** Client -> Server

Model for equip req

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| Index | uint |


## 5129 - EquipAckAll

**Direction:** Server -> Client

Model for equip ack all

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| ItemId | int |
| SerialNumber | ulong |
| Position | ItemPositionType |


## 5130 - UnEquipReq

**Direction:** Client -> Server

Model for un equip req

| Field | Type |
| --- | --- |
| Position | ItemPositionType |


## 5131 - UnEquipAckAll

**Direction:** Server -> Client

Model for un equip ack all

| Field | Type |
| --- | --- |
| Position | ItemPositionType |
| SessionGameId | UniqueId |


## 5132 - AttackAck

**Direction:** Server -> Client

Model for attack ack

| Field | Type |
| --- | --- |
| OffenseSessionGameId | UniqueId |
| DefenseSessionGameId | UniqueId |
| TypeHit | TypeHit |
| OffensePosition | Vector3 |
| HpAttacked | short |


## 5133 - AttackReq

**Direction:** Client -> Server

Attack req model

| Field | Type |
| --- | --- |
| TargetSessionGameId | UniqueId |
| AttackType | ushort |
| AttackPosition | Vector3 |
| AttackFlag | byte |


## 5134 - AttackStopAck

**Direction:** Server -> Client

Model for attack stop ack

| Field | Type |
| --- | --- |
| OffenceSesionGameId | UniqueId |


## 5137 - DeadAck

**Direction:** Server -> Client

Model for dead ack

| Field | Type |
| --- | --- |
| DefenseSessionGameId | UniqueId |
| OffenseSessionGameId | UniqueId |
| Chaotic | int |
| ChaoticStatus | ChaoticStatusType |


## 5139 - InfoExpAck

**Direction:** Server -> Client

Model for info exp ack

| Field | Type |
| --- | --- |
| Level | short |
| Exp | long |
| ExpAim | ulong |


## 5140 - LevelUpAck

**Direction:** Server -> Client

Model for level up ack

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Hp | short |
| Mp | short |


## 5141 - RespawnReq

**Direction:** Client -> Server

Model for respawn req

| Field | Type |
| --- | --- |
| IsRespawn | byte |


## 5142 - RespawnAck

**Direction:** Server -> Client

Model for respawn ack

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Position | Vector3 |


## 5145 - InventoryCharacteristic

**Direction:** Server -> Client

Model characteristic in inventory

| Field | Type |
| --- | --- |
| DDv | short |
| MDv | short |
| RDv | short |
| DPv | short |
| MPv | short |
| RPv | short |
| DDD | short |
| DHit | short |
| RDD | short |
| RHit | short |
| MDD | short |
| MHit | short |
| Str | short |
| Dex | short |
| Int | short |
| CriticalHit | short |
| HpMax | int |
| MpMax | int |


## 5146 - HealthPointCharacteristic

**Direction:** Server -> Client

Model health and mana

| Field | Type |
| --- | --- |
| Hp | int |
| Mp | int |


## 5147 - SpeedCharacteristic

**Direction:** Server -> Client

Model characteristic sped

| Field | Type |
| --- | --- |
| AttackRate | short |
| MoveRate | short |
| SessionGameId | UniqueId |


## 5149 - InfoWeightAck

**Direction:** Server -> Client

Model for info weight ack

| Field | Type |
| --- | --- |
| MaxWeight | int |
| Weight | int |
| WeightStatus | byte |


## 5151 - ScriptReq

**Direction:** Client -> Server

Model for req npc dialog

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| Param | int |


## 5152 - ScriptProcReq

**Direction:** Client -> Server

Model for req npc dialog

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| ScriptAction | ScriptAction |
| Param | int |


## 5158 - ItemUseReq

**Direction:** Client -> Server

Model for un equip req

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| ItemId | int |
| UniqueIdentifier | UniqueId |
| TargetSerialNumber | ulong |
| IsTeam | byte |


## 5159 - ItemDropReq

**Direction:** Client -> Server

Model for item drop req

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| Stack | uint |


## 5160 - AbnormalAck

**Direction:** Server -> Client

Model abnormal character

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| BuffId | int |
| EndTick | uint |
| Position | Vector3 |


## 5161 - AbnormaleReleaseAck

**Direction:** Server -> Client

Model abnormal character

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| Type | int |


## 5163 - AbnormalRemoveReq

**Direction:** Client -> Server

Model for abnormal remove req

| Field | Type |
| --- | --- |
| Type | int |


## 5168 - ReinforceReq

**Direction:** Client -> Server

Model for reinforce req

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| SerialNumber0 | ulong |
| SerialNumber1 | ulong |
| SerialNumber2 | ulong |
| Count | int |


## 5169 - ReinforceAck

**Direction:** Server -> Client

Model for reinforce ack

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| ItemId | int |
| IsConfirm | byte |


## 5170 - ReinforceNak1

**Direction:** Server -> Client

Model for reinforce Nak

| Field | Type |
| --- | --- |
| IsDestroy | byte |
| SerialNumber | ulong |


## 5173 - InfoStomachAck

**Direction:** Server -> Client

Модель голода

| Field | Type |
| --- | --- |
| Stomach | int |
| StomachStatus | byte |


## 5177 - ItemPickupReq

**Direction:** Client -> Server

Model for item pick up req

| Field | Type |
| --- | --- |
| UniqueIdentifierItem | UniqueId |


## 5179 - TransformAck

**Direction:** Server -> Client

Model displayed character

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| MonsterId | int |


## 5188 - DoMoveReq

**Direction:** Client -> Server

Model for do move req

| Field | Type |
| --- | --- |
| Position | Vector3 |
| Direction | float |
| Action | uint |
| Flag | byte |


## 5189 - MovedCharacter

**Direction:** Server -> Client

Model moved character

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Position | Vector3 |
| MoveRate | short |
| Flag | byte |
| DirectionSight | float |
| Action | uint |


## 5190 - DoMoveToAck

**Direction:** Server -> Client

Model for do move to ack

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Position | Vector3 |
| PointPosition | Vector3 |
| Flag | byte |
| Velocity | float |


## 5192 - CharJumpReq

**Direction:** Client -> Server

Model for char jump req

| Field | Type |
| --- | --- |
| Action | uint |
| MoveDirection | float |


## 5193 - JumpEndCharacter

**Direction:** Server -> Client

Model end jump

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Action | uint |
| DirectionSight | float |


## 5194 - CharDirReq

**Direction:** Client -> Server

Model for char dir req

| Field | Type |
| --- | --- |
| Direction | float |


## 5195 - CharDirAck

**Direction:** Server -> Client

Model for char direction ack

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| DirectionSight | float |


## 5212 - GossipAck

**Direction:** Server -> Client

Model for chat ack

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| FromName | string |
| ToName | string |
| Message | string |


## 5225 - GlobalChatReq

**Direction:** Client -> Server

Model for global chat req

| Field | Type |
| --- | --- |
| Message | string |


## 5226 - GlobalChatAck

**Direction:** Server -> Client

Model for global chat ack

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Name | string |
| Message | string |


## 5232 - ItemAddAck

**Direction:** Server -> Client

Model for item add ack

| Field | Type |
| --- | --- |
| Item | ItemApiModel |
| SessionGameId | UniqueId |
| Reason | byte |


## 5233 - ItemRemoveAck

**Direction:** Server -> Client

Model for item remove ack

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| Count | int |
| SessionGameId | UniqueId |
| Reason | byte |


## 5237 - ItemChangeAck

**Direction:** Server -> Client

Model for item change ack

| Field | Type |
| --- | --- |
| SerialNumber | ulong |
| ItemId | int |
| Reason | byte |
| IsCreate | int |


## 5271 - MerchantListAck

**Direction:** Server -> Client

Model for merchant list ack

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| MerchantId | int |
| ParmA | int |
| ParmB | int |
| CountBuy | int |
| CountSell | int |
| CountCharge | int |
| PaymentType | PaymentType |
| ItemList | List<Item> |
| ItemId | int |
| Price | int |
| Flag | int |
| SortKey | int |


## 5273 - MerchantBuyReq

**Direction:** Client -> Server

Model for merchant buy

| Field | Type |
| --- | --- |
| UniqueIdentifier | UniqueId |
| ItemId | int |
| Count | int |
| ParmA | int |
| ParmB | int |
| SortKey | int |


## 5326 - StopMoveCharacter

**Direction:** Server -> Client

No description.

| Field | Type |
| --- | --- |
| SessionGameId | UniqueId |
| Position | Vector3 |


## 5652 - ServerTime

**Direction:** Server -> Client

Model server time

| Field | Type |
| --- | --- |
| ServerTick | int |
| Year | short |
| Month | short |
| DayOfWeek | short |
| Day | short |
| Hour | short |
| Minute | short |
| Second | short |
| Millisecond | short |


## 5653 - ItemUseAck

**Direction:** Server -> Client

Model displayed character

| Field | Type |
| --- | --- |
| ItemId | int |


## 5654 - ItemCooldown

**Direction:** Server -> Client

Model cooldown ack

| Field | Type |
| --- | --- |
| ItemId | int |


## 5784 - UseSkillPackReq

**Direction:** Client -> Server

Model for use skill pack req

| Field | Type |
| --- | --- |
| SkillId | int |
| TargetUniqueId | UniqueId |
| IsTeam | byte |


## 5792 - UseSkillPackAck

**Direction:** Server -> Client

Model for use skill pack ack

| Field | Type |
| --- | --- |
| SkillId | int |


## 5813 - CheckNeedMoney

**Direction:** Server -> Client

Model check need money


## 5834 - EmoticonReq

**Direction:** Client -> Server

Model for emoji req

| Field | Type |
| --- | --- |
| Type | int |


## 5835 - EmoticonAck

**Direction:** Server -> Client

Model for emoji ack

| Field | Type |
| --- | --- |
| Type | int |
| SessionGameId | UniqueId |
| Name | string |


## 5902 - ScrDialogNoMsg2Ack

**Direction:** Server -> Client

Model for ack npc dialog

| Field | Type |
| --- | --- |
| ScriptId | int |
| UniqueIdentifier | UniqueId |
| Param | int |

