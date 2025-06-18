// Тестовый клиент для R2 Online эмулятора
// Для запуска: cd client_test && go run main.go
// Убедитесь, что сервер запущен перед тестированием
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// Структуры пакетов (клиентские версии)
type ClientConnectionPacket struct {
	DecryptKey []byte
}

type ClientAuthPacket struct {
	Login    string
	Password string
}

type ClientServersPacket struct {
	AccountId int            `json:"accountId"`
	SessionId int            `json:"sessionId"`
	Servers   []ClientServer `json:"servers"`
}

type ClientServer struct {
	ServerId   int    `json:"serverId"`
	ServerName string `json:"serverName"`
	ServerIP   string `json:"serverIP"`
	ServerPort int    `json:"serverPort"`
	Online     int    `json:"online"`
	MaxOnline  int    `json:"maxOnline"`
	Status     int    `json:"status"`
}

type ClientCreateCharPacket struct {
	Slot     byte   `json:"slot"`
	Class    byte   `json:"class"`
	Sex      byte   `json:"sex"`
	Head     byte   `json:"head"`
	Face     byte   `json:"face"`
	TypeBody byte   `json:"typeBody"`
	Name     string `json:"name"`
}

type ClientCharCreatedPacket struct {
	CharacterId int `json:"characterId"`
	Str         int `json:"str"`
	Dex         int `json:"dex"`
	Int         int `json:"int"`
}

const (
	// Константы пакетов
	ConnClient         = 1103
	AuthLogin          = 3100
	SendSvrs           = 3101
	LoginSvrError      = 3102
	SelectSvr          = 3120
	SelectedSvr        = 3121
	CreateChar         = 5118
	CompleteCreateChar = 5119
)

type TestClient struct {
	conn net.Conn
}

func NewTestClient() *TestClient {
	return &TestClient{}
}

func (c *TestClient) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}
	c.conn = conn
	return nil
}

func (c *TestClient) Disconnect() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *TestClient) sendPacket(packetType int16, data interface{}) error {
	// Сериализуем данные в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %v", err)
	}

	// Создаем буфер для пакета
	buffer := &bytes.Buffer{}

	// Записываем заголовок пакета
	packetDataLen := int16(len(jsonData))
	totalLen := packetDataLen + 6 // 2 (size) + 1 (crypt) + 1 (number) + 2 (type)

	binary.Write(buffer, binary.LittleEndian, totalLen)   // Размер пакета
	binary.Write(buffer, binary.LittleEndian, byte(0))    // Флаг шифрования
	binary.Write(buffer, binary.LittleEndian, byte(1))    // Номер пакета
	binary.Write(buffer, binary.LittleEndian, packetType) // Тип пакета
	buffer.Write(jsonData)                                // Данные пакета

	// Отправляем пакет
	_, err = c.conn.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("ошибка отправки пакета: %v", err)
	}

	log.Printf("Отправлен пакет типа %d с данными: %s", packetType, string(jsonData))
	return nil
}

func (c *TestClient) readPacket() (int16, []byte, error) {
	// Читаем заголовок пакета
	header := make([]byte, 6)
	n, err := c.conn.Read(header)
	if err != nil {
		return 0, nil, fmt.Errorf("ошибка чтения заголовка: %v", err)
	}
	if n < 6 {
		return 0, nil, fmt.Errorf("неполный заголовок пакета")
	}

	// Парсим заголовок
	reader := bytes.NewReader(header)
	var size int16
	var isCrypt byte
	var number byte
	var packetType int16

	binary.Read(reader, binary.LittleEndian, &size)
	binary.Read(reader, binary.LittleEndian, &isCrypt)
	binary.Read(reader, binary.LittleEndian, &number)
	binary.Read(reader, binary.LittleEndian, &packetType)

	// Читаем данные пакета
	dataSize := size - 6
	if dataSize > 0 {
		data := make([]byte, dataSize)
		n, err := c.conn.Read(data)
		if err != nil {
			return packetType, nil, fmt.Errorf("ошибка чтения данных пакета: %v", err)
		}
		if int16(n) < dataSize {
			return packetType, nil, fmt.Errorf("неполные данные пакета")
		}
		return packetType, data, nil
	}

	return packetType, nil, nil
}

func (c *TestClient) performLogin() error {
	log.Println("Выполняем авторизацию...")

	loginPacket := ClientAuthPacket{
		Login:    "admin",
		Password: "test",
	}

	return c.sendPacket(AuthLogin, loginPacket)
}

func (c *TestClient) selectServer(serverId int) error {
	log.Printf("Выбираем сервер %d...", serverId)

	selectPacket := map[string]interface{}{
		"accountId": 1,
		"serverId":  serverId,
		"login":     "admin",
	}

	return c.sendPacket(SelectSvr, selectPacket)
}

