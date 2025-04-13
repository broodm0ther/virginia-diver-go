// 📁 routes/product_admin.go
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

func GetApprovedProducts(c *fiber.Ctx) error {
	var products []models.Product
	if err := database.DB.Preload("User").Where("status = ?", "approved").Find(&products).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка получения товаров"})
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

func DeleteProduct(c *fiber.Ctx) error {
	admin := c.Locals("user").(models.User)
	if admin.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID не передан"})
	}

	type ReasonBody struct {
		Reason string `json:"reason"`
	}
	var body ReasonBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Невалидный запрос"})
	}

	if body.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Укажите причину удаления"})
	}

	var product models.Product
	if err := database.DB.First(&product, "id = ?", idParam).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Товар не найден"})
	}

	if err := database.DB.Delete(&product).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка удаления товара"})
	}

	return c.JSON(fiber.Map{"message": "Объявление удалено по причине: " + body.Reason})
}
