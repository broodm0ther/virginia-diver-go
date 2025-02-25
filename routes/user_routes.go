package routes

import (
	"fmt"
	"os"
	"path/filepath"

	"kostya/database"
	"kostya/models"

	"github.com/gofiber/fiber/v2"
)

// ✅ Используем уже объявленный `AuthMiddleware`
func UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	username := c.FormValue("username")
	if username != "" {
		user.Username = username
	}

	// Загрузка аватара
	file, err := c.FormFile("avatar")
	if err == nil {
		uploadDir := "uploads/"
		os.MkdirAll(uploadDir, os.ModePerm)

		fileExt := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("%d%s", user.ID, fileExt)
		filePath := filepath.Join(uploadDir, fileName)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки аватарки"})
		}

		user.Avatar = "/" + filePath
	}

	database.DB.Save(&user)
	return c.JSON(fiber.Map{"message": "Профиль обновлен!", "user": user})
}
