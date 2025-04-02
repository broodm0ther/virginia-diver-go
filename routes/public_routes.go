package routes

import (
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"strings"
	"time"

	"kostya/database"
	"kostya/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// ✅ Стилизация HTML-письма
func buildHTMLMessage(title, code string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="ru">
		<head>
			<meta charset="UTF-8">
			<title>%s</title>
			<style>
				body {
					background-color: #f2f2f2;
					font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
					margin: 0;
					padding: 0;
					color: #111;
				}
				.wrapper {
					max-width: 480px;
					margin: 30px auto;
					background-color: #ffffff;
					border-radius: 12px;
					box-shadow: 0 5px 15px rgba(0,0,0,0.07);
					padding: 30px;
				}
				.logo {
					font-family: 'Courier New', Courier, monospace;
					font-size: 22px;
					letter-spacing: 1px;
					text-align: center;
					margin-bottom: 25px;
					color: #000;
				}
				.title {
					font-size: 20px;
					font-weight: bold;
					text-align: center;
					margin-bottom: 20px;
				}
				.code {
					font-size: 28px;
					font-weight: bold;
					color: #000;
					background-color: #f7f7f7;
					border-radius: 8px;
					padding: 12px 20px;
					text-align: center;
					letter-spacing: 4px;
					margin: 0 auto 25px;
					width: fit-content;
				}
				.footer {
					margin-top: 30px;
					font-size: 12px;
					color: #999;
					text-align: center;
				}
			</style>
		</head>
		<body>
			<div class="wrapper">
				<div class="logo">amazonica project</div>
				<div class="title">%s</div>
				<div class="code">%s</div>
				<p style="text-align:center;">Введите этот код в приложении для подтверждения.</p>
				<div class="footer">
					Вы получили это письмо, потому что кто-то использовал ваш email<br/>
					Если это были не вы — просто проигнорируйте его.
				</div>
			</div>
		</body>
		</html>
	`, title, title, code)
}

// ✅ Отправка HTML-письма
func sendEmail(to string, subject string, code string) error {
	from := os.Getenv("EMAIL_USER")
	password := os.Getenv("EMAIL_PASS")

	htmlBody := buildHTMLMessage(subject, code)

	msg := "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		htmlBody

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", from, password, "smtp.gmail.com"),
		from,
		[]string{to},
		[]byte(msg),
	)

	if err != nil {
		log.Println("❌ Ошибка отправки:", err)
	} else {
		log.Println("📨 Письмо отправлено на:", to)
	}
	return err
}

// ✅ Регистрация с письмом
func RegisterUser(c *fiber.Ctx) error {
	var newUser models.User
	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные"})
	}

	var existing models.User
	if err := database.DB.Where("email = ?", newUser.Email).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email уже зарегистрирован"})
	}

	newUser.Password = strings.TrimSpace(newUser.Password)
	if newUser.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пароль не может быть пустым"})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка хеширования пароля"})
	}
	newUser.Password = string(hashed)

	code := generateCode()
	newUser.VerifyCode = code
	newUser.IsVerified = false

	sendEmail(newUser.Email, "Подтверждение почты", code)

	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при сохранении"})
	}

	return c.JSON(fiber.Map{"message": "Письмо отправлено на почту"})
}

// 🔐 Подтверждение
func VerifyEmail(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат"})
	}

	var user models.User
	database.DB.Where("email = ?", req.Email).First(&user)

	if user.VerifyCode != req.Code {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный код"})
	}

	user.IsVerified = true
	user.VerifyCode = ""
	database.DB.Save(&user)
	return c.JSON(fiber.Map{"message": "Почта подтверждена"})
}

// 🔑 Логин
func LoginUser(c *fiber.Ctx) error {
	var login struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&login); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат"})
	}

	login.Password = strings.TrimSpace(login.Password)

	var user models.User
	result := database.DB.Where("email = ?", login.Email).First(&user)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Почта не найдена"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Неверный пароль"})
	}

	if !user.IsVerified {
		return c.Status(403).JSON(fiber.Map{"error": "Пожалуйста, подтвердите почту"})
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка токена"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// 🔁 Сброс пароля (запрос кода)
func ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	_ = c.BodyParser(&req)

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Email не найден"})
	}

	code := generateCode()
	user.ResetCode = code
	database.DB.Save(&user)

	sendEmail(user.Email, "Сброс пароля", code)
	return c.JSON(fiber.Map{"message": "Код отправлен на почту"})
}

// 🔐 Сброс пароля по коду
func ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	_ = c.BodyParser(&req)

	var user models.User
	database.DB.Where("email = ?", req.Email).First(&user)

	if user.ResetCode != req.Code {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный код"})
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.Password = string(hashed)
	user.ResetCode = ""
	database.DB.Save(&user)

	return c.JSON(fiber.Map{"message": "Пароль обновлён"})
}

// 🔐 Профиль
func GetProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	return c.JSON(fiber.Map{
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
		"role":     user.Role,
	})
}

// 🔐 Выход
func LogoutUser(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Вы успешно вышли!"})
}

// 🔐 Генерация JWT
func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
