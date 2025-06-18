package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Константы типов пакетов (как в оригинале)
const (
	ConnectionClient        = 1103
	AuthorizationLogin      = 3100
	SendServers             = 3101
	LoginServerError        = 3102
	RefreshServers          = 3115
	RefreshedServers        = 3116
	SelectServer            = 3120
	SelectedServer          = 3121
	CreatePcReq             = 5118
	CompleteCreateCharacter = 5119
)

// Коды ошибок (как в оригинале)
const (
	NoUser             = 267425736
	PasswordWrong      = 2493403489
	NoUserLoginAnother = 2412965047
	IncorrectServer    = 166780708
	NoCharInvalidSlot  = 1234567890
	NoUserCharSlotBusy = 1234567891
)

// Фиксированный ключ Blowfish как в оригинальном сервере
var LoginServerBlowfishKey = []byte{
	0xd5, 0x49, 0x82, 0x55, 0x1d, 0x1a, 0x17, 0x2d, 0xbb, 0x4a, 0x45, 0x43, 0xb7, 0x25, 0xe2, 0x18,
	0xd0, 0x33, 0xd4, 0x08, 0xe6, 0x79, 0x6e, 0x46, 0x2a, 0x1a, 0xef, 0x71, 0xea, 0x34, 0x86, 0x03,
	0xb5, 0x2d, 0x14, 0x17, 0x66, 0x65, 0xfb, 0x91, 0x54, 0x5b, 0x4c, 0x08, 0x5a, 0x11, 0x43, 0x36,
	0x3e, 0xbb, 0x24, 0x41, 0x39, 0x9d, 0x73, 0x30, 0x2c, 0x2e, 0x1d, 0x03, 0x45, 0x48, 0x3c, 0x99,
	0x2c, 0xf9, 0x31, 0xcc, 0x54, 0xae, 0x71, 0x69, 0x70, 0xc7, 0x03, 0x5d, 0xef, 0x2b, 0xe1, 0x19,
	0x3a, 0x35, 0x56, 0x2a, 0x7f, 0x51, 0xbb, 0x74, 0x4e, 0x7c, 0x70, 0x1f, 0x6e, 0x1f, 0x0d, 0x79,
	0xc8, 0x07, 0x11, 0x4e, 0xa9, 0x35, 0xa9, 0x02, 0xe3, 0x19, 0xb1, 0x7e, 0xb3, 0x59, 0xeb, 0x53,
	0xfe, 0x76, 0xd4, 0x28, 0x7e, 0x53, 0x24, 0x73, 0x75, 0x3d, 0x27, 0x0a, 0xcd, 0x6a, 0x98, 0x57,
	0x32, 0x7c, 0xe2, 0x47, 0xbc, 0x72, 0x0d, 0x09, 0xcc, 0x26, 0x00, 0x00, 0x85, 0x1a, 0x48, 0x60,
	0xed, 0x77, 0x5f, 0x3b, 0x55, 0x42, 0x41, 0x28, 0x9d, 0x08, 0x7a, 0x40, 0x5e, 0x5e, 0x38, 0x62,
	0x1b, 0x2a, 0x4d, 0x89, 0x5a, 0xb9, 0x70, 0xf6, 0x3f, 0xc6, 0x76, 0xf2, 0x16, 0x0b, 0x12, 0x98,
	0x51, 0x56, 0x75, 0x18, 0x44, 0x4e, 0x46, 0xeb, 0x5e, 0x45, 0x37, 0xd1, 0x07, 0x45, 0x46, 0xeb,
	0x01, 0x00, 0x94, 0x00, 0xf0, 0x1d,
}

