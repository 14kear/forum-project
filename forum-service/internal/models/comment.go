package models

import "time"

type Comment struct {
	ID        int
	TopicID   int
	UserID    int64
	UserEmail string
	Content   string
	CreatedAt time.Time
}
