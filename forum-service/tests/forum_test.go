package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/14kear/forum-project/forum-service/tests/suite"
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func getTestUserToken(t *testing.T, st *suite.Suite, ctx context.Context) (string, string) {
	email := gofakeit.Email()
	password := "someStrongPassword123!"

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    1,
	})
	require.NoError(t, err)

	return respLogin.GetAccessToken(), respLogin.GetRefreshToken()
}

func TestCreateTopic_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	title := "Интеграционный тест"
	content := "Контент топика из теста"

	body := map[string]string{
		"title":   title,
		"content": content,
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		st.BaseURL+"/api/forum/topics",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var got struct {
		TopicID int64 `json:"topic_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	assert.True(t, got.TopicID > 0)
}

func TestCreateTopic_EmptyValues(t *testing.T) {
	ctx, st := suite.New(t)

	title := ""
	content := ""

	token, _ := getTestUserToken(t, st, ctx)

	body := map[string]string{
		"title":   title,
		"content": content,
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		st.BaseURL+"/api/forum/topics",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// сервер корректно отреагировал на пустые поля
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateTopic_Unauthorized_NoToken(t *testing.T) {
	ctx, st := suite.New(t)

	body := map[string]string{
		"title":   "Title",
		"content": "Content",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// нет заголовка Authorization

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateTopic_InvalidJSON(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	invalidJSON := `{"title": "Valid Title", "content": 123}` // content должно быть string, тут ошибка типа

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", strings.NewReader(invalidJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListTopics_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	title := "new topic"
	content := "topic content"

	body := map[string]string{
		"title":   title,
		"content": content,
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Теперь вызываем GET /api/forum/topics
	newReq, err := http.NewRequestWithContext(ctx, http.MethodGet, st.BaseURL+"/api/forum/topics", nil)
	require.NoError(t, err)

	newResp, err := st.HTTPClient.Do(newReq)
	require.NoError(t, err)
	defer newResp.Body.Close()

	assert.Equal(t, http.StatusOK, newResp.StatusCode)

	// структура для парсинга ответа
	type Topic struct {
		ID        int    `json:"ID"`
		Title     string `json:"Title"`
		Content   string `json:"Content"`
		UserID    int64  `json:"UserID"`
		UserEmail string `json:"UserEmail"`
		CreatedAt string `json:"CreatedAt"`
	}

	type ListTopicsResponse struct {
		Topics []Topic `json:"topics"`
	}

	var listResp ListTopicsResponse
	err = json.NewDecoder(newResp.Body).Decode(&listResp)
	require.NoError(t, err)

	// проверяем, что список не пуст
	require.NotEmpty(t, listResp.Topics, "Topics list should not be empty")

	found := false
	for _, topic := range listResp.Topics {
		if topic.Title == title && topic.Content == content {
			found = true
			break
		}
	}
	require.True(t, found, "Created topic should be present in the topics list")
}

func TestGetTopicByID_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	// GET по ID созданного топика
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/forum/topics/%d", st.BaseURL, createRespBody.TopicID), nil)
	require.NoError(t, err)

	getResp, err := st.HTTPClient.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var getRespBody struct {
		Topic struct {
			ID      int    `json:"ID"`
			Title   string `json:"Title"`
			Content string `json:"Content"`
			UserID  int64  `json:"UserID"`
		} `json:"topic"`
	}
	err = json.NewDecoder(getResp.Body).Decode(&getRespBody)
	require.NoError(t, err)

	assert.Equal(t, createRespBody.TopicID, getRespBody.Topic.ID)
	assert.Equal(t, "Test Topic", getRespBody.Topic.Title)
	assert.Equal(t, "Test content", getRespBody.Topic.Content)
}

func TestGetTopicByID_InvalidID(t *testing.T) {
	ctx, st := suite.New(t)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, st.BaseURL+"/api/forum/topics/invalid-id", nil)
	require.NoError(t, err)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetTopicByID_NotFound(t *testing.T) {
	ctx, st := suite.New(t)

	nonExistentID := 9999999

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/forum/topics/%d", st.BaseURL, nonExistentID), nil)
	require.NoError(t, err)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDeleteTopic_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/forum/topics/%d", st.BaseURL, createRespBody.TopicID), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)

	delResp, err := st.HTTPClient.Do(delReq)
	require.NoError(t, err)
	defer delResp.Body.Close()

	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)
}

func TestDeleteTopic_FakeID(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/forum/topics/invalid-id", st.BaseURL), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)

	delResp, err := st.HTTPClient.Do(delReq)
	require.NoError(t, err)
	defer delResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, delResp.StatusCode)
}

func TestDeleteTopic_Unauthorized(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	fakeToken := "fakeToken"

	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/forum/topics/%d", st.BaseURL, createRespBody.TopicID), nil)
	delReq.Header.Set("Authorization", "Bearer "+fakeToken)

	delResp, err := st.HTTPClient.Do(delReq)
	require.NoError(t, err)
	defer delResp.Body.Close()
}

func TestDeleteTopic_ServerInternalError(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/forum/topics/%d", st.BaseURL, 100000000), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)

	delResp, err := st.HTTPClient.Do(delReq)
	require.NoError(t, err)
	defer delResp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, delResp.StatusCode)
}

func TestCreateComment_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)
	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	createBodyForComment := map[string]string{
		"content": "Test content",
	}
	createBodyForCommentBytes, err := json.Marshal(createBodyForComment)
	require.NoError(t, err)

	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, createRespBody.TopicID)

	createCommentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		commentURL,
		bytes.NewBuffer(createBodyForCommentBytes),
	)
	require.NoError(t, err)

	createCommentReq.Header.Set("Content-Type", "application/json")
	createCommentReq.Header.Set("Authorization", "Bearer "+token)

	createCommentResp, err := st.HTTPClient.Do(createCommentReq)
	require.NoError(t, err)
	defer createCommentResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createCommentResp.StatusCode)

	var got struct {
		CommentID int64 `json:"comment_id"`
	}
	err = json.NewDecoder(createCommentResp.Body).Decode(&got)
	require.NoError(t, err)

	assert.True(t, got.CommentID > 0)
}

func TestCreateComment_ServerInternalError(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	createBodyForComment := map[string]string{
		"content": "Test content",
	}
	createBodyForCommentBytes, err := json.Marshal(createBodyForComment)
	require.NoError(t, err)

	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, 1000000)

	createCommentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		commentURL,
		bytes.NewBuffer(createBodyForCommentBytes),
	)
	require.NoError(t, err)

	createCommentReq.Header.Set("Content-Type", "application/json")
	createCommentReq.Header.Set("Authorization", "Bearer "+token)

	createCommentResp, err := st.HTTPClient.Do(createCommentReq)
	require.NoError(t, err)
	defer createCommentResp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, createCommentResp.StatusCode)
}

func TestCreateComment_EmptyValues(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)
	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	createBodyForComment := map[string]string{
		"content": "",
	}
	createBodyForCommentBytes, err := json.Marshal(createBodyForComment)
	require.NoError(t, err)

	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, createRespBody.TopicID)

	createCommentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		commentURL,
		bytes.NewBuffer(createBodyForCommentBytes),
	)
	require.NoError(t, err)

	createCommentReq.Header.Set("Content-Type", "application/json")
	createCommentReq.Header.Set("Authorization", "Bearer "+token)

	createCommentResp, err := st.HTTPClient.Do(createCommentReq)
	require.NoError(t, err)
	defer createCommentResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, createCommentResp.StatusCode)
}

func TestCreateComment_Unauthorized_NoToken(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)
	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	createBodyForComment := map[string]string{
		"content": "Test content",
	}
	createBodyForCommentBytes, err := json.Marshal(createBodyForComment)
	require.NoError(t, err)

	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, createRespBody.TopicID)

	createCommentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		commentURL,
		bytes.NewBuffer(createBodyForCommentBytes),
	)
	require.NoError(t, err)

	createCommentReq.Header.Set("Content-Type", "application/json")

	createCommentResp, err := st.HTTPClient.Do(createCommentReq)
	require.NoError(t, err)
	defer createCommentResp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, createCommentResp.StatusCode)
}

func TestCreateComment_InvalidJSON(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	createBody := map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	}
	createBodyBytes, err := json.Marshal(createBody)
	require.NoError(t, err)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, st.BaseURL+"/api/forum/topics", bytes.NewBuffer(createBodyBytes))
	require.NoError(t, err)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := st.HTTPClient.Do(createReq)
	require.NoError(t, err)
	defer createResp.Body.Close()

	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createRespBody struct {
		TopicID int `json:"topic_id"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createRespBody)
	require.NoError(t, err)

	invalidJSON := `{"content": 123}` // content должно быть string, тут ошибка типа

	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, createRespBody.TopicID)

	createCommentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		commentURL,
		strings.NewReader(invalidJSON),
	)
	require.NoError(t, err)

	createCommentReq.Header.Set("Content-Type", "application/json")
	createCommentReq.Header.Set("Authorization", "Bearer "+token)

	createCommentResp, err := st.HTTPClient.Do(createCommentReq)
	require.NoError(t, err)
	defer createCommentResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, createCommentResp.StatusCode)
}

