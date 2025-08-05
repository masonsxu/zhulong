package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/metadata"
	metadatamocks "github.com/manteia/zhulong/pkg/metadata/mocks"
	storagemocks "github.com/manteia/zhulong/pkg/storage/mocks"
)

func TestVideoService_DeleteVideo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetadata := metadatamocks.NewMockMetadataServiceInterface(ctrl)
	mockStorage := storagemocks.NewMockStorageInterface(ctrl)
	service := &VideoService{
		metadataService: mockMetadata,
		storageClient:   mockStorage,
	}
	ctx := context.Background()
	videoID := "test-video-id-123"

	videoMeta := &metadata.FileMetadata{
		FileID:     videoID,
		BucketName: "zhulong-videos",
		ObjectName: "video.mp4",
		Thumbnail:  "thumb.jpg",
	}

	mockMetadata.EXPECT().GetMetadata(ctx, videoID).Return(videoMeta, nil)
	mockStorage.EXPECT().DeleteFile(ctx, videoMeta.BucketName, videoMeta.ObjectName).Return(nil)
	mockStorage.EXPECT().DeleteFile(ctx, videoMeta.BucketName, videoMeta.Thumbnail).Return(nil)
	mockMetadata.EXPECT().DeleteMetadata(ctx, videoID).Return(nil)

	req := &api.VideoDeleteRequest{VideoID: videoID}
	resp, err := service.DeleteVideo(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}