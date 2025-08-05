package upload

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/manteia/zhulong/pkg/storage"
	"github.com/manteia/zhulong/pkg/storage/mocks"
)

func TestUploadService_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorageInterface(ctrl)
	uploadService := NewUploadService(mockStorage)
	ctx := context.Background()

	bucketName := "test-bucket"
	fileName := "test-video.mp4"
	contentType := "video/mp4"
	testData := []byte("test data")

	mockStorage.EXPECT().UploadFile(ctx, bucketName, gomock.Any(), testData, contentType).Return(&storage.UploadResult{ETag: "test-etag"}, nil)

	uploadRequest := &UploadRequest{
		FileName:    fileName,
		ContentType: contentType,
		Size:        int64(len(testData)),
		Reader:      bytes.NewReader(testData),
		BucketName:  bucketName,
	}

	result, err := uploadService.UploadFile(ctx, uploadRequest)

	require.NoError(t, err)
	assert.NotEmpty(t, result.FileID)
	assert.Equal(t, "test-etag", result.ETag)
}