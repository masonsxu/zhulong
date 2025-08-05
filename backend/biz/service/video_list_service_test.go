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
)

func TestVideoService_GetVideoList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetadataService := metadatamocks.NewMockMetadataServiceInterface(ctrl)
	service := &VideoService{metadataService: mockMetadataService}
	ctx := context.Background()

	testVideos := []*metadata.FileMetadata{
		{FileID: "video1", Title: "Video 1"},
	}

	mockMetadataService.EXPECT().ListMetadata(gomock.Any(), gomock.Any()).Return(&metadata.ListMetadataResponse{
		Items: testVideos,
		Total: 1,
	}, nil)

	req := &api.VideoListRequest{Page: 1, PageSize: 10}
	resp, err := service.GetVideoList(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Videos, 1)
	assert.Equal(t, "video1", resp.Videos[0].ID)
}