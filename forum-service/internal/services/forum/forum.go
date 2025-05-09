package forum

import (
	"context"
	"errors"
	"fmt"
	"github.com/14kear/forum-project/forum-service/internal/models"
	"log/slog"
)

var ErrValidation = errors.New("validation error")

type Forum struct {
	log                *slog.Logger
	topicStorage       TopicStorage
	commentStorage     CommentStorage
	chatMessageStorage ChatMessageStorage
}

type TopicStorage interface {
	SaveTopic(ctx context.Context, title, content string, userID int64, email string) (int64, error)
	TopicByID(ctx context.Context, id int) (models.Topic, error)
	Topics(ctx context.Context) ([]models.Topic, error)
	DeleteTopic(ctx context.Context, id int) error
	GetTopicAuthorID(ctx context.Context, id int) (int64, error)
}

type CommentStorage interface {
	SaveComment(ctx context.Context, topicID int, userID int64, content string, email string) (int64, error)
	CommentByID(ctx context.Context, id, topicID int, userID int64) (models.Comment, error)
	CommentsByTopicID(ctx context.Context, topicID int) ([]models.Comment, error)
	DeleteComment(ctx context.Context, id int, topicID int) error
	GetCommentAuthorID(ctx context.Context, id int) (int64, error)
}

type ChatMessageStorage interface {
	SaveChatMessage(ctx context.Context, userID int64, content string, email string) (int64, error)
	ChatMessageByID(ctx context.Context, id int, userID int64) (models.ChatMessage, error)
	ChatMessages(ctx context.Context) ([]models.ChatMessage, error)
	DeleteChatMessage(ctx context.Context, id int) error
}

func NewForum(
	log *slog.Logger,
	topicStorage TopicStorage,
	commentStorage CommentStorage,
	chatMessageStorage ChatMessageStorage,
) *Forum {
	return &Forum{
		log:                log,
		topicStorage:       topicStorage,
		commentStorage:     commentStorage,
		chatMessageStorage: chatMessageStorage,
	}
}

