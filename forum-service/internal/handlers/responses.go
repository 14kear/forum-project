package handlers

import "github.com/14kear/forum-project/forum-service/internal/models"

// SuccessIDResponse представляет ID созданного объекта
// swagger:model
type SuccessIDResponse struct {
	// Пример: 123
	ID int64 `json:"id"`
}

// ErrorResponse представляет ответ с ошибкой
// swagger:model
type ErrorResponse struct {
	// Пример: invalid input
	Error string `json:"error"`
}

// ListTopicsResponse представляет список топиков
// swagger:model
type ListTopicsResponse struct {
	Topics []models.Topic `json:"topics"`
}

// ListCommentsResponse представляет список комментариев
// swagger:model
type ListCommentsResponse struct {
	Comments []models.Comment `json:"comments"`
}

// SingleTopicResponse представляет один топик
// swagger:model
type SingleTopicResponse struct {
	Topic models.Topic `json:"topic"`
}

// SingleCommentResponse представляет один комментарий
// swagger:model
type SingleCommentResponse struct {
	Comment models.Comment `json:"comment"`
}
