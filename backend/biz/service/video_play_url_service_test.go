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

func TestGetVideoPlayURL_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetadata := metadatamocks.NewMockMetadataServiceInterface(ctrl)
	mockStorage := storagemocks.NewMockStorageInterface(ctrl)
	service := &VideoService{
		metadataService: mockMetadata,
		storageClient:   mockStorage,
	}
	ctx := context.Background()
	videoID := "test-video-id-001"

	videoMeta := &metadata.FileMetadata{
		FileID:     videoID,
		BucketName: "zhulong-videos",
		ObjectName: "video.mp4",
	}

	mockMetadata.EXPECT().GetMetadata(ctx, videoID).Return(videoMeta, nil)
	mockStorage.EXPECT().GetPresignedURL(ctx, videoMeta.BucketName, videoMeta.ObjectName, gomock.Any()).Return("http://example.com/play.mp4", nil)

	req := &api.VideoPlayURLRequest{VideoID: videoID}
	resp, err := service.GetVideoPlayURL(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "http://example.com/play.mp4", resp.PlayURL)
}