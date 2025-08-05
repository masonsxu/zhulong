package delete

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/manteia/zhulong/pkg/storage/mocks"
)

func TestDeleteService_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorageInterface(ctrl)
	deleteService := NewDeleteService(mockStorage)
	ctx := context.Background()

	bucketName := "test-bucket"
	objectName := "test-object"

	mockStorage.EXPECT().FileExists(ctx, bucketName, objectName).Return(true, nil)
	mockStorage.EXPECT().DeleteFile(ctx, bucketName, objectName).Return(nil)

	deleteRequest := &DeleteRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}

	result, err := deleteService.DeleteFile(ctx, deleteRequest)

	require.NoError(t, err)
	assert.True(t, result.Success)
}