// Ключ для расшифровки пакетов (из оригинального C# кода)
var BlowfishDecryptKey = []byte{
	0x00, 0x00, 0x90, 0x9A, 0xE2, 0xF4, 0x51, 0xBB, 0xB2, 0x13, 0xD6, 0x48, 0x0E, 0xE3, 0x59, 0x04,
	0x07, 0x03, 0xDA, 0x19, 0x47, 0xCF, 0x81, 0xA4, 0x41, 0x37, 0x40, 0xAB, 0xA6, 0xDC, 0xE1, 0x0A,
	0x63, 0x4D, 0x20, 0x53, 0xFD, 0x15, 0xFB, 0x11, 0xF3, 0x79, 0xA1, 0x10, 0xF5, 0x58, 0x38, 0x5C,
	0x69, 0x0B, 0xC6, 0x4A, 0x5A, 0x6E, 0x72, 0x9B, 0x87, 0x1C, 0x7E, 0x82, 0xF8, 0x71, 0x62, 0x14,
	0x6A, 0x39, 0xAF, 0x73, 0x30, 0x86, 0x61, 0x93, 0xB8, 0x05, 0x92, 0x9C, 0x77, 0xE9, 0x6C, 0x0F,
	0x2B, 0x89, 0xDB, 0x6D, 0xA8, 0xA3, 0x24, 0x12, 0xB5, 0x4C, 0x97, 0x02, 0xCE, 0x88, 0x57, 0xDD,
	0xBE, 0x8A, 0x50, 0x6F, 0x7A, 0x2D, 0x8C, 0x3C, 0x22, 0x9F, 0xFA, 0x3E, 0xD3, 0x52, 0xCC, 0x91,
	0xC0, 0x31, 0x08, 0xD0, 0x74, 0xB3, 0x43, 0x46, 0x2C, 0x4B, 0x95, 0x16, 0x9E, 0xB6, 0xB9, 0x00,
	0x5F, 0xB0, 0x1F, 0x8F, 0x25, 0xA5, 0xAC, 0xC7, 0xC4, 0xBC, 0x83, 0x45, 0x99, 0x5B, 0xA2, 0xFC,
	0x34, 0xED, 0x6B, 0x7C, 0xEA, 0xF1, 0xAD, 0x27, 0xFF, 0xB4, 0x26, 0x5D, 0xC5, 0x7B, 0x56, 0xB7,
	0xE6, 0xD7, 0x67, 0xA7, 0x1E, 0x60, 0xC8, 0xA0, 0x80, 0x3F, 0x4F, 0x98, 0x2E, 0x8B, 0x5E, 0x21,
	0xEB, 0x49, 0xCD, 0x0C, 0x3D, 0x1D, 0xBD, 0xD1, 0x64, 0xCA, 0x9D, 0xE8, 0x28, 0xC9, 0xD9, 0x01,
	0xBF, 0xC3, 0xE5, 0xE7, 0x06, 0x96, 0x3A, 0x29, 0x8E, 0x42, 0xF9, 0x8D, 0x94, 0x17, 0x32, 0xDF,
	0x36, 0x1B, 0xCB, 0x7F, 0x1A, 0x33, 0x84, 0x2A, 0x44, 0xF7, 0x0D, 0x7D, 0xE4, 0x35, 0xEC, 0x68,
	0x4E, 0xF6, 0xF0, 0x66, 0x3B, 0x70, 0xE0, 0xA9, 0xD4, 0x76, 0x18, 0xD5, 0x09, 0x2F, 0xD2, 0xC1,
	0xDE, 0xC2, 0x85, 0xB1, 0xF2, 0xEE, 0x54, 0xFE, 0xAE, 0xD8, 0x78, 0x55, 0xBA, 0x23, 0x65, 0xEF,
	0x75, 0xAA, 0x00, 0x00,
}

// Функция расшифровки пакетов (портирована из C# кода)
func BlowfishDecrypt(packet []byte) []byte {
	if len(packet) == 0 {
		return packet
	}

	// Создаем копию ключа для работы
	key := make([]byte, len(BlowfishDecryptKey))
	copy(key, BlowfishDecryptKey)

	// Создаем копию пакета для результата
	result := make([]byte, len(packet))
	copy(result, packet)

	pointerKey := key[0]
	v6 := key[1]

	for i := 0; i < len(result); i++ {
		pointerKey = pointerKey + 1
		v7 := key[pointerKey+2]
		v6 = v7 + v6
		v8 := key[v6+2]
		key[pointerKey+2] = v8
		result[i] ^= key[byte(v7+v8)+2]
		key[v6+2] = v7
	}

	key[1] = v6

	return result
}

// Функция для извлечения null-terminated строки из массива байтов (аналог C# GetText)
func getText(data []byte, offset int) string {
	if offset >= len(data) {
		return ""
	}

	result := make([]byte, 0)
	for i := offset; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		result = append(result, data[i])
	}

	return string(result)
}