func TestListCommentsByTopic_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	// получаем список комментариев
	getReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, commentURL, nil)
	getResp, err := st.HTTPClient.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	// структура ответа
	type Comment struct {
		ID        int    `json:"ID"`
		TopicID   int    `json:"TopicID"`
		Content   string `json:"Content"`
		UserID    int64  `json:"UserID"`
		UserEmail string `json:"UserEmail"`
		CreatedAt string `json:"CreatedAt"`
	}
	var listResp struct {
		Comments []Comment `json:"comments"`
	}
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&listResp))

	require.NotEmpty(t, listResp.Comments, "comments list should not be empty")

	found := false
	for _, cmt := range listResp.Comments {
		if cmt.TopicID == topicRespBody.TopicID && cmt.Content == "new comment" {
			found = true
			break
		}
	}
	require.True(t, found, "created comment must be in the list")
}

func TestListCommentsByTopic_EmptyValues(t *testing.T) {
	ctx, st := suite.New(t)

	invalidURL := fmt.Sprintf("%s/api/forum/topics/invalid-id/comments", st.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, invalidURL, nil)
	require.NoError(t, err)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetCommentByID_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	var CommentRespBody struct {
		CommentID int `json:"comment_id"`
	}
	require.NoError(t, json.NewDecoder(commentResp.Body).Decode(&CommentRespBody))

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/%d/comments/%d", st.BaseURL, topicRespBody.TopicID, CommentRespBody.CommentID)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, commentByIdURL, nil)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusOK, commentByIdResp.StatusCode)
}

