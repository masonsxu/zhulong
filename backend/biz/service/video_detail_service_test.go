package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/metadata"
	metadatamocks "github.com/manteia/zhulong/pkg/metadata/mocks"
)

func TestGetVideoDetail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetadataService := metadatamocks.NewMockMetadataServiceInterface(ctrl)
	service := &VideoService{metadataService: mockMetadataService}
	ctx := context.Background()

	videoID := uuid.New().String()
	expectedVideo := &metadata.FileMetadata{
		FileID:     videoID,
		Title:      "测试视频",
		Resolution: "1920x1080",
	}

	mockMetadataService.EXPECT().GetMetadata(ctx, videoID).Return(expectedVideo, nil)

	req := &api.VideoDetailRequest{VideoID: videoID}
	resp, err := service.GetVideoDetail(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	require.NotNil(t, resp.Video)
	assert.Equal(t, videoID, resp.Video.ID)
}