package routes

import (
	"encoding/json"
	"kostya/database"
	"kostya/models"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func AddProduct(c *fiber.Ctx) error {
	log.Println("📥 Запрос на добавление товара")

	form, err := c.MultipartForm()
	if err != nil {
		log.Println("❌ Ошибка парсинга формы:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат формы"})
	}

	title := form.Value["title"]
	description := form.Value["description"]
	priceStr := form.Value["price"]
	location := form.Value["location"]
	category := form.Value["category"]
	gender := form.Value["gender"]
	size := form.Value["size"]

	log.Println("🔎 Данные:", title, description, priceStr, location, category, gender, size)

	if len(title) == 0 || len(description) == 0 || len(priceStr) == 0 ||
		len(location) == 0 || len(category) == 0 || len(gender) == 0 || len(size) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Все поля обязательны"})
	}

	price, err := strconv.ParseFloat(priceStr[0], 64)
	if err != nil {
		log.Println("❌ Ошибка преобразования цены:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Неверная цена"})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Минимум 1 изображение"})
	}

	if len(files) > 10 {
		return c.Status(400).JSON(fiber.Map{"error": "Максимум 10 изображений"})
	}

	imagePaths := []string{}
	err = os.MkdirAll("./uploads/products", os.ModePerm)
	if err != nil {
		log.Println("❌ Ошибка создания директории uploads/products:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка на сервере"})
	}

	for _, file := range files {
		path := filepath.Join("uploads/products", file.Filename)
		err := c.SaveFile(file, path)
		if err != nil {
			log.Println("❌ Ошибка сохранения файла:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Ошибка загрузки файлов"})
		}
		imagePaths = append(imagePaths, "/"+path)
	}

	user, ok := c.Locals("user").(models.User)
	if !ok {
		log.Println("❌ Не удалось получить пользователя из контекста")
		return c.Status(401).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	jsonImages, err := json.Marshal(imagePaths)
	if err != nil {
		log.Println("❌ Ошибка сериализации изображений:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка обработки изображений"})
	}

	product := models.Product{
		Title:       title[0],
		Description: description[0],
		Price:       price,
		Location:    location[0],
		Category:    category[0],
		Gender:      gender[0],
		Size:        size[0],
		Images:      string(jsonImages),
		Status:      "pending",
		UserID:      user.ID,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		log.Println("❌ Ошибка создания товара в БД:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка создания товара", "details": err.Error()})
	}

	log.Println("✅ Товар создан с ID:", product.ID)
	return c.JSON(fiber.Map{"message": "Товар на проверке", "product_id": product.ID})
}
