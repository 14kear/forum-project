package forum

import (
	"context"
	"errors"
	"github.com/14kear/forum-project/forum-service/internal/config"
	"github.com/14kear/forum-project/forum-service/internal/models"
	"github.com/14kear/forum-project/forum-service/internal/services/mocks"
	"github.com/14kear/forum-project/forum-service/utils"
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var configPath = "C:\\Users\\shini\\OneDrive\\Рабочий стол\\forum-project\\forum-service\\config\\local.yaml"

func newTestForum(
	ctrl *gomock.Controller,
	topicStorage *mocks.MockTopicStorage,
	commentStorage *mocks.MockCommentStorage,
	chatMessagesStorage *mocks.MockChatMessageStorage,
	authClient ssov1.AuthClient) *Forum {
	return NewForum(utils.New(config.Load(configPath).Env), topicStorage, commentStorage, chatMessagesStorage, authClient)
}

func TestForum_CreateTopic_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().SaveTopic(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(155), nil)

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	topicID, err := testForum.CreateTopic(context.Background(), "new topic", "about tests", 66, "test@test.com")
	require.NoError(t, err)
	require.Equal(t, int64(155), topicID)
}

func TestForum_CreateTopic_EmptyTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testForum := newTestForum(ctrl, nil, nil, nil, nil)

	_, err := testForum.CreateTopic(context.Background(), "", "", 66, "test@test.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValidation.Error())
}

func TestForum_CreateTopic_FailSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().SaveTopic(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.New("save failed"))

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	id, err := testForum.CreateTopic(context.Background(), "a", "b", 66, "test@test.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save failed")
	assert.Equal(t, int64(0), id)
}

func TestForum_DeleteTopic_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().GetTopicAuthorID(gomock.Any(), gomock.Any()).Return(int64(1), nil)
	topicStorage.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).Return(nil)

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	err := testForum.DeleteTopic(context.Background(), 111, 1)
	require.NoError(t, err)
}

func TestForum_DeleteTopic_FailGetAuthorID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().GetTopicAuthorID(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("GetAuthorID failed"))

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	err := testForum.DeleteTopic(context.Background(), 111, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GetAuthorID failed")
}

func TestForum_DeleteTopic_FailDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)
	authClient := mocks.NewMockAuthClient(ctrl)

	topicID := 10
	userID := int64(123)
	authorID := int64(999)

	topicStorage.EXPECT().GetTopicAuthorID(gomock.Any(), topicID).Return(authorID, nil)

	authClient.EXPECT().
		IsAdmin(gomock.Any(), &ssov1.IsAdminRequest{UserId: userID}).
		Return(nil, errors.New("grpc error"))

	testForum := newTestForum(ctrl, topicStorage, nil, nil, authClient)

	err := testForum.DeleteTopic(context.Background(), topicID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check admin rights")
}

func TestForum_DeleteTopic_FailIsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)
	authClient := mocks.NewMockAuthClient(ctrl)

	topicID := 10
	userID := int64(123)
	authorID := int64(999)

	topicStorage.EXPECT().GetTopicAuthorID(gomock.Any(), topicID).Return(authorID, nil)

	authClient.EXPECT().
		IsAdmin(gomock.Any(), &ssov1.IsAdminRequest{UserId: userID}).
		Return(&ssov1.IsAdminResponse{IsAdmin: false}, nil)

	testForum := newTestForum(ctrl, topicStorage, nil, nil, authClient)

	err := testForum.DeleteTopic(context.Background(), topicID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not authorized to delete this topic")
}

func TestForum_DeleteTopic_FailDeleteTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().GetTopicAuthorID(gomock.Any(), gomock.Any()).Return(int64(1), nil)
	topicStorage.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).Return(errors.New("DeleteTopic failed"))

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	err := testForum.DeleteTopic(context.Background(), 111, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DeleteTopic failed")
}

func TestForum_ListTopics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().Topics(gomock.Any()).Return([]models.Topic{}, nil)

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	topics, err := testForum.ListTopics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []models.Topic{}, topics)
}

func TestForum_ListTopics_FailList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().Topics(gomock.Any()).Return(nil, errors.New("List failed"))

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	_, err := testForum.ListTopics(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "List failed")
}

