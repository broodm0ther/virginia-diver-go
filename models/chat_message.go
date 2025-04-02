package models

import "time"

type ChatMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	User      string    `json:"user"`
	Room      string    `json:"room"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
