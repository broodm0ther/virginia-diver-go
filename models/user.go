package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username   string `gorm:"unique" json:"username"`
	Email      string `gorm:"unique" json:"email"`
	Password   string `json:"password"`
	Avatar     string `json:"avatar,omitempty"`
	Role       string `gorm:"default:'registered'"`
	IsVerified bool   `gorm:"default:false" json:"is_verified"`
	VerifyCode string `json:"verify_code,omitempty"`
	ResetCode  string `json:"-"`
}
