package main

import (
	"log"

	"kostya/database"
	"kostya/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectDatabase()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🚀 API работает!")
	})

	authGroup := app.Group("/api/auth")
	authGroup.Post("/login", routes.LoginUser)
	authGroup.Post("/register", routes.RegisterUser)
	authGroup.Get("/profile", routes.AuthMiddleware(), routes.GetProfile)
	authGroup.Post("/logout", routes.LogoutUser)

	log.Println("📌 Маршруты зарегистрированы:")
	log.Println("➡️ POST /api/auth/register")
	log.Println("➡️ POST /api/auth/login")

	log.Println("🚀 Сервер запущен на порту 8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("❌ Ошибка при запуске сервера: %v", err)
	}
}
