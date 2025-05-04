package models

import "time"

type Comment struct {
	ID        int
	TopicID   int
	UserID    int64
	Content   string
	CreatedAt time.Time
}
