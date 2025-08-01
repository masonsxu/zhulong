package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/metadata"
)

func TestVideoService_GetVideoList(t *testing.T) {
	// 创建视频服务（使用内存存储）
	service := createTestVideoService(t)

	// 准备测试数据
	testVideos := []*metadata.FileMetadata{
		{
			FileID:      "video1",
			Title:       "测试视频1",
			FileName:    "test1.mp4",
			ContentType: "video/mp4",
			FileSize:    1024000,
			Duration:    60,
			Resolution:  "1920x1080",
			Thumbnail:   "thumbnails/video1.jpg",
			CreatedBy:   "system",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			FileID:      "video2",
			Title:       "测试视频2",
			FileName:    "test2.mp4",
			ContentType: "video/mp4",
			FileSize:    2048000,
			Duration:    120,
			Resolution:  "1280x720",
			Thumbnail:   "thumbnails/video2.jpg",
			CreatedBy:   "system",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			FileID:      "video3",
			Title:       "测试视频3",
			FileName:    "test3.mp4",
			ContentType: "video/mp4",
			FileSize:    512000,
			Duration:    30,
			Resolution:  "1920x1080",
			Thumbnail:   "thumbnails/video3.jpg",
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// 保存测试数据
	ctx := context.Background()
	for _, video := range testVideos {
		err := service.metadataService.SaveMetadata(ctx, video)
		require.NoError(t, err)
	}

	t.Run("获取视频列表_默认参数", func(t *testing.T) {
		req := &api.VideoListRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := service.GetVideoList(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Base.Code)
		assert.Equal(t, "获取成功", resp.Base.Message)
		assert.Equal(t, int32(3), resp.Total)
		assert.Len(t, resp.Videos, 3)

		// 验证排序（应该按创建时间倒序）
		assert.Equal(t, "video3", resp.Videos[0].ID)
		assert.Equal(t, "video2", resp.Videos[1].ID)
		assert.Equal(t, "video1", resp.Videos[2].ID)
	})

	t.Run("获取视频列表_分页", func(t *testing.T) {
		req := &api.VideoListRequest{
			Page:     1,
			PageSize: 2,
		}

		resp, err := service.GetVideoList(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Base.Code)
		assert.Equal(t, int32(3), resp.Total)
		assert.Len(t, resp.Videos, 2)

		// 第二页
		req.Page = 2
		resp, err = service.GetVideoList(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Videos, 1)
		assert.Equal(t, "video1", resp.Videos[0].ID)
	})

	t.Run("获取视频列表_排序", func(t *testing.T) {
		req := &api.VideoListRequest{
			Page:     1,
			PageSize: 10,
			SortBy:   "title",
		}

		resp, err := service.GetVideoList(ctx, req)
		require.NoError(t, err)
		// 按title降序排列应该是：测试视频3, 测试视频2, 测试视频1
		assert.Equal(t, "测试视频3", resp.Videos[0].Title)
		assert.Equal(t, "测试视频2", resp.Videos[1].Title)
		assert.Equal(t, "测试视频1", resp.Videos[2].Title)
	})

	t.Run("获取视频列表_空结果", func(t *testing.T) {
		// 创建新的服务实例，没有数据
		emptyService := createTestVideoService(t)

		req := &api.VideoListRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := emptyService.GetVideoList(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, int32(0), resp.Base.Code)
		assert.Equal(t, int32(0), resp.Total)
		assert.Len(t, resp.Videos, 0)
	})

	t.Run("获取视频列表_无效参数", func(t *testing.T) {
		req := &api.VideoListRequest{
			Page:     -1, // 无效页码
			PageSize: 10,
		}

		resp, err := service.GetVideoList(ctx, req)
		require.NoError(t, err)
		assert.NotEqual(t, int32(0), resp.Base.Code)
		assert.Contains(t, resp.Base.Message, "页码")
	})

	t.Run("获取视频列表_页面大小超限", func(t *testing.T) {
		req := &api.VideoListRequest{
			Page:     1,
			PageSize: 101, // 超过最大限制
		}

		resp, err := service.GetVideoList(ctx, req)
		require.NoError(t, err)
		assert.NotEqual(t, int32(0), resp.Base.Code)
		assert.Contains(t, resp.Base.Message, "页面大小")
	})
}

// createTestVideoService 创建测试用的视频服务
func createTestVideoService(t *testing.T) *VideoService {
	return &VideoService{
		metadataService: metadata.NewMetadataService(),
	}
}