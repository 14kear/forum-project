package models

import "time"

type ChatMessage struct {
	ID        int
	UserID    int64
	UserEmail string
	Content   string
	CreatedAt time.Time
}
