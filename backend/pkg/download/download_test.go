package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/manteia/zhulong/pkg/storage"
	"github.com/manteia/zhulong/pkg/storage/mocks"
)

func TestDownloadService_DownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorageInterface(ctrl)
	downloadService := NewDownloadService(mockStorage)
	ctx := context.Background()

	bucketName := "test-bucket"
	objectName := "test-object"
	testData := []byte("test data")

	mockStorage.EXPECT().FileExists(ctx, bucketName, objectName).Return(true, nil)
	mockStorage.EXPECT().GetFileInfo(ctx, bucketName, objectName).Return(&storage.FileInfo{
		Key:         objectName,
		Size:        int64(len(testData)),
		ContentType: "application/octet-stream", // 示例内容类型
		ETag:        "test-etag",
	}, nil)
	mockStorage.EXPECT().DownloadFile(ctx, bucketName, objectName).Return(testData, nil)

	downloadRequest := &DownloadRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}

	result, err := downloadService.DownloadFile(ctx, downloadRequest)

	require.NoError(t, err)
	assert.Equal(t, testData, result.Data)
}