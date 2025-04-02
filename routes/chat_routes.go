package routes

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"kostya/database"
	"kostya/models"
)

type Client struct {
	Conn *websocket.Conn
	User string
	Room string
}

var (
	clients   = make(map[*websocket.Conn]Client)
	roomMutex sync.Mutex
)

func SetupChatRoutes(app *fiber.App) {
	// 📡 WebSocket
	app.Get("/ws/chat", websocket.New(func(c *websocket.Conn) {
		user := c.Query("user", "Unknown")
		room := c.Query("room", "global")

		roomMutex.Lock()
		clients[c] = Client{Conn: c, User: user, Room: room}
		roomMutex.Unlock()

		defer func() {
			roomMutex.Lock()
			delete(clients, c)
			roomMutex.Unlock()
			c.Close()
		}()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("🔌 WebSocket ReadMessage error:", err)
				break
			}

			log.Printf("💬 [%s/%s]: %s", room, user, string(msg))

			message := models.ChatMessage{
				User:      user,
				Room:      room,
				Text:      string(msg),
				CreatedAt: time.Now(),
			}

			// 💾 Сохраняем в БД
			if err := database.DB.Create(&message).Error; err != nil {
				log.Println("❌ Ошибка при сохранении сообщения:", err)
				continue
			}

			// 📤 Рассылаем в комнату
			roomMutex.Lock()
			for conn, client := range clients {
				if client.Room == room {
					conn.WriteJSON(fiber.Map{
						"id":         message.ID,
						"text":       message.Text,
						"user":       message.User,
						"room":       message.Room,
						"created_at": message.CreatedAt,
					})
				}
			}
			roomMutex.Unlock()
		}
	}))

	// 📥 История чата
	app.Get("/api/chat/history/:roomId", func(c *fiber.Ctx) error {
		roomId := c.Params("roomId")
		user := c.Query("user", "")

		if user == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Не передан user"})
		}

		// Проверяем, что user является участником комнаты
		parts := strings.Split(roomId, "_")
		if len(parts) != 2 || (parts[0] != user && parts[1] != user) {
			return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
		}

		var messages []models.ChatMessage
		if err := database.DB.
			Where("room = ?", roomId).
			Order("created_at asc").
			Find(&messages).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка при получении истории"})
		}

		return c.JSON(messages)
	})

}
