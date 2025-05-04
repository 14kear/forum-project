package storage

import "errors"

var (
	ErrTopicNotFound       = errors.New("topic not found")
	ErrCommentNotFound     = errors.New("comment not found")
	ErrChatMessageNotFound = errors.New("chat message not found")
)
