package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"kostya/database"
	"kostya/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env не найден или не загружен")
	}
	database.ConnectDatabase()

	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// 🖼 Статические файлы
	app.Static("/uploads/", "./uploads")

	// 📦 WebSocket чат
	routes.SetupChatRoutes(app)

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
	authGroup.Get("/all-users", routes.AuthMiddleware(), routes.GetAllUsers) // 🔒 для админки

	// 📦 Группа товаров
	// 📦 Публичный маршрут для approved товаров (гостям доступен)
	app.Get("/api/products/approved", routes.GetApprovedProducts)

	productGroup := app.Group("/api/products", routes.AuthMiddleware())
	productGroup.Get("/pending", routes.GetPendingProducts)
	productGroup.Post("/approve/:id", routes.ApproveProduct)
	productGroup.Post("/reject/:id", routes.RejectProduct)
	productGroup.Post("/delete/:id", routes.DeleteProduct)
	productGroup.Post("/add", routes.AddProduct) // 👈 ЭТОГО НЕ ХВАТАЛО!

	// ✅ Одобренные товары
	app.Get("/api/products/approved", routes.GetApprovedProducts)

	// 🌍 Публичный маршрут получения юзеров
	app.Get("/api/public/users", routes.GetAllUsersPublic)
	app.Get("/api/public/chat-partners", routes.GetUserChatPartners)

	log.Println("🚀 Сервер запущен на порту 8080")
	if err := app.Listen("0.0.0.0:8080"); err != nil {
		log.Fatalf("❌ Ошибка при запуске сервера: %v", err)
	}
}
