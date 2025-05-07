package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/14kear/forum-project/forum-service/internal/models"
	"github.com/14kear/forum-project/forum-service/internal/storage"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(postgresURL string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveTopic(ctx context.Context, title, content string, userID int64) (int64, error) {
	const op = "storage.postgres.NewTopic"

	stmt, err := s.db.Prepare("INSERT INTO topics(title, content, user_id) VALUES ($1, $2, $3) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRowContext(ctx, title, content, userID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) TopicByID(ctx context.Context, id int) (models.Topic, error) {
	const op = "storage.postgres.Topic"

	stmt, err := s.db.Prepare("SELECT id, title, content, user_id, created_at FROM topics WHERE id = $1")
	if err != nil {
		return models.Topic{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var topic models.Topic
	err = stmt.QueryRowContext(ctx, id).Scan(&topic.ID, &topic.Title, &topic.Content, &topic.UserID, &topic.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Topic{}, fmt.Errorf("%s: %w", op, storage.ErrTopicNotFound)
		}
		return models.Topic{}, fmt.Errorf("%s: %w", op, err)
	}

	return topic, nil
}

func (s *Storage) Topics(ctx context.Context) ([]models.Topic, error) {
	const op = "storage.postgres.GetAllTopics"

	rows, err := s.db.QueryContext(ctx, `
        SELECT id, title, content, user_id, created_at
        FROM topics
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var topic models.Topic
		if err := rows.Scan(&topic.ID, &topic.Title, &topic.Content, &topic.UserID, &topic.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return topics, nil
}

func (s *Storage) DeleteTopic(ctx context.Context, id int) error {
	const op = "storage.postgres.DeleteTopic"

	stmt, err := s.db.Prepare("DELETE FROM topics WHERE id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrTopicNotFound)
	}

	return nil
}

func (s *Storage) SaveComment(ctx context.Context, topicID int, userID int64, content string) (int64, error) {
	const op = "storage.postgres.SaveComment"

	stmt, err := s.db.Prepare("INSERT INTO comments(topic_id, user_id, content) VALUES ($1, $2, $3) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRowContext(ctx, topicID, userID, content).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) CommentByID(ctx context.Context, id, topicID int, userID int64) (models.Comment, error) {
	const op = "storage.postgres.Comment"

	stmt, err := s.db.Prepare("SELECT id, topic_id, user_id, content, created_at FROM comments WHERE topic_id = $1 AND id = $2 AND user_id = $3")
	if err != nil {
		return models.Comment{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var comment models.Comment
	err = stmt.QueryRowContext(ctx, topicID, id, userID).Scan(
		&comment.ID,
		&comment.TopicID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Comment{}, fmt.Errorf("%s: %w", op, storage.ErrCommentNotFound)
		}
		return models.Comment{}, fmt.Errorf("%s: %w", op, err)
	}

	return comment, nil
}

func (s *Storage) CommentsByTopicID(ctx context.Context, topicID int) ([]models.Comment, error) {
	const op = "storage.postgres.CommentsByTopicID"

	rows, err := s.db.QueryContext(ctx, `
        SELECT id, topic_id, user_id, content, created_at
        FROM comments 
        WHERE topic_id = $1
        ORDER BY created_at DESC
    `, topicID)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.TopicID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return comments, nil
}

func (s *Storage) DeleteComment(ctx context.Context, id int, topicID int) error {
	const op = "storage.postgres.DeleteComment"

	res, err := s.db.ExecContext(ctx, "DELETE FROM comments WHERE id = $1 AND topic_id = $2", id, topicID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrCommentNotFound)
	}

	return nil
}

func (s *Storage) SaveChatMessage(ctx context.Context, userID int64, content string) (int64, error) {
	const op = "storage.postgres.SaveChatMessage"

	stmt, err := s.db.Prepare("INSERT INTO chat_messages(user_id, content) VALUES ($1, $2) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRowContext(ctx, userID, content).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) ChatMessageByID(ctx context.Context, id int, userID int64) (models.ChatMessage, error) {
	const op = "storage.postgres.ChatMessage"

	stmt, err := s.db.Prepare("SELECT id, user_id, content, created_at FROM chat_messages WHERE id = $1")
	if err != nil {
		return models.ChatMessage{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var msg models.ChatMessage
	err = stmt.QueryRowContext(ctx, id, userID).Scan(
		&msg.ID,
		&msg.UserID,
		&msg.Content,
		&msg.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ChatMessage{}, fmt.Errorf("%s: %w", op, storage.ErrChatMessageNotFound)
		}
		return models.ChatMessage{}, fmt.Errorf("%s: %w", op, err)
	}

	return msg, nil
}

func (s *Storage) ChatMessages(ctx context.Context) ([]models.ChatMessage, error) {
	const op = "storage.postgres.ChatMessages"

	rows, err := s.db.QueryContext(ctx, `
        SELECT id, user_id, content, created_at
        FROM chat_messages
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return messages, nil
}

func (s *Storage) DeleteChatMessage(ctx context.Context, id int) error {
	const op = "storage.postgres.DeleteChatMessage"

	stmt, err := s.db.Prepare("DELETE FROM chat_messages WHERE id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrCommentNotFound)
	}

	return nil
}