func TestForum_GetTopicByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)
	topic := models.Topic{ID: 1, Title: "test", Content: "test", UserID: 111, UserEmail: "test@test.com", CreatedAt: time.Unix(500, 0)}

	topicStorage.EXPECT().TopicByID(gomock.Any(), 111).Return(topic, nil)

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	topicFinal, err := testForum.GetTopicByID(context.Background(), int(topic.UserID))
	require.NoError(t, err)
	assert.Equal(t, topic, topicFinal)
}

func TestForum_GetTopicByID_FailGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topicStorage := mocks.NewMockTopicStorage(ctrl)

	topicStorage.EXPECT().TopicByID(gomock.Any(), gomock.Any()).Return(models.Topic{}, errors.New("Get failed"))

	testForum := newTestForum(ctrl, topicStorage, nil, nil, nil)

	topic, err := testForum.GetTopicByID(context.Background(), 111)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Get failed")
	assert.Equal(t, models.Topic{}, topic)
}

func TestForum_CreateComment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().SaveComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(55), nil)

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	commentID, err := testForum.CreateComment(context.Background(), 1, 11, "new comment", "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, int64(55), commentID)
}

func TestForum_CreateComment_EmptyComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testForum := newTestForum(ctrl, nil, nil, nil, nil)

	_, err := testForum.CreateComment(context.Background(), 1, 11, "", "test@test.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValidation.Error())
}

func TestForum_CreateComment_FailCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().SaveComment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.New("CreateComment failed"))

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	_, err := testForum.CreateComment(context.Background(), 1, 11, "new comment", "test@test.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreateComment failed")
}

func TestForum_CommentByTopicID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().CommentsByTopicID(gomock.Any(), gomock.Any()).Return([]models.Comment{}, nil)

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	comments, err := testForum.CommentsByTopicID(context.Background(), 50)
	require.NoError(t, err)
	assert.Equal(t, []models.Comment{}, comments)
}

func TestForum_CommentByTopicID_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().CommentsByTopicID(gomock.Any(), gomock.Any()).Return(nil, errors.New("CommentsByTopicID failed"))

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	_, err := testForum.CommentsByTopicID(context.Background(), 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CommentsByTopicID failed")
}

func TestForum_GetCommentByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	comment := models.Comment{ID: 1, TopicID: 5, UserID: 1}

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().CommentByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(comment, nil)

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	commentFinal, err := testForum.GetCommentByID(context.Background(), comment.ID, comment.TopicID)
	require.NoError(t, err)
	assert.Equal(t, comment, commentFinal)
}

func TestForum_GetCommentByID_FailGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	comment := models.Comment{ID: 1, TopicID: 5}

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().CommentByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(models.Comment{}, errors.New("Get failed"))

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	commentFinal, err := testForum.GetCommentByID(context.Background(), comment.ID, comment.TopicID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Get failed")
	assert.Equal(t, models.Comment{}, commentFinal)
}

func TestForum_DeleteComment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().GetCommentAuthorID(gomock.Any(), gomock.Any()).Return(int64(1), nil)
	commentStorage.EXPECT().DeleteComment(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	err := testForum.DeleteComment(context.Background(), 111, 1, 1)
	require.NoError(t, err)
}

func TestForum_DeleteComment_FailGetAuthorID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().GetCommentAuthorID(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("GetAuthorID failed"))

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	err := testForum.DeleteComment(context.Background(), 111, 1, 55)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GetAuthorID failed")
}

func TestForum_DeleteComment_FailDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)
	authClient := mocks.NewMockAuthClient(ctrl)

	commentID := 10
	userID := int64(123)
	authorID := int64(999)
	topicID := 10

	commentStorage.EXPECT().GetCommentAuthorID(gomock.Any(), commentID).Return(authorID, nil)

	authClient.EXPECT().
		IsAdmin(gomock.Any(), &ssov1.IsAdminRequest{UserId: userID}).
		Return(nil, errors.New("grpc error"))

	testForum := newTestForum(ctrl, nil, commentStorage, nil, authClient)

	err := testForum.DeleteComment(context.Background(), commentID, topicID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check admin rights")
}