// Извлечение логина из пакета авторизации (портировано из C# кода)
func (s *R2Server) extractLogin(data []byte) string {
	if len(data) <= 256 {
		log.Printf("🐛 Недостаточно данных для извлечения логина: %d байт", len(data))
		return ""
	}

	codeLogin := data[256] / 8
	log.Printf("🔍 Код логина: %d (из байта %d)", codeLogin, data[256])

	var offsetLogin int
	switch codeLogin {
	case 0:
		offsetLogin = 151
	case 1:
		offsetLogin = 37
	case 2:
		offsetLogin = 87
	case 3:
		offsetLogin = 336
	case 4:
		offsetLogin = 129
	case 5:
		offsetLogin = 289
	case 6:
		offsetLogin = 172
	case 7:
		offsetLogin = 199
	default:
		offsetLogin = 220
	}

	log.Printf("🔍 Offset для логина: %d", offsetLogin)

	if offsetLogin >= len(data) {
		log.Printf("🐛 Offset логина (%d) больше размера данных (%d)", offsetLogin, len(data))
		return ""
	}

	login := getText(data, offsetLogin)
	log.Printf("🔍 Извлеченный логин: '%s'", login)
	return login
}

// Извлечение пароля из пакета авторизации (портировано из C# кода)
func (s *R2Server) extractPassword(data []byte) string {
	if len(data) <= 81 {
		log.Printf("🐛 Недостаточно данных для извлечения пароля: %d байт", len(data))
		return ""
	}

	codePassword := data[81] / 2
	log.Printf("🔍 Код пароля: %d (из байта %d)", codePassword, data[81])

	var offsetPassword int
	switch codePassword {
	case 0:
		offsetPassword = 260
	case 1:
		offsetPassword = 60
	case 2:
		offsetPassword = 108
	case 3:
		offsetPassword = 4
	case 4:
		offsetPassword = 314
	case 5:
		offsetPassword = 357
	default:
		offsetPassword = 390
	}

	log.Printf("🔍 Offset для пароля: %d", offsetPassword)

	if offsetPassword >= len(data) {
		log.Printf("🐛 Offset пароля (%d) больше размера данных (%d)", offsetPassword, len(data))
		return ""
	}

	password := getText(data, offsetPassword)
	log.Printf("🔍 Извлеченный пароль: '%s'", password)
	return password
}

// Структуры данных
type Account struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Session struct {
	ID        int  `json:"id"`
	AccountId int  `json:"accountId"`
	ServerId  int  `json:"serverId"`
	InGame    bool `json:"inGame"`
}

type Character struct {
	ID      int        `json:"id"`
	Owner   int        `json:"owner"`
	Slot    byte       `json:"slot"`
	Name    string     `json:"name"`
	Class   byte       `json:"class"`
	Sex     byte       `json:"sex"`
	Head    byte       `json:"head"`
	Face    byte       `json:"face"`
	Level   int        `json:"level"`
	HP      int        `json:"hp"`
	MP      int        `json:"mp"`
	PosX    float32    `json:"posX"`
	PosY    float32    `json:"posY"`
	PosZ    float32    `json:"posZ"`
	Str     int        `json:"str"`
	Dex     int        `json:"dex"`
	Int     int        `json:"int"`
	RegDate time.Time  `json:"regDate"`
	DelDate *time.Time `json:"delDate,omitempty"`
}

type ServerModel struct {
	ServerId   int    `json:"serverId"`
	ServerName string `json:"serverName"`
	ServerIP   string `json:"serverIP"`
	ServerPort int    `json:"serverPort"`
	Online     int    `json:"online"`
	MaxOnline  int    `json:"maxOnline"`
	Status     int    `json:"status"`
}

type ClientSession struct {
	conn      net.Conn
	id        string
	account   *Account
	session   *Session
	character *Character
	isAuth    bool
	mutex     sync.Mutex
}

type R2Server struct {
	address    string
	port       int
	clients    map[string]*ClientSession
	accounts   map[string]*Account
	sessions   map[int]*Session
	characters map[int]*Character
	servers    []ServerModel
	mutex      sync.RWMutex
	nextId     int
	nextCharId int
}

func NewR2Server(address string, port int) *R2Server {
	server := &R2Server{
		address:    address,
		port:       port,
		clients:    make(map[string]*ClientSession),
		accounts:   make(map[string]*Account),
		sessions:   make(map[int]*Session),
		characters: make(map[int]*Character),
		servers:    make([]ServerModel, 0),
		nextId:     1,
		nextCharId: 1,
	}
	server.initTestData()
	return server
}

