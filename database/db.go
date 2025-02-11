package database

import (
	"log"

	"kostya/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "host=localhost user=kostya password=secret dbname=kostya_db port=5432 sslmode=disable TimeZone=UTC"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Ошибка подключения к базе данных:", err)
	}

	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("❌ Ошибка при выполнении миграции:", err)
	}

	log.Println("✅ База данных PostgreSQL подключена и миграции выполнены!")
}