func (f *Forum) CreateTopic(ctx context.Context, title, content string, userID int64, email string) (int64, error) {
	const op = "forum.CreateTopic"

	log := f.log.With(slog.String("op", op))
	log.Info("creating topic")

	if title == "" || content == "" {
		err := errors.New("title or content is empty")
		log.Error("failed to create topic", slog.String("reason", err.Error()))
		return 0, fmt.Errorf("%w: title or content is empty", ErrValidation)
	}

	topicID, err := f.topicStorage.SaveTopic(ctx, title, content, userID, email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("topic created", slog.Int64("topicID", topicID))

	return topicID, nil
}

func (f *Forum) ListTopics(ctx context.Context) ([]models.Topic, error) {
	const op = "forum.ListTopics"

	log := f.log.With(slog.String("op", op))
	log.Info("listing topics")

	topics, err := f.topicStorage.Topics(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("topics listed", slog.Int("topics", len(topics)))

	return topics, nil
}

func (f *Forum) GetTopicByID(ctx context.Context, id int) (models.Topic, error) {
	const op = "forum.GetTopicByID"

	log := f.log.With(slog.String("op", op))
	log.Info("getting topic by ID")

	topic, err := f.topicStorage.TopicByID(ctx, id)
	if err != nil {
		return models.Topic{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("topic found", slog.Int("topicID", topic.ID))

	return topic, nil
}

func (f *Forum) DeleteTopic(ctx context.Context, id int, userID int64) error {
	const op = "forum.DeleteTopic"

	log := f.log.With(slog.String("op", op))
	log.Info("deleting topic")

	authorID, err := f.topicStorage.GetTopicAuthorID(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: unable to get topic author: %w", op, err)
	}

	if authorID != userID {
		return fmt.Errorf("%s: user not authorized to delete this topic", op)
	}

	err = f.topicStorage.DeleteTopic(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("topic deleted", slog.Int("topicID", id))

	return nil
}

func (f *Forum) CreateComment(ctx context.Context, topicID int, userID int64, content string, email string) (int64, error) {
	const op = "forum.CreateComment"

	log := f.log.With(slog.String("op", op))
	log.Info("creating comment")

	if content == "" {
		err := errors.New("content is empty")
		log.Error("failed to create comment", slog.String("reason", err.Error()))
		return 0, fmt.Errorf("%w: content is empty", ErrValidation)
	}

	commentID, err := f.commentStorage.SaveComment(ctx, topicID, userID, content, email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("comment created", slog.Int64("commentID", commentID))

	return commentID, nil
}

func (f *Forum) CommentsByTopicID(ctx context.Context, topicID int) ([]models.Comment, error) {
	const op = "forum.ListComments"

	log := f.log.With(slog.String("op", op), slog.Int("topicID", topicID))
	log.Info("listing comments")

	comments, err := f.commentStorage.CommentsByTopicID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("comments listed", slog.Int("comments", len(comments)))

	return comments, nil
}

func (f *Forum) GetCommentByID(ctx context.Context, id int, topicID int, userID int64) (models.Comment, error) {
	const op = "forum.GetCommentByID"

	log := f.log.With(slog.String("op", op))
	log.Info("getting comment by ID")

	comment, err := f.commentStorage.CommentByID(ctx, id, topicID, userID)
	if err != nil {
		return models.Comment{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("comment found", slog.Int("commentID", comment.ID))

	return comment, nil
}

func (f *Forum) DeleteComment(ctx context.Context, id int, topicID int, userID int64) error {
	const op = "forum.DeleteComment"

	log := f.log.With(slog.String("op", op))
	log.Info("deleting comment")

	authorID, err := f.commentStorage.GetCommentAuthorID(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: unable to get comment author: %w", op, err)
	}

	if authorID != userID {
		return fmt.Errorf("%s: user not authorized to delete this topic", op)
	}

	err = f.commentStorage.DeleteComment(ctx, id, topicID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("comment deleted", slog.Int("commentID", id), slog.Int("topicID", topicID))

	return nil
}

func (f *Forum) CreateChatMessage(ctx context.Context, userID int64, content string, email string) (int64, error) {
	const op = "forum.CreateChatMessage"

	log := f.log.With(slog.String("op", op))
	log.Info("creating chat message")

	if content == "" {
		err := errors.New("content is empty")
		log.Error("failed to create chat message", slog.String("reason", err.Error()))
		return 0, fmt.Errorf("%w: content is empty", ErrValidation)
	}

	chatMessageID, err := f.chatMessageStorage.SaveChatMessage(ctx, userID, content, email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("chat message created", slog.Int64("chatMessageID", chatMessageID))

	return chatMessageID, nil
}

func (f *Forum) ListChatMessages(ctx context.Context) ([]models.ChatMessage, error) {
	const op = "forum.ListChatMessages"

	log := f.log.With(slog.String("op", op))
	log.Info("listing chat messages")

	chatMessages, err := f.chatMessageStorage.ChatMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("chat messages listed", slog.Int("chatMessages", len(chatMessages)))

	return chatMessages, nil
}

func (f *Forum) GetChatMessageByID(ctx context.Context, id int, userID int64) (models.ChatMessage, error) {
	const op = "forum.GetChatMessageByID"

	log := f.log.With(slog.String("op", op))
	log.Info("getting chat message by ID")

	chatMessage, err := f.chatMessageStorage.ChatMessageByID(ctx, id, userID)
	if err != nil {
		return models.ChatMessage{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("chat message found", slog.Int("chatMessageID", chatMessage.ID))

	return chatMessage, nil
}

func (f *Forum) DeleteChatMessage(ctx context.Context, id int) error {
	const op = "forum.DeleteChatMessage"

	log := f.log.With(slog.String("op", op))
	log.Info("deleting chat message")

	err := f.chatMessageStorage.DeleteChatMessage(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("chat message deleted", slog.Int("chatMessageID", id))

	return nil
}
