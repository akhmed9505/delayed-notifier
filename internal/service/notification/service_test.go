package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()

	notification := domain.Notification{
		Message: "test message",
		Channel: domain.Email,
	}

	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(testID, nil).Do(func(ctx context.Context, n domain.Notification) {
		// Simulate setting timestamps
	})
	mockCache.EXPECT().SetStatus(ctx, testID, domain.Pending).Return(nil)
	mockPub.EXPECT().Publish(ctx, gomock.Any()).Return(nil)

	id, err := service.Create(ctx, notification)

	assert.NoError(t, err)
	assert.Equal(t, testID, id)
}

func TestService_Create_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()

	notification := domain.Notification{
		Message: "test message",
		Channel: domain.Email,
	}
	repoErr := errors.New("repo error")

	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(uuid.Nil, repoErr)

	id, err := service.Create(ctx, notification)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "create notification")
}

func TestService_Create_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()

	notification := domain.Notification{
		Message: "test message",
		Channel: domain.Email,
	}
	publishErr := errors.New("publish error")

	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(testID, nil)
	mockCache.EXPECT().SetStatus(ctx, testID, domain.Pending).Return(nil)
	mockPub.EXPECT().Publish(ctx, gomock.Any()).Return(publishErr)

	id, err := service.Create(ctx, notification)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "publish notification")
}

func TestService_GetStatusByID_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()
	expectedStatus := domain.Pending

	mockCache.EXPECT().GetStatus(ctx, testID).Return(expectedStatus, nil)

	status, err := service.GetStatusByID(ctx, testID)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func TestService_GetStatusByID_CacheMiss_RepositorySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()
	expectedStatus := domain.Sent

	mockCache.EXPECT().GetStatus(ctx, testID).Return(domain.NotificationStatus(""), ErrCacheMiss)
	mockRepo.EXPECT().GetStatusByID(ctx, testID).Return(expectedStatus, nil)
	mockCache.EXPECT().SetStatus(ctx, testID, expectedStatus).Return(nil)

	status, err := service.GetStatusByID(ctx, testID)

	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func TestService_GetStatusByID_CacheError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()
	cacheErr := errors.New("cache error")

	mockCache.EXPECT().GetStatus(ctx, testID).Return(domain.NotificationStatus(""), cacheErr)

	status, err := service.GetStatusByID(ctx, testID)

	assert.Error(t, err)
	assert.Equal(t, domain.NotificationStatus(""), status)
	assert.Contains(t, err.Error(), "cache get status")
}

func TestService_GetStatusByID_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()
	repoErr := errors.New("repo error")

	mockCache.EXPECT().GetStatus(ctx, testID).Return(domain.NotificationStatus(""), ErrCacheMiss)
	mockRepo.EXPECT().GetStatusByID(ctx, testID).Return(domain.NotificationStatus(""), repoErr)

	status, err := service.GetStatusByID(ctx, testID)

	assert.Error(t, err)
	assert.Equal(t, domain.NotificationStatus(""), status)
	assert.Contains(t, err.Error(), "get status")
}

func TestService_UpdateStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()
	newStatus := domain.Sent

	mockRepo.EXPECT().UpdateStatus(ctx, testID, newStatus).Return(nil)
	mockCache.EXPECT().SetStatus(ctx, testID, newStatus).Return(nil)

	err := service.UpdateStatus(ctx, testID, newStatus)

	assert.NoError(t, err)
}

func TestService_UpdateStatus_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockPub := NewMockPublisher(ctrl)
	mockCache := NewMockCache(ctrl)

	service := New(mockRepo, mockPub, mockCache)
	ctx := context.Background()
	testID := uuid.New()
	newStatus := domain.Sent
	repoErr := errors.New("repo error")

	mockRepo.EXPECT().UpdateStatus(ctx, testID, newStatus).Return(repoErr)

	err := service.UpdateStatus(ctx, testID, newStatus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update status")
}
