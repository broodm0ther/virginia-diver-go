// routes/product_admin.go
package routes

import (
	"kostya/database"
	"kostya/models"

	"github.com/gofiber/fiber/v2"
)

// Получение товаров со статусом pending для администратора
func GetPendingProducts(c *fiber.Ctx) error {
	admin := c.Locals("user").(models.User)
	if admin.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var products []models.Product
	if err := database.DB.Where("status = ?", "pending").Find(&products).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при получении объявлений"})
	}

	return c.JSON(fiber.Map{"products": products})
}

// Подтверждение товара
func ApproveProduct(c *fiber.Ctx) error {
	admin := c.Locals("user").(models.User)
	if admin.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	id := c.Params("id")
	if err := database.DB.Model(&models.Product{}).Where("id = ?", id).Update("status", "approved").Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка подтверждения"})
	}

	return c.JSON(fiber.Map{"message": "Товар подтверждён"})
}

// Отклонение товара
func RejectProduct(c *fiber.Ctx) error {
	admin := c.Locals("user").(models.User)
	if admin.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	id := c.Params("id")
	if err := database.DB.Model(&models.Product{}).Where("id = ?", id).Update("status", "rejected").Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка отклонения"})
	}

	return c.JSON(fiber.Map{"message": "Товар отклонён"})
}
