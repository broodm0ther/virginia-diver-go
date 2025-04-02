package routes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kostya/database"
	"kostya/models"

	"github.com/gofiber/fiber/v2"
)

// ✅ Обновление профиля
func UpdateProfile(c *fiber.Ctx) error {
	fmt.Println("📥 Получен запрос на обновление профиля")

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ошибка парсинга формы"})
	}

	username := form.Value["username"][0]
	region := form.Value["region"][0]
	bio := form.Value["bio"][0]

	var avatarPath string
	files := form.File["avatar"]
	if len(files) > 0 {
		file := files[0]
		fmt.Println("📸 Файл получен:", file.Filename)

		os.MkdirAll("./uploads", os.ModePerm)
		avatarPath = "/uploads/" + file.Filename
		filePath := filepath.Join("./uploads", file.Filename)

		err = c.SaveFile(file, filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка сохранения файла"})
		}
		fmt.Println("✅ Файл успешно сохранён:", filePath)
	}

	userID := c.Locals("user_id").(int)
	query := "UPDATE users SET username=?, region=?, bio=?, avatar=? WHERE id=?"
	err = database.DB.Exec(query, username, region, bio, avatarPath, userID).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления профиля"})
	}

	fmt.Println("✅ Данные профиля обновлены в БД")
	return c.JSON(fiber.Map{"message": "Профиль обновлён", "avatar": avatarPath})
}

// ✅ Установить роль (только админ)
func SetUserRole(c *fiber.Ctx) error {
	adminID := c.Locals("user_id").(int)

	var adminUser models.User
	if err := database.DB.First(&adminUser, adminID).Error; err != nil || adminUser.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	type RequestBody struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный JSON"})
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", body.UserID).Update("role", body.Role).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления роли"})
	}

	return c.JSON(fiber.Map{"message": "Роль успешно обновлена"})
}

// ✅ Публично: получить всех пользователей
func GetAllUsersPublic(c *fiber.Ctx) error {
	search := c.Query("search", "")
	var users []models.User

	query := database.DB.Model(&models.User{})

	if search != "" {
		query = query.Where("username ILIKE ?", "%"+search+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения пользователей"})
	}

	var cleanUsers []map[string]interface{}
	for _, u := range users {
		cleanUsers = append(cleanUsers, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"avatar":   u.Avatar,
		})
	}

	return c.JSON(cleanUsers)
}

// ✅ Только для админа: получить всех пользователей
func GetAllUsers(c *fiber.Ctx) error {
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	userID, ok := userIDInterface.(int)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения ID пользователя"})
	}

	var adminUser models.User
	if err := database.DB.First(&adminUser, userID).Error; err != nil || adminUser.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения пользователей"})
	}

	var cleanUsers []map[string]interface{}
	for _, u := range users {
		cleanUsers = append(cleanUsers, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"role":     u.Role,
		})
	}

	return c.JSON(fiber.Map{"users": cleanUsers})
}

// ✅ Получить всех пользователей, с кем был чат
func GetUserChatPartners(c *fiber.Ctx) error {
	user := c.Query("user", "")
	if user == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Не передан юзер"})
	}

	var messages []models.ChatMessage
	if err := database.DB.
		Where("\"user\" = ? OR room LIKE ?", user, "%"+user+"%").
		Order("created_at desc").
		Find(&messages).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при получении чатов"})
	}

	unique := map[string]bool{}
	users := []string{}

	for _, msg := range messages {
		parts := strings.Split(msg.Room, "_")
		for _, u := range parts {
			if u != user && !unique[u] {
				unique[u] = true
				users = append(users, u)
			}
		}
	}

	if len(users) == 0 {
		// 🛠 Возвращаем ПУСТОЙ массив, а не null
		return c.JSON([]interface{}{})
	}

	var found []models.User
	if err := database.DB.Where("username IN ?", users).Find(&found).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при получении пользователей"})
	}

	var response []map[string]interface{}
	for _, u := range found {
		response = append(response, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"avatar":   u.Avatar,
		})
	}

	return c.JSON(response)
}
