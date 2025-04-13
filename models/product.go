package models

import "gorm.io/gorm"

type Product struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Location    string  `json:"location"`
	Category    string  `json:"category"`
	Gender      string  `json:"gender"`
	Size        string  `json:"size"`
	Images      string  `json:"images"`
	Status      string  `json:"status"`
	UserID      uint    `json:"user_id"`
	User        User    `json:"user" gorm:"foreignKey:UserID"`
	gorm.Model
}
