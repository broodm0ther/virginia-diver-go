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

	app.Static("/uploads/", "./uploads")

	authGroup := app.Group("/api/auth")
	authGroup.Post("/login", routes.LoginUser)
	authGroup.Post("/register", routes.RegisterUser)
	authGroup.Get("/profile", routes.AuthMiddleware(), routes.GetProfile)
	authGroup.Post("/update-profile", routes.AuthMiddleware(), routes.UpdateProfile)
	authGroup.Post("/logout", routes.LogoutUser)

	authGroup.Post("/set-role", routes.AuthMiddleware(), routes.SetUserRole)
	authGroup.Get("/all-users", routes.AuthMiddleware(), routes.GetAllUsers)

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