func TestGetCommentByID_InvalidCommentID(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/topic-id/comments/comment-id", st.BaseURL)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, commentByIdURL, nil)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusBadRequest, commentByIdResp.StatusCode)
}

func TestListCommentByID_ServerInternalError(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/%d/comments/%d", st.BaseURL, topicRespBody.TopicID, 100000000)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, commentByIdURL, nil)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, commentByIdResp.StatusCode)
}

func TestListCommentsByTopic_ServerInternalError(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	commentListingURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, 100000000000)

	getReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, commentListingURL, nil)
	getResp, err := st.HTTPClient.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, getResp.StatusCode)
}

func TestGetCommentByID_InvalidTopicID(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	var CommentRespBody struct {
		CommentID int `json:"comment_id"`
	}
	require.NoError(t, json.NewDecoder(commentResp.Body).Decode(&CommentRespBody))

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/topic-id/comments/%d", st.BaseURL, CommentRespBody.CommentID)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, commentByIdURL, nil)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusBadRequest, commentByIdResp.StatusCode)
}

func TestDeleteComment_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	var CommentRespBody struct {
		CommentID int `json:"comment_id"`
	}
	require.NoError(t, json.NewDecoder(commentResp.Body).Decode(&CommentRespBody))

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/%d/comments/%d", st.BaseURL, topicRespBody.TopicID, CommentRespBody.CommentID)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, commentByIdURL, nil)

	commentByIdReq.Header.Set("Content-Type", "application/json")
	commentByIdReq.Header.Set("Authorization", "Bearer "+token)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusNoContent, commentByIdResp.StatusCode)
}