func (c *TestClient) createCharacter(slot byte, class byte) error {
	log.Printf("Создаем персонажа в слоте %d класса %d...", slot, class)

	// Упрощенная отправка данных для создания персонажа
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.LittleEndian, slot)    // Слот
	binary.Write(buffer, binary.LittleEndian, class)   // Класс
	binary.Write(buffer, binary.LittleEndian, byte(1)) // Пол (мужской)
	binary.Write(buffer, binary.LittleEndian, byte(1)) // Голова
	binary.Write(buffer, binary.LittleEndian, byte(1)) // Лицо
	binary.Write(buffer, binary.LittleEndian, byte(1)) // Тип тела

	// Отправляем как сырые байты вместо JSON
	totalLen := int16(buffer.Len() + 6)
	packetBuffer := &bytes.Buffer{}

	binary.Write(packetBuffer, binary.LittleEndian, totalLen)
	binary.Write(packetBuffer, binary.LittleEndian, byte(0))
	binary.Write(packetBuffer, binary.LittleEndian, byte(1))
	binary.Write(packetBuffer, binary.LittleEndian, int16(CreateChar))
	packetBuffer.Write(buffer.Bytes())

	_, err := c.conn.Write(packetBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("ошибка отправки пакета создания персонажа: %v", err)
	}

	log.Printf("Отправлен пакет создания персонажа")
	return nil
}

func (c *TestClient) runTest() {
	log.Println("=== Запуск тестового клиента R2 Online ===")

	// Подключаемся к серверу
	log.Println("Подключение к серверу...")
	if err := c.Connect("127.0.0.1:8001"); err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer c.Disconnect()

	log.Println("Подключение успешно!")

	// Ждем и читаем приветственный пакет
	log.Println("Ожидание приветственного пакета...")
	packetType, data, err := c.readPacket()
	if err != nil {
		log.Printf("Ошибка чтения приветственного пакета: %v", err)
	} else {
		log.Printf("Получен пакет типа %d, размер данных: %d", packetType, len(data))
		if packetType == ConnClient {
			log.Println("✓ Приветственный пакет получен!")
		}
	}

	// Небольшая пауза
	time.Sleep(1 * time.Second)

	// Выполняем авторизацию
	if err := c.performLogin(); err != nil {
		log.Printf("Ошибка авторизации: %v", err)
		return
	}

	// Читаем ответ на авторизацию
	log.Println("Ожидание ответа на авторизацию...")
	packetType, data, err = c.readPacket()
	if err != nil {
		log.Printf("Ошибка чтения ответа на авторизацию: %v", err)
	} else {
		log.Printf("Получен пакет типа %d", packetType)
		if packetType == SendSvrs {
			log.Println("✓ Список серверов получен!")

			// Парсим список серверов
			var serversPacket ClientServersPacket
			if err := json.Unmarshal(data, &serversPacket); err == nil {
				log.Printf("Аккаунт ID: %d, Сессия ID: %d", serversPacket.AccountId, serversPacket.SessionId)
				log.Printf("Доступно серверов: %d", len(serversPacket.Servers))
				for i, server := range serversPacket.Servers {
					log.Printf("  %d. %s (%s:%d) - Онлайн: %d/%d",
						i+1, server.ServerName, server.ServerIP, server.ServerPort,
						server.Online, server.MaxOnline)
				}
			}
		} else if packetType == LoginSvrError {
			log.Println("❌ Ошибка авторизации!")
			return
		}
	}

	// Небольшая пауза
	time.Sleep(1 * time.Second)

	// Выбираем сервер
	if err := c.selectServer(1); err != nil {
		log.Printf("Ошибка выбора сервера: %v", err)
		return
	}

	// Читаем подтверждение выбора сервера
	log.Println("Ожидание подтверждения выбора сервера...")
	packetType, data, err = c.readPacket()
	if err != nil {
		log.Printf("Ошибка чтения подтверждения: %v", err)
	} else {
		log.Printf("Получен пакет типа %d", packetType)
		if packetType == SelectedSvr {
			log.Println("✓ Сервер выбран!")
		}
	}

	// Небольшая пауза
	time.Sleep(1 * time.Second)

	// Создаем персонажа
	if err := c.createCharacter(0, 0); err != nil {
		log.Printf("Ошибка создания персонажа: %v", err)
		return
	}

	// Читаем подтверждение создания персонажа
	log.Println("Ожидание подтверждения создания персонажа...")
	packetType, data, err = c.readPacket()
	if err != nil {
		log.Printf("Ошибка чтения подтверждения: %v", err)
	} else {
		log.Printf("Получен пакет типа %d", packetType)
		if packetType == CompleteCreateChar {
			log.Println("✓ Персонаж создан!")

			// Парсим данные персонажа
			var characterPacket ClientCharCreatedPacket
			if err := json.Unmarshal(data, &characterPacket); err == nil {
				log.Printf("ID персонажа: %d", characterPacket.CharacterId)
				log.Printf("Характеристики - Сила: %d, Ловкость: %d, Интеллект: %d",
					characterPacket.Str, characterPacket.Dex, characterPacket.Int)
			}
		}
	}

	log.Println("=== Тест завершен ===")
}

func main() {
	client := NewTestClient()
	client.runTest()
}