func (s *R2Server) initTestData() {
	// Создаем тестовые аккаунты
	s.accounts["admin"] = &Account{
		ID:       1,
		Login:    "admin",
		Password: "test",
	}

	// Добавляем аккаунт для тестирования с реальным клиентом
	s.accounts["CHINA"] = &Account{
		ID:       2,
		Login:    "CHINA",
		Password: "CHINA",
	}

	// Создаем тестовый сервер
	s.servers = append(s.servers, ServerModel{
		ServerId:   1,
		ServerName: "TestServer",
		ServerIP:   "127.0.0.1",
		ServerPort: 8002,
		Online:     0,
		MaxOnline:  100,
		Status:     1,
	})
}

func (s *R2Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.address, s.port))
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %v", err)
	}
	defer listener.Close()

	log.Printf("🚀 R2 Online Эмулятор запущен на %s:%d", s.address, s.port)
	log.Printf("📡 Ожидаем подключения настоящего клиента R2 Online...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Ошибка принятия соединения: %v", err)
			continue
		}

		clientId := fmt.Sprintf("%s", conn.RemoteAddr().String())
		client := &ClientSession{
			conn: conn,
			id:   clientId,
		}

		s.mutex.Lock()
		s.clients[clientId] = client
		s.mutex.Unlock()

		log.Printf("🔗 Клиент подключен: %s", clientId)
		go s.handleClient(client)
	}
}

func (s *R2Server) handleClient(client *ClientSession) {
	defer func() {
		s.mutex.Lock()
		delete(s.clients, client.id)
		s.mutex.Unlock()
		client.conn.Close()
		log.Printf("🔌 Клиент отключен: %s", client.id)
	}()

	// Отправляем приветственный пакет
	s.sendWelcomePacket(client)

	buffer := make([]byte, 4096)
	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			log.Printf("Ошибка чтения от клиента %s: %v", client.id, err)
			break
		}

		if n > 0 {
			s.processPacket(client, buffer[:n])
		}
	}
}

func (s *R2Server) processPacket(client *ClientSession, data []byte) {
	if len(data) < 6 {
		log.Printf("Слишком короткий пакет от клиента %s", client.id)
		return
	}

	reader := bytes.NewReader(data)
	var packetSize int16
	var cryptFlag byte
	var packetNumber byte
	var packetType int16

	binary.Read(reader, binary.LittleEndian, &packetSize)
	binary.Read(reader, binary.LittleEndian, &cryptFlag)

	log.Printf("📦 Получен пакет от %s: размер=%d, шифрование=%d",
		client.id, packetSize, cryptFlag)

	var packetData []byte

	// Если пакет зашифрован, расшифровываем его (как в оригинальном C# коде)
	if cryptFlag == 1 {
		log.Printf("🔓 Расшифровываем пакет от %s", client.id)
		// Расшифровываем все данные после флага шифрования
		encryptedData := data[3:] // Пропускаем размер пакета (2 байта) + флаг шифрования (1 байт)
		decryptedData := BlowfishDecrypt(encryptedData)

		// Читаем из расшифрованных данных
		reader = bytes.NewReader(decryptedData)
		binary.Read(reader, binary.LittleEndian, &packetNumber)
		binary.Read(reader, binary.LittleEndian, &packetType)
		packetData = decryptedData[3:] // Данные после номера пакета (1 байт) + тип пакета (2 байта)

		log.Printf("🔓 Расшифрован пакет: номер=%d, тип=%d", packetNumber, packetType)
	} else {
		// Незашифрованный пакет - читаем как обычно
		binary.Read(reader, binary.LittleEndian, &packetNumber)
		binary.Read(reader, binary.LittleEndian, &packetType)
		packetData = data[6:]

		log.Printf("📦 Незашифрованный пакет: номер=%d, тип=%d", packetNumber, packetType)
	}

	switch packetType {
	case AuthorizationLogin:
		s.handleAuthorizationLogin(client, packetData)
	case RefreshServers:
		s.handleRefreshServers(client)
	case SelectServer:
		s.handleSelectServer(client, packetData)
	case CreatePcReq:
		s.handleCreateCharacter(client, packetData)
	default:
		log.Printf("❓ Неизвестный тип пакета: %d от клиента %s", packetType, client.id)
	}
}

