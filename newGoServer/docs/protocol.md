# R2 Online — Сетевой протокол и шифрование

## Рабочий флоу (полная цепочка)

```
┌─────────────────────────────────────────────────────────────────┐
│                       LOGIN SERVER (:11004)                      │
└─────────────────────────────────────────────────────────────────┘

Клиент  ──TCP connect──►  Login Server

Server  ──► ConnectionClient (opcode 1103)
              • 198 байт (WelcomeKey), plaintext (enc_flag=0x00)
              • Структура CTrCryptKeyAck: содержит 10 ключей CCryptKey по 18 байт,
                поле __mIdx=80 (обязательно), __mSvrNo, __mUnique
              • __mKey8 (bytes[96..113]) — реальный ключ шифрования клиента
              • После отправки сервер устанавливает DecryptSbox для входящих пакетов

Клиент  ──► AuthorizationLogin (opcode 3100)
              • Зашифрован RC4(DecryptSbox), enc_flag=0x01
              • Содержит: логин, пароль (обфусцированные смещения)
              • Размер: ~3011 байт (с паддингом)

Server  ──► SendServers (opcode 3101), plaintext
              • Layout: [AccountID: int32][SessionID: int32][Count: uint8][...серверы]
              • AccountID и SessionID (Redis-токен) передаются клиенту здесь
              • Клиент сохраняет оба значения и отправит их геймсерверу

Клиент  ──► RefreshServers (opcode 3115), зашифрован  [опционально]
Server  ──► RefreshedServers (opcode 3116), plaintext  [опционально]
              • Тот же layout что SendServers (AccountID + SessionID + серверы)

Клиент  ──► SelectServer (opcode 3120), зашифрован
              • AccountID, Login (20 байт), ServerID

Server  ──► SelectedServer (opcode 3121), plaintext
              • 4 нулевых байта
              • Клиент закрывает соединение и переподключается к геймсерверу

┌─────────────────────────────────────────────────────────────────┐
│                       GAME SERVER (:5000)                        │
└─────────────────────────────────────────────────────────────────┘

Клиент  ──TCP connect──►  Game Server

Server  ──► ConnectionClient (opcode 1103)
              • Те же 198 байт (WelcomeKey), plaintext
              • После отправки сервер устанавливает DecryptSbox для входящих пакетов

Клиент  ──► LoginUserReq (opcode 5100), зашифрован
              • [AccountID: uint32][SessionID: int32][Pad: 4 байта][Password: 21 байт CP-1251]
              • SessionID — тот токен, что клиент получил в SendServers (3101)
              • Сервер валидирует токен через Redis → получает accountID

Server  ──► InformationCharacter (opcode 5101), plaintext
              • Список из 3 слотов персонажей (144 байта каждый)
              • Экипировка персонажей (320 байт каждый)

Клиент  ──► ChoosePcReq (opcode 5116), зашифрован
              • Номер слота персонажа

Server  ──► CompleteEnterWorld (opcode 5117), plaintext
Server  ──► (пакеты характеристик, инвентаря, скорости, опыта...)

Клиент находится в игровом мире.
```

---

## Формат пакета (wire format)

```
[SIZE:  uint16 LE]  — полный размер пакета включая эти 2 байта
[ENC:   uint8]      — 0x00 = plaintext, 0x01 = зашифрован RC4
[SEQ:   uint8]      — порядковый номер пакета
[OP:    uint16 LE]  — опкод
[DATA:  bytes...]   — полезная нагрузка

SIZE = 6 + len(DATA)

Шифруется: SEQ + OP + DATA (то есть всё после ENC-байта)
```

---

## Шифрование — исследованное состояние

### Структура CTrCryptKeyAck (ConnectionClient, opcode 1103)

198-байтовый пакет содержит структуру `CTrCryptKeyAck` (из реверса `FIELDW_REVERSED_exe.c`):

```c
struct CCryptKey {
    short __mKey[9];  // 18 байт
};

struct CTrCryptKeyAck {
    byte    field0, field1;       // offset 0-1
    CCryptKey __mKey0;            // offset 2-19
    short   __mDummy0;            // offset 20-21
    CCryptKey __mKey1;            // offset 22-39
    byte    __mDummy1;            // offset 40
    CCryptKey __mKey2;            // offset 41-58
    CCryptKey __mKey3;            // offset 59-76
    byte    __mIdx;               // offset 77 — ОБЯЗАТЕЛЬНО = 80, иначе клиент крашит
    CCryptKey __mKey4;            // offset 78-95
    short   __mDummy2;            // offset 96-97
    CCryptKey __mKey5;            // offset 98-115
    CCryptKey __mKey6;            // offset 116-133
    short   __mDummy3;            // offset 134-135
    CCryptKey __mKey7;            // offset 136-153
    byte    __mDummy4;            // offset 154
    CCryptKey __mKey8;            // offset 155-172  ← КЛЮЧ ШИФРОВАНИЯ
    CCryptKey __mKey9;            // offset 173-190
    short   __mDummy5;            // offset 191-192
    byte    __mDummy6;            // offset 193
    ushort  __mSvrNo;             // offset 194-195
    ulong   __mUnique;            // offset 196-199
};
```

