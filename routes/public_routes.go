package routes

import (
	"log"
	"strings"
	"time"

	"kostya/database"
	"kostya/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Регистрация пользователя (с хешированием пароля)
func RegisterUser(c *fiber.Ctx) error {
	var newUser models.User

	if err := c.BodyParser(&newUser); err != nil {
		log.Println("❌ Ошибка парсинга JSON:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные"})
	}

	// Проверяем, существует ли уже такой пользователь
	var existingUser models.User
	if err := database.DB.Where("email = ?", newUser.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email уже зарегистрирован"})
	}

	// Убираем лишние пробелы
	newUser.Password = strings.TrimSpace(newUser.Password)

	// Проверяем, что пароль не пустой
	if newUser.Password == "" {
		log.Println("❌ Ошибка: Пароль не может быть пустым!")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пароль не может быть пустым"})
	}

	// Хешируем пароль перед сохранением
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("❌ Ошибка хеширования пароля:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка создания пароля"})
	}
	newUser.Password = string(hashedPassword)

	log.Println("✅ Захешированный пароль перед сохранением:", newUser.Password)

	// Сохраняем пользователя в базе
	if err := database.DB.Create(&newUser).Error; err != nil {
		log.Println("❌ Ошибка сохранения пользователя:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при регистрации пользователя"})
	}

	log.Println("✅ Пользователь зарегистрирован:", newUser.Email)
	return c.JSON(fiber.Map{"message": "Регистрация успешна!"})
}

func LoginUser(c *fiber.Ctx) error {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginData); err != nil {
		log.Println("❌ Ошибка парсинга JSON:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные"})
	}

	// Убираем пробелы в пароле
	loginData.Password = strings.TrimSpace(loginData.Password)

	var user models.User
	result := database.DB.Where("email = ?", loginData.Email).First(&user)

	// Проверяем, существует ли email
	if result.Error != nil {
		log.Println("❌ Почта не зарегистрирована:", loginData.Email)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Данная почта не зарегистрирована"})
	}

	log.Println("🔑 Введённый пароль:", loginData.Password)
	log.Println("🔑 Хеш пароля из базы:", user.Password)

	// Проверяем пароль через bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		log.Println("❌ Ошибка: Неправильный пароль")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неправильный пароль"})
	}

	// Генерируем токен
	token, err := GenerateToken(user.ID)
	if err != nil {
		log.Println("❌ Ошибка создания токена:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при создании токена"})
	}

	log.Println("✅ Пользователь вошёл:", user.Email)

	// ✅ Теперь возвращаем и `token`, и `user`
	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Функция генерации JWT токена
func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Получение профиля пользователя
func GetProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)

	return c.JSON(fiber.Map{
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
	})
}

func LogoutUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Вы успешно вышли!"})
}