func TestDeleteComment_InvalidCommentID(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/%d/comments/comment-id", st.BaseURL, topicRespBody.TopicID)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, commentByIdURL, nil)

	commentByIdReq.Header.Set("Content-Type", "application/json")
	commentByIdReq.Header.Set("Authorization", "Bearer "+token)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusBadRequest, commentByIdResp.StatusCode)
}

func TestDeleteComment_InvalidTopicID(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	var CommentRespBody struct {
		CommentID int `json:"comment_id"`
	}
	require.NoError(t, json.NewDecoder(commentResp.Body).Decode(&CommentRespBody))

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/topic-id/comments/%d", st.BaseURL, CommentRespBody.CommentID)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, commentByIdURL, nil)

	commentByIdReq.Header.Set("Content-Type", "application/json")
	commentByIdReq.Header.Set("Authorization", "Bearer "+token)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusBadRequest, commentByIdResp.StatusCode)
}

func TestDeleteComment_Unauthorized(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	var CommentRespBody struct {
		CommentID int `json:"comment_id"`
	}
	require.NoError(t, json.NewDecoder(commentResp.Body).Decode(&CommentRespBody))

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/%d/comments/%d", st.BaseURL, topicRespBody.TopicID, CommentRespBody.CommentID)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, commentByIdURL, nil)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, commentByIdResp.StatusCode)
}

func TestDeleteComment_ServerInternalError(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	topicBody, _ := json.Marshal(map[string]string{
		"title":   "Test Topic",
		"content": "Test content",
	})
	topicReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		st.BaseURL+"/api/forum/topics", bytes.NewBuffer(topicBody))
	topicReq.Header.Set("Content-Type", "application/json")
	topicReq.Header.Set("Authorization", "Bearer "+token)

	topicResp, err := st.HTTPClient.Do(topicReq)
	require.NoError(t, err)
	defer topicResp.Body.Close()
	require.Equal(t, http.StatusCreated, topicResp.StatusCode)

	var topicRespBody struct {
		TopicID int `json:"topic_id"`
	}
	require.NoError(t, json.NewDecoder(topicResp.Body).Decode(&topicRespBody))

	// создаём комментарий
	commentBody, _ := json.Marshal(map[string]string{
		"content": "new comment",
	})
	commentURL := fmt.Sprintf("%s/api/forum/topics/%d/comments", st.BaseURL, topicRespBody.TopicID)

	commentReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		commentURL, bytes.NewBuffer(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)

	commentResp, err := st.HTTPClient.Do(commentReq)
	require.NoError(t, err)
	defer commentResp.Body.Close()
	require.Equal(t, http.StatusCreated, commentResp.StatusCode)

	commentByIdURL := fmt.Sprintf("%s/api/forum/topics/%d/comments/%d", st.BaseURL, topicRespBody.TopicID, 10000000000)
	commentByIdReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, commentByIdURL, nil)

	commentByIdReq.Header.Set("Content-Type", "application/json")
	commentByIdReq.Header.Set("Authorization", "Bearer "+token)

	commentByIdResp, err := st.HTTPClient.Do(commentByIdReq)
	require.NoError(t, err)
	defer commentByIdResp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, commentByIdResp.StatusCode)
}

