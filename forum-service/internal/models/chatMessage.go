package models

import "time"

type ChatMessage struct {
	ID        int
	UserID    int64
	Content   string
	CreatedAt time.Time
}
