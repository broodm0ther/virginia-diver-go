package routes

import (
	"fmt"
	"os"
	"path/filepath"

	"kostya/database"

	"github.com/gofiber/fiber/v2"
)

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

	// 🔥 Обновляем профиль пользователя в базе данных
	userID := c.Locals("user_id").(int) // ID пользователя из middleware
	query := "UPDATE users SET username=?, region=?, bio=?, avatar=? WHERE id=?"
	err = database.DB.Exec(query, username, region, bio, avatarPath, userID).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка обновления профиля"})
	}

	fmt.Println("✅ Данные профиля обновлены в БД")
	return c.JSON(fiber.Map{"message": "Профиль обновлён", "avatar": avatarPath})
}