func TestWebSocketChat_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	// Формируем ws URL с accessToken
	wsURL := fmt.Sprintf("ws%s/api/forum/ws/chat?accessToken=%s", strings.TrimPrefix(st.BaseURL, "http"), token)

	// Устанавливаем WebSocket-соединение
	wsConn, resp, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	defer wsConn.Close()

	// Отправляем сообщение в чат
	err = wsConn.WriteJSON(map[string]string{
		"content": "hello from test",
	})
	require.NoError(t, err)

	// Читаем ответ
	var response struct {
		ID        int64  `json:"ID"`
		Content   string `json:"Content"`
		UserID    int64  `json:"UserID"`
		UserEmail string `json:"UserEmail"`
	}
	err = wsConn.ReadJSON(&response)
	require.NoError(t, err)

	// Проверки
	require.NotZero(t, response.ID)
	assert.Equal(t, "hello from test", response.Content)
	assert.NotZero(t, response.UserID)
	assert.Contains(t, response.UserEmail, "@")
}

// запрос без accessToken вообще: должны получить HTTP-401 и никакого апгрейда.
func TestWebSocketChat_MissingToken_Unauthorized(t *testing.T) {
	ctx, st := suite.New(t)

	url := st.BaseURL + "/api/forum/ws/chat"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "missing access token", body.Error)
}

// неверный (невалидный) accessToken: соединение апгрейдится, но сервер сразу присылает {"error":"unauthorized"} и закрывает сокет.
func TestWebSocketChat_InvalidToken_ShouldCloseWithUnauthorized(t *testing.T) {
	ctx, st := suite.New(t)

	badToken := "invalidToken"

	scheme := "ws"
	if strings.HasPrefix(st.BaseURL, "https") {
		scheme = "wss"
	}
	wsURL := fmt.Sprintf("%s%s/api/forum/ws/chat?accessToken=%s",
		scheme, strings.TrimPrefix(st.BaseURL, "http"), badToken)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// читаем первый и единственный фрейм
	var errResp map[string]string
	require.NoError(t, conn.ReadJSON(&errResp))
	assert.Equal(t, "unauthorized", errResp["error"])

	// после этого сервер закрывает соединение; попытка читать ещё должна дать ошибку
	_, _, err = conn.ReadMessage()
	assert.Error(t, err)
}

// токен валидный, но клиент шлёт НЕ JSON. Сервер разрывает соединение.
func TestWebSocketChat_BadJSON_ShouldClose(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	scheme := "ws"
	if strings.HasPrefix(st.BaseURL, "https") {
		scheme = "wss"
	}
	wsURL := fmt.Sprintf("%s%s/api/forum/ws/chat?accessToken=%s",
		scheme, strings.TrimPrefix(st.BaseURL, "http"), token)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// отправляем не-JSON
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("not json")))

	// сервер должен закрыть соединение; чтение вернёт ошибку
	_, _, err = conn.ReadMessage()
	assert.Error(t, err)
}

