// 📁 models/product.go
package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Location    string  `json:"location"`
	Category    string  `json:"category"` // "clothing" или "shoes"
	Gender      string  `json:"gender"`   // "male" или "female"
	Size        string  `json:"size"`     // XS-XXXL или 35-46
	Images      string  `json:"images"`   // JSON-строка с путями к фото
	Status      string  `json:"status"`   // pending/approved
	UserID      uint    `json:"user_id"`
}
