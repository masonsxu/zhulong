package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/metadata"
)

// TestGetVideoDetail_Success 测试成功获取视频详情
func TestGetVideoDetail_Success(t *testing.T) {
	// 准备测试数据
	videoID := uuid.New().String()
	expectedVideo := &metadata.FileMetadata{
		FileID:      videoID,
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/01/test.mp4",
		FileName:    "test.mp4",
		Title:       "测试视频",
		Description: "这是一个测试视频",
		ContentType: "video/mp4",
		FileSize:    1024000,
		Duration:    300, // 5分钟
		Resolution:  "1920x1080",
		Thumbnail:   "thumbnails/2024/01/test.jpg",
		CreatedBy:   "system",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 创建测试服务实例
	service := createTestVideoService(t)

	// 保存测试数据
	err := service.metadataService.SaveMetadata(context.Background(), expectedVideo)
	if err != nil {
		t.Fatalf("保存测试数据失败: %v", err)
	}

	// 测试获取视频详情
	req := &api.VideoDetailRequest{
		VideoID: videoID,
	}

	resp, err := service.GetVideoDetail(context.Background(), req)

	// 验证结果
	if err != nil {
		t.Fatalf("获取视频详情失败: %v", err)
	}

	if resp == nil {
		t.Fatal("响应不能为空")
	}

	if resp.Base == nil {
		t.Fatal("基础响应不能为空")
	}

	if resp.Base.Code != 0 {
		t.Errorf("预期code为0，实际为%d，消息：%s", resp.Base.Code, resp.Base.Message)
	}

	if resp.Video == nil {
		t.Fatal("视频信息不能为空")
	}

	// 验证视频信息
	video := resp.Video
	if video.ID != videoID {
		t.Errorf("预期视频ID为%s，实际为%s", videoID, video.ID)
	}

	if video.Title != expectedVideo.Title {
		t.Errorf("预期标题为%s，实际为%s", expectedVideo.Title, video.Title)
	}

	if video.Filename != expectedVideo.FileName {
		t.Errorf("预期文件名为%s，实际为%s", expectedVideo.FileName, video.Filename)
	}

	if video.ContentType != expectedVideo.ContentType {
		t.Errorf("预期内容类型为%s，实际为%s", expectedVideo.ContentType, video.ContentType)
	}

	if video.Size != expectedVideo.FileSize {
		t.Errorf("预期文件大小为%d，实际为%d", expectedVideo.FileSize, video.Size)
	}

	if video.Duration != expectedVideo.Duration {
		t.Errorf("预期时长为%d，实际为%d", expectedVideo.Duration, video.Duration)
	}

	if video.StoragePath != expectedVideo.ObjectName {
		t.Errorf("预期存储路径为%s，实际为%s", expectedVideo.ObjectName, video.StoragePath)
	}

	if video.ThumbnailPath != expectedVideo.Thumbnail {
		t.Errorf("预期缩略图路径为%s，实际为%s", expectedVideo.Thumbnail, video.ThumbnailPath)
	}

	// 验证分辨率解析
	if video.Width != 1920 {
		t.Errorf("预期宽度为1920，实际为%d", video.Width)
	}

	if video.Height != 1080 {
		t.Errorf("预期高度为1080，实际为%d", video.Height)
	}
}

// TestGetVideoDetail_NotFound 测试视频不存在的情况
func TestGetVideoDetail_NotFound(t *testing.T) {
	// 创建测试服务实例
	service := createTestVideoService(t)

	// 测试获取不存在的视频
	req := &api.VideoDetailRequest{
		VideoID: "non-existent-id",
	}

	resp, err := service.GetVideoDetail(context.Background(), req)

	// 验证结果
	if err != nil {
		t.Fatalf("不应该返回错误: %v", err)
	}

	if resp == nil {
		t.Fatal("响应不能为空")
	}

	if resp.Base == nil {
		t.Fatal("基础响应不能为空")
	}

	if resp.Base.Code != 3001 {
		t.Errorf("预期错误码为3001，实际为%d", resp.Base.Code)
	}

	if resp.Video != nil {
		t.Error("视频不存在时，video字段应该为空")
	}
}

// TestGetVideoDetail_EmptyVideoID 测试空视频ID的情况
func TestGetVideoDetail_EmptyVideoID(t *testing.T) {
	// 创建测试服务实例
	service := createTestVideoService(t)

	// 测试空视频ID
	req := &api.VideoDetailRequest{
		VideoID: "",
	}

	resp, err := service.GetVideoDetail(context.Background(), req)

	// 验证结果
	if err != nil {
		t.Fatalf("不应该返回错误: %v", err)
	}

	if resp == nil {
		t.Fatal("响应不能为空")
	}

	if resp.Base == nil {
		t.Fatal("基础响应不能为空")
	}

	if resp.Base.Code != 3000 {
		t.Errorf("预期错误码为3000，实际为%d", resp.Base.Code)
	}

	if resp.Video != nil {
		t.Error("参数错误时，video字段应该为空")
	}
}

// TestGetVideoDetail_InvalidVideoID 测试无效视频ID格式
func TestGetVideoDetail_InvalidVideoID(t *testing.T) {
	// 创建测试服务实例
	service := createTestVideoService(t)

	// 测试各种无效格式的视频ID
	invalidIDs := []string{
		"   ", // 只有空格
		"invalid-format-with-special-chars!@#",
		"too-short",
	}

	for _, invalidID := range invalidIDs {
		t.Run("InvalidID_"+invalidID, func(t *testing.T) {
			req := &api.VideoDetailRequest{
				VideoID: invalidID,
			}

			resp, err := service.GetVideoDetail(context.Background(), req)

			// 验证结果
			if err != nil {
				t.Fatalf("不应该返回错误: %v", err)
			}

			if resp == nil {
				t.Fatal("响应不能为空")
			}

			if resp.Base == nil {
				t.Fatal("基础响应不能为空")
			}

			// 应该返回参数错误或不存在错误
			if resp.Base.Code != 3000 && resp.Base.Code != 3001 {
				t.Errorf("预期错误码为3000或3001，实际为%d", resp.Base.Code)
			}

			if resp.Video != nil {
				t.Error("错误情况下，video字段应该为空")
			}
		})
	}
}

// TestGetVideoDetail_WithoutResolution 测试没有分辨率信息的视频
func TestGetVideoDetail_WithoutResolution(t *testing.T) {
	// 准备测试数据（没有分辨率）
	videoID := uuid.New().String()
	expectedVideo := &metadata.FileMetadata{
		FileID:      videoID,
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/01/no-resolution.mp4",
		FileName:    "no-resolution.mp4",
		Title:       "无分辨率信息视频",
		ContentType: "video/mp4",
		FileSize:    512000,
		Duration:    180,
		Resolution:  "", // 空分辨率
		CreatedBy:   "system",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 创建测试服务实例
	service := createTestVideoService(t)

	// 保存测试数据
	err := service.metadataService.SaveMetadata(context.Background(), expectedVideo)
	if err != nil {
		t.Fatalf("保存测试数据失败: %v", err)
	}

	// 测试获取视频详情
	req := &api.VideoDetailRequest{
		VideoID: videoID,
	}

	resp, err := service.GetVideoDetail(context.Background(), req)

	// 验证结果
	if err != nil {
		t.Fatalf("获取视频详情失败: %v", err)
	}

	if resp == nil || resp.Base == nil || resp.Video == nil {
		t.Fatal("响应结构不完整")
	}

	if resp.Base.Code != 0 {
		t.Errorf("预期成功，实际错误码：%d", resp.Base.Code)
	}

	// 没有分辨率信息时，宽度和高度应该为0
	if resp.Video.Width != 0 {
		t.Errorf("预期宽度为0，实际为%d", resp.Video.Width)
	}

	if resp.Video.Height != 0 {
		t.Errorf("预期高度为0，实际为%d", resp.Video.Height)
	}
}