func TestGetChatMessages_Success(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	wsURL := fmt.Sprintf("ws%s/api/forum/ws/chat?accessToken=%s", strings.TrimPrefix(st.BaseURL, "http"), token)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// шлём сообщение
	err = conn.WriteJSON(map[string]string{
		"content": "first message",
	})
	require.NoError(t, err)

	var msg struct {
		ID        int64  `json:"ID"`
		Content   string `json:"Content"`
		UserID    int64  `json:"UserID"`
		UserEmail string `json:"UserEmail"`
	}
	require.NoError(t, conn.ReadJSON(&msg))

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, st.BaseURL+"/api/forum/ws/chat/messages", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var allMsgs []struct {
		ID        int64  `json:"ID"`
		Content   string `json:"Content"`
		UserID    int64  `json:"UserID"`
		UserEmail string `json:"UserEmail"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&allMsgs))

	require.NotEmpty(t, allMsgs)

	found := false
	for _, m := range allMsgs {
		if m.ID == msg.ID && m.Content == "first message" {
			found = true
			break
		}
	}
	assert.True(t, found, "sent message not found in chat history")
}

func TestCleanupOldMessages_RemovesOldChatMessages(t *testing.T) {
	ctx, st := suite.New(t)

	token, _ := getTestUserToken(t, st, ctx)

	wsURL := fmt.Sprintf("ws%s/api/forum/ws/chat?accessToken=%s",
		strings.TrimPrefix(st.BaseURL, "http"), token)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	err = conn.WriteJSON(map[string]string{
		"content": "message to be deleted",
	})
	require.NoError(t, err)

	var resp struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, conn.ReadJSON(&resp))
	require.NotZero(t, resp.ID)

	err = st.ForumService.CleanupOldMessages(ctx, 0) // 0 = всё старше "сейчас"
	require.NoError(t, err)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, st.BaseURL+"/api/forum/ws/chat/messages", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp2, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var messages []struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&messages))

	for _, m := range messages {
		assert.NotEqual(t, resp.ID, m.ID, "удалённое сообщение всё ещё в списке")
	}
}

func TestMiddleware_TokenRefreshFlow(t *testing.T) {
	ctx, st := suite.New(t)

	// Получаем нормальный токен и refresh
	_, refreshToken := getTestUserToken(t, st, ctx)

	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBfaWQiOjEsImVtYWlsIjoibG9oQG1haWwucnUiLCJleHAiOjE3NDY1OTk2ODEsInR5cCI6ImFjY2VzcyIsInVpZCI6ODN9.BPOAmK9u3y6rxXhGcBlam0wPexBqhoEU8V7TTLeevAQ"

	title := "Интеграционный тест"
	content := "Контент топика из теста"

	body := map[string]string{
		"title":   title,
		"content": content,
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		st.BaseURL+"/api/forum/topics",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Refresh-Token", refreshToken)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var got struct {
		TopicID int64 `json:"topic_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	assert.True(t, got.TopicID > 0)

	// Проверяем что обновлённые токены вернулись
	newAccess := resp.Header.Get("X-New-Access-Token")
	newRefresh := resp.Header.Get("X-New-Refresh-Token")
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)

	// Опционально — убедиться, что новые токены работают
	req2, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		st.BaseURL+"/api/forum/topics",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)
	req2.Header.Set("Authorization", "Bearer "+newAccess)
	resp2, err := st.HTTPClient.Do(req2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
}

func TestMiddleware_TokenRefreshFail(t *testing.T) {
	ctx, st := suite.New(t)

	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBfaWQiOjEsImVtYWlsIjoibG9oQG1haWwucnUiLCJleHAiOjE3NDY1OTk2ODEsInR5cCI6ImFjY2VzcyIsInVpZCI6ODN9.BPOAmK9u3y6rxXhGcBlam0wPexBqhoEU8V7TTLeevAQ"
	refreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImxvaEBtYWlsLnJ1IiwiZXhwIjoxNzQ3MjAzNTgxLCJ0eXAiOiJyZWZyZXNoIiwidWlkIjo4M30.IqMl2MvEKfPWqx2M8KaXlXFclomG1STDkAOSPP01lgA"

	title := "Интеграционный тест"
	content := "Контент топика из теста"

	body := map[string]string{
		"title":   title,
		"content": content,
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		st.BaseURL+"/api/forum/topics",
		bytes.NewBuffer(bodyBytes),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Refresh-Token", refreshToken)

	resp, err := st.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем что обновлённые токены не вернулись
	newAccess := resp.Header.Get("X-New-Access-Token")
	newRefresh := resp.Header.Get("X-New-Refresh-Token")
	assert.Empty(t, newAccess)
	assert.Empty(t, newRefresh)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