**Важно:** Сообщество подтвердило эмпирически — изменение `__mKey8` меняет шифрование пакетов.
Остальные ключи (`__mKey0..7`, `__mKey9`) "фейковые" — их изменение не влияет на шифрование.

Фактические байты `__mKey8` в нашем WelcomeKey (offset 96-113 из-за смещения dummy-полей):
```
c8 07 11 4e a9 35 a9 02 e3 19 b1 7e b3 59 eb 53 fe 76
```

### Алгоритм шифра

Кастомный RC4 без стандартного KSA. S-box используется напрямую как начальное состояние:

```go
// Одна итерация (применяется к каждому байту тела пакета):
i++
v := S[i]      // S[i]
j += v         // j += S[i]
u := S[j]      // S[j]
S[i] = u       // своп
S[j] = v       // своп
byte ^= S[v+u] // XOR с S[S[i]+S[j]]
```

**Шифр сбрасывается к начальному состоянию для каждого пакета** (не стримовый).

### Статический keystream

Поскольку S-box фиксирован и сбрасывается per-packet, keystream **всегда одинаковый**:

```
Keystream[0..63] = 8A 7B 65 7E B7 DC C4 04 1A D7 BF B1 11 3A DF B5
                   05 78 07 AD 0F 72 71 1E 53 1E FC 97 B2 22 72 3E
                   E6 AF CE 16 82 8C 2C E1 7E 47 32 6D 29 89 A7 35
                   00 0B D5 F1 09 24 E2 DB E6 5F 95 AD 82 06 3A 00
```

Сообщество R2 Online подтвердило: один статичный XOR-ключ (3000 байт) работает для расшифровки **любого** пакета — что и является доказательством статического keystream.

### Направления

| Направление | Шифрование |
|---|---|
| Сервер → Клиент | **Всегда plaintext** (enc_flag=0x00) |
| Клиент → Сервер | **RC4(DecryptSbox)** (enc_flag=0x01), сброс на каждый пакет |

### Алгоритм KSA (реверс FnlApiW.dll)

Реверс `FnlApiW.dll` (C:\r2_server\lib\Lib\FnlApiW.dll) показал:

- `FnlApi::CRc4A::SState { int mX; int mY; int mM[256]; }` — S-box хранится как 256 int
- `CRc4A::SetKey(SState&, const void*, unsigned long)` — **стандартный RC4 KSA**, просто с int32 S-box
- Разобранные байты функции подтверждают: инициализация `mM[i]=i`, затем стандартный KSA-shuffle с байт-маской j

Значит клиент тоже использует стандартный RC4 KSA для вычисления S-box из `__mKey8`.

---

## Per-session шифрование (реализовано)

**РЕАЛИЗОВАНО** в текущем коде. Алгоритм:

1. Сервер генерирует случайные 18 байт (`crypto.GenerateSessionKey()`)
2. Встраивает их в `__mKey8` (смещение 155) WelcomeKey
3. Вычисляет S-box через `crypto.KSA(key8[:])` — стандартный RC4 KSA
4. Отправляет ConnectionClient с этим WelcomeKey
5. Клиент читает `__mKey8`, запускает тот же KSA → получает тот же S-box
6. Все последующие пакеты клиента зашифрованы этим уникальным S-box

```go
// Реализация в login/server.go и game/server.go:
key8, _ := crypto.GenerateSessionKey()           // 18 случайных байт
pkt := &send.ConnectionClient{Key8: key8}        // __mKey8 в WelcomeKey
conn.Send(opcode.ConnectionClient, pkt.Encode()) // клиент получает ключ
conn.SetRecvSbox(crypto.KSA(key8[:]))            // уникальный per-session S-box
```

Каждое соединение теперь имеет уникальный S-box. Перехват трафика одного игрока
не позволяет расшифровать трафик другого.

---

## Файлы реализации

| Файл | Назначение |
|---|---|
| `internal/crypto/rc4.go` | Алгоритм RC4 без KSA |
| `internal/crypto/sbox.go` | `DecryptSbox` — текущий фиксированный S-box |
| `internal/packet/login/send/connection_client.go` | `WelcomeKey` (CTrCryptKeyAck) — 198 байт |
| `internal/network/conn.go` | Фреймер: чтение/запись пакетов, расшифровка входящих |
| `internal/login/server.go` | Login-сервер: отправка ConnectionClient, установка RecvSbox |
| `internal/game/server.go` | Game-сервер: то же самое |
| `internal/packet/login/send/send_servers.go` | SendServers: AccountID+SessionID+серверы |
| `internal/repository/session.go` | Redis-токены: создание при авторизации, валидация на геймсервере |