// Отправка приветственного пакета с фиксированным ключом Blowfish
func (s *R2Server) sendWelcomePacket(client *ClientSession) {
	s.sendBinaryPacket(client, ConnectionClient, LoginServerBlowfishKey)
	log.Printf("✅ Отправлен приветственный пакет (Blowfish ключ) клиенту %s", client.id)
}

// Обработка авторизации (по алгоритму из C# кода)
func (s *R2Server) handleAuthorizationLogin(client *ClientSession, data []byte) {
	if len(data) < 400 {
		log.Printf("❌ Недостаточно данных для авторизации от %s (получено %d байт)", client.id, len(data))
		s.sendLoginError(client, NoUser)
		return
	}

	// Извлекаем логин по алгоритму из C# кода
	login := s.extractLogin(data)
	password := s.extractPassword(data)

	log.Printf("🔑 Попытка авторизации: логин='%s', пароль='%s' от клиента %s", login, password, client.id)

	s.mutex.RLock()
	account, exists := s.accounts[login]
	s.mutex.RUnlock()

	if !exists {
		log.Printf("❌ Аккаунт не найден: %s", login)
		s.sendLoginError(client, NoUser)
		return
	}

	if account.Password != password {
		log.Printf("❌ Неверный пароль для аккаунта: %s", login)
		s.sendLoginError(client, PasswordWrong)
		return
	}

	// Создаем сессию
	s.mutex.Lock()
	session := &Session{
		ID:        s.nextId,
		AccountId: account.ID,
		ServerId:  0,
		InGame:    false,
	}
	s.sessions[session.ID] = session
	s.nextId++
	client.account = account
	client.session = session
	client.isAuth = true
	s.mutex.Unlock()

	log.Printf("✅ Авторизация успешна для %s (аккаунт ID: %d, сессия ID: %d)", login, account.ID, session.ID)

	// Отправляем список серверов
	s.sendServersList(client)
}

// Отправка ошибки авторизации
func (s *R2Server) sendLoginError(client *ClientSession, errorType int) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, uint32(errorType))
	s.sendBinaryPacket(client, LoginServerError, buf.Bytes())
}

// Отправка списка серверов
func (s *R2Server) sendServersList(client *ClientSession) {
	buf := &bytes.Buffer{}

	// Простой формат списка серверов
	binary.Write(buf, binary.LittleEndian, int32(client.account.ID))
	binary.Write(buf, binary.LittleEndian, int32(client.session.ID))
	binary.Write(buf, binary.LittleEndian, int32(len(s.servers)))

	for _, server := range s.servers {
		binary.Write(buf, binary.LittleEndian, int32(server.ServerId))
		// Записываем имя сервера (фиксированная длина 20 байт)
		nameBytes := make([]byte, 20)
		copy(nameBytes, server.ServerName)
		buf.Write(nameBytes)
		// IP адрес (фиксированная длина 16 байт)
		ipBytes := make([]byte, 16)
		copy(ipBytes, server.ServerIP)
		buf.Write(ipBytes)
		binary.Write(buf, binary.LittleEndian, int32(server.ServerPort))
		binary.Write(buf, binary.LittleEndian, int32(server.Online))
		binary.Write(buf, binary.LittleEndian, int32(server.MaxOnline))
		binary.Write(buf, binary.LittleEndian, int32(server.Status))
	}

	s.sendBinaryPacket(client, SendServers, buf.Bytes())
	log.Printf("📋 Отправлен список серверов клиенту %s", client.id)
}

// Обработка обновления списка серверов
func (s *R2Server) handleRefreshServers(client *ClientSession) {
	log.Printf("🔄 Запрос обновления списка серверов от клиента %s", client.id)
	s.sendServersList(client)
}

// Обработка выбора сервера
func (s *R2Server) handleSelectServer(client *ClientSession, data []byte) {
	if !client.isAuth {
		log.Printf("❌ Попытка выбора сервера без авторизации от %s", client.id)
		return
	}

	serverId := 1
	log.Printf("🎯 Выбор сервера %d клиентом %s", serverId, client.id)

	s.mutex.Lock()
	client.session.ServerId = serverId
	s.mutex.Unlock()

	// Подтверждаем выбор сервера
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, int32(0)) // Успех

	s.sendBinaryPacket(client, SelectedServer, buf.Bytes())
	log.Printf("✅ Сервер %d выбран клиентом %s", serverId, client.id)
}