func TestForum_DeleteComment_FailIsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)
	authClient := mocks.NewMockAuthClient(ctrl)

	commentID := 10
	topicID := 10
	userID := int64(123)
	authorID := int64(999)

	commentStorage.EXPECT().GetCommentAuthorID(gomock.Any(), commentID).Return(authorID, nil)

	authClient.EXPECT().
		IsAdmin(gomock.Any(), &ssov1.IsAdminRequest{UserId: userID}).
		Return(&ssov1.IsAdminResponse{IsAdmin: false}, nil)

	testForum := newTestForum(ctrl, nil, commentStorage, nil, authClient)

	err := testForum.DeleteComment(context.Background(), commentID, topicID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not authorized to delete this topic")
}

func TestForum_DeleteComment_FailDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentStorage := mocks.NewMockCommentStorage(ctrl)

	commentStorage.EXPECT().GetCommentAuthorID(gomock.Any(), gomock.Any()).Return(int64(1), nil)
	commentStorage.EXPECT().DeleteComment(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("DeleteComment failed"))

	testForum := newTestForum(ctrl, nil, commentStorage, nil, nil)

	err := testForum.DeleteComment(context.Background(), 111, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DeleteComment failed")
}

func TestForum_CreateChatMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatMessageStorage := mocks.NewMockChatMessageStorage(ctrl)

	chatMessageStorage.EXPECT().SaveChatMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(15), nil)

	testForum := newTestForum(ctrl, nil, nil, chatMessageStorage, nil)

	chatMessageID, err := testForum.CreateChatMessage(context.Background(), 55, "hi", "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, int64(15), chatMessageID)
}

func TestForum_CreateChatMessage_EmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testForum := newTestForum(ctrl, nil, nil, nil, nil)

	_, err := testForum.CreateChatMessage(context.Background(), 15, "", "test@test.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValidation.Error())
}

func TestForum_CreateChatMessage_FailSaveChatMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatMessageStorage := mocks.NewMockChatMessageStorage(ctrl)

	chatMessageStorage.EXPECT().SaveChatMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.New("SaveChatMessage failed"))

	testForum := newTestForum(ctrl, nil, nil, chatMessageStorage, nil)
	_, err := testForum.CreateChatMessage(context.Background(), 55, "hi", "test@test.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SaveChatMessage failed")
}

func TestForum_ListChatMessages_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatMessageStorage := mocks.NewMockChatMessageStorage(ctrl)

	chatMessageStorage.EXPECT().ChatMessages(gomock.Any()).Return([]models.ChatMessage{}, nil)

	testForum := newTestForum(ctrl, nil, nil, chatMessageStorage, nil)

	chatMessages, err := testForum.ListChatMessages(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []models.ChatMessage{}, chatMessages)
}

func TestForum_ListChatMessages_FailChatMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatMessageStorage := mocks.NewMockChatMessageStorage(ctrl)

	chatMessageStorage.EXPECT().ChatMessages(gomock.Any()).Return(nil, errors.New("ChatMessages failed"))

	testForum := newTestForum(ctrl, nil, nil, chatMessageStorage, nil)

	_, err := testForum.ListChatMessages(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ChatMessages failed")
}

func TestForum_CleanupOldMessages_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatMessageStorage := mocks.NewMockChatMessageStorage(ctrl)

	chatMessageStorage.EXPECT().DeleteChatMessagesBefore(gomock.Any(), gomock.Any()).Return(nil)

	testForum := newTestForum(ctrl, nil, nil, chatMessageStorage, nil)

	err := testForum.CleanupOldMessages(context.Background(), 24*time.Hour)
	require.NoError(t, err)
}

func TestForum_CleanupOldMessages_FailDeleteOldMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatMessageStorage := mocks.NewMockChatMessageStorage(ctrl)

	chatMessageStorage.EXPECT().DeleteChatMessagesBefore(gomock.Any(), gomock.Any()).Return(errors.New("DeleteChatMessages failed"))

	testForum := newTestForum(ctrl, nil, nil, chatMessageStorage, nil)

	err := testForum.CleanupOldMessages(context.Background(), 24*time.Hour)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DeleteChatMessages failed")
}
