package worker

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestWorker_Run_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConsumer := NewMockConsumer(ctrl)
	worker := New(mockConsumer)

	ctx := context.Background()

	mockConsumer.EXPECT().Start(ctx).Return(nil)

	err := worker.Run(ctx)

	assert.NoError(t, err)
}

func TestWorker_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConsumer := NewMockConsumer(ctrl)
	worker := New(mockConsumer)

	ctx := context.Background()
	expectedErr := assert.AnError

	mockConsumer.EXPECT().Start(ctx).Return(expectedErr)

	err := worker.Run(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
