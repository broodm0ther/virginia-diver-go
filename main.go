package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"kostya/database"
	"kostya/routes"
)

func main() {
	// 📥 Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env не найден или не загружен")
	}

	// 📡 Подключение к базе данных
	database.ConnectDatabase()

	app := fiber.New()

	// 🖼 Статические файлы
	app.Static("/uploads/", "./uploads")

	// 📦 Группа авторизации и аккаунта
	authGroup := app.Group("/api/auth")
	authGroup.Post("/login", routes.LoginUser)
	authGroup.Post("/register", routes.RegisterUser)
	authGroup.Post("/verify-email", routes.VerifyEmail)
	authGroup.Post("/forgot-password", routes.ForgotPassword)
	authGroup.Post("/reset-password", routes.ResetPassword)
	authGroup.Get("/profile", routes.AuthMiddleware(), routes.GetProfile)
	authGroup.Post("/update-profile", routes.AuthMiddleware(), routes.UpdateProfile)
	authGroup.Post("/logout", routes.LogoutUser)
	authGroup.Post("/set-role", routes.AuthMiddleware(), routes.SetUserRole)
	authGroup.Get("/all-users", routes.AuthMiddleware(), routes.GetAllUsers)

	// 📦 Группа товаров
	productGroup := app.Group("/api/products", routes.AuthMiddleware())
	productGroup.Get("/pending", routes.GetPendingProducts)
	productGroup.Post("/approve/:id", routes.ApproveProduct)
	productGroup.Post("/reject/:id", routes.RejectProduct)

	// ✅ Одобренные товары — доступны всем
	app.Get("/api/products/approved", routes.GetApprovedProducts)

	log.Println("🚀 Сервер запущен на порту 8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("❌ Ошибка при запуске сервера: %v", err)
	}
}