// Обработка создания персонажа
func (s *R2Server) handleCreateCharacter(client *ClientSession, data []byte) {
	if !client.isAuth {
		log.Printf("❌ Попытка создания персонажа без авторизации от %s", client.id)
		return
	}

	// Примитивный парсинг данных персонажа
	if len(data) < 6 {
		log.Printf("❌ Недостаточно данных для создания персонажа от %s", client.id)
		return
	}

	slot := data[0]
	class := data[1]
	sex := data[2]
	head := data[3]
	face := data[4]
	_ = data[5] // typeBody - игнорируем пока что

	name := fmt.Sprintf("Char_%d_%d", client.account.ID, slot)

	log.Printf("👤 Создание персонажа: слот=%d, класс=%d, пол=%d, имя=%s клиентом %s",
		slot, class, sex, name, client.id)

	// Создаем персонажа
	character := &Character{
		ID:      s.nextCharId,
		Owner:   client.account.ID,
		Slot:    slot,
		Name:    name,
		Class:   class,
		Sex:     sex,
		Head:    head,
		Face:    face,
		Level:   1,
		HP:      93,
		MP:      51,
		PosX:    364000.2,
		PosY:    313483.7,
		PosZ:    12339.71,
		RegDate: time.Now(),
	}

	// Устанавливаем базовые характеристики в зависимости от класса
	switch class {
	case 0: // Fighter/Knight
		character.Str = 15
		character.Dex = 10
		character.Int = 10
	case 1: // Dragoon/Ranger
		character.Str = 10
		character.Dex = 15
		character.Int = 10
	case 2: // Wizard
		character.Str = 10
		character.Dex = 10
		character.Int = 15
	default:
		character.Str = 12
		character.Dex = 12
		character.Int = 12
	}

	s.mutex.Lock()
	s.characters[character.ID] = character
	s.nextCharId++
	s.mutex.Unlock()

	log.Printf("✅ Персонаж создан: ID=%d, имя=%s, класс=%d для аккаунта %d",
		character.ID, character.Name, character.Class, client.account.ID)

	// Отправляем подтверждение создания персонажа
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, int32(character.ID))
	binary.Write(buf, binary.LittleEndian, int32(character.Str))
	binary.Write(buf, binary.LittleEndian, int32(character.Dex))
	binary.Write(buf, binary.LittleEndian, int32(character.Int))

	s.sendBinaryPacket(client, CompleteCreateCharacter, buf.Bytes())
}

// Отправка бинарного пакета
func (s *R2Server) sendBinaryPacket(client *ClientSession, packetType int16, data []byte) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	// Создаем буфер для пакета
	buffer := &bytes.Buffer{}

	// Записываем заголовок пакета
	totalLen := int16(len(data) + 6) // 2 (size) + 1 (crypt) + 1 (number) + 2 (type)

	binary.Write(buffer, binary.LittleEndian, totalLen)   // Размер пакета
	binary.Write(buffer, binary.LittleEndian, byte(0))    // Флаг шифрования
	binary.Write(buffer, binary.LittleEndian, byte(1))    // Номер пакета
	binary.Write(buffer, binary.LittleEndian, packetType) // Тип пакета
	buffer.Write(data)                                    // Данные пакета

	// Отправляем пакет
	_, err := client.conn.Write(buffer.Bytes())
	if err != nil {
		log.Printf("❌ Ошибка отправки пакета клиенту %s: %v", client.id, err)
	}
}

func (s *R2Server) Shutdown() {
	log.Println("⏹️  Завершение работы сервера...")

	// Закрываем все соединения
	s.mutex.Lock()
	for _, client := range s.clients {
		client.conn.Close()
	}
	s.mutex.Unlock()

	log.Println("🔚 Сервер остановлен")
}

func main() {
	// Создаем сервер
	server := NewR2Server("127.0.0.1", 8001)

	log.Println("🎮 === R2 Online Сервер Эмулятор (Бинарный протокол) ===")
	log.Println("🔑 Данные для входа: логин=admin, пароль=test")
	log.Println("🌐 Подключайтесь настоящим клиентом R2 Online!")
	log.Println("📡 Поддерживается:")
	log.Println("   ✓ Приветственный пакет с Blowfish ключом")
	log.Println("   ✓ Авторизация")
	log.Println("   ✓ Список серверов")
	log.Println("   ✓ Выбор сервера")
	log.Println("   ✓ Создание персонажей")
	log.Println("")

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("💥 Ошибка запуска сервера: %v", err)
	}
}
