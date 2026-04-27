package rabbitmq

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPublisher_Publish_GetChannelError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChannelProvider(ctrl)
	publisher := NewPublisher(mockClient, QueueConfig{
		Exchange:   "test-exchange",
		Queue:      "test-queue",
		RoutingKey: "test-key",
	})

	ctx := context.Background()
	msg := NotificationMessage{
		ID:        "test-id",
		Message:   "test message",
		Recipient: "test@example.com",
		Channel:   "email",
		SendAt:    time.Now().Add(time.Hour),
		Attempt:   0,
	}

	expectedErr := errors.New("channel error")
	mockClient.EXPECT().GetChannel().Return(nil, expectedErr)

	err := publisher.Publish(ctx, msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "channel")
}
