package service

import (
	"context"
	"testing"
	"time"

	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/metadata"
)

// TestVideoService_DeleteVideo_Success 测试成功删除视频（包含缩略图）
func TestVideoService_DeleteVideo_Success(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)
	ctx := context.Background()

	// 创建测试视频元数据
	testMetadata := &metadata.FileMetadata{
		FileID:      "test-video-id-123",
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/01/test-video-id-123.mp4",
		FileName:    "test-video.mp4",
		Title:       "测试视频",
		Description: "这是一个测试视频",
		ContentType: "video/mp4",
		FileSize:    1024000,
		Duration:    60,
		Resolution:  "1920x1080",
		Thumbnail:   "thumbnails/2024/01/test-video-id-123.jpg",
		Tags:        []string{"test"},
		CreatedBy:   "test-user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存测试元数据
	err := service.metadataService.SaveMetadata(ctx, testMetadata)
	if err != nil {
		t.Fatalf("保存测试元数据失败: %v", err)
	}

	// 模拟上传视频文件和缩略图文件到存储
	// 在实际测试中，这些文件应该存在于存储中
	// 这里我们假设文件已经存在

	// 执行删除操作
	req := &api.VideoDeleteRequest{
		VideoID: "test-video-id-123",
	}

	resp, err := service.DeleteVideo(ctx, req)

	// 验证结果
	if err != nil {
		t.Fatalf("删除视频失败: %v", err)
	}

	if resp == nil {
		t.Fatal("响应不能为空")
	}

	if resp.Base == nil {
		t.Fatal("响应Base不能为空")
	}

	if resp.Base.Code != 0 {
		t.Errorf("期望响应码为0，实际为%d，消息: %s", resp.Base.Code, resp.Base.Message)
	}

	if resp.Base.Message != "删除成功" {
		t.Errorf("期望消息为'删除成功'，实际为'%s'", resp.Base.Message)
	}

	// 验证元数据已被删除
	_, err = service.metadataService.GetMetadata(ctx, req.VideoID)
	if err == nil {
		t.Error("元数据应该已被删除")
	}
}

// TestVideoService_DeleteVideo_SuccessWithoutThumbnail 测试成功删除视频（无缩略图）
func TestVideoService_DeleteVideo_SuccessWithoutThumbnail(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)
	ctx := context.Background()

	// 创建测试视频元数据（无缩略图）
	testMetadata := &metadata.FileMetadata{
		FileID:      "test-video-id-no-thumb",
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/01/test-video-id-no-thumb.mp4",
		FileName:    "test-video-no-thumb.mp4",
		Title:       "无缩略图测试视频",
		Description: "这是一个无缩略图的测试视频",
		ContentType: "video/mp4",
		FileSize:    512000,
		Duration:    30,
		Resolution:  "1280x720",
		Thumbnail:   "", // 无缩略图
		Tags:        []string{"test"},
		CreatedBy:   "test-user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存测试元数据
	err := service.metadataService.SaveMetadata(ctx, testMetadata)
	if err != nil {
		t.Fatalf("保存测试元数据失败: %v", err)
	}

	// 执行删除操作
	req := &api.VideoDeleteRequest{
		VideoID: "test-video-id-no-thumb",
	}

	resp, err := service.DeleteVideo(ctx, req)

	// 验证结果
	if err != nil {
		t.Fatalf("删除视频失败: %v", err)
	}

	if resp.Base.Code != 0 {
		t.Errorf("期望响应码为0，实际为%d，消息: %s", resp.Base.Code, resp.Base.Message)
	}

	// 验证元数据已被删除
	_, err = service.metadataService.GetMetadata(ctx, req.VideoID)
	if err == nil {
		t.Error("元数据应该已被删除")
	}
}

// TestVideoService_DeleteVideo_VideoNotExists 测试视频不存在的情况
func TestVideoService_DeleteVideo_VideoNotExists(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)
	ctx := context.Background()

	// 执行删除不存在的视频
	req := &api.VideoDeleteRequest{
		VideoID: "non-existent-video-id",
	}

	resp, err := service.DeleteVideo(ctx, req)

	// 验证结果
	if err != nil {
		t.Fatalf("不应该返回错误: %v", err)
	}

	if resp.Base.Code != 5001 {
		t.Errorf("期望响应码为5001，实际为%d", resp.Base.Code)
	}

	expectedMessage := "视频不存在"
	if resp.Base.Message != expectedMessage {
		t.Errorf("期望消息为'%s'，实际为'%s'", expectedMessage, resp.Base.Message)
	}
}

// TestVideoService_DeleteVideo_EmptyVideoID 测试空视频ID的情况
func TestVideoService_DeleteVideo_EmptyVideoID(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)
	ctx := context.Background()

	testCases := []struct {
		name     string
		videoID  string
		expected int32
	}{
		{
			name:     "空字符串",
			videoID:  "",
			expected: 5000,
		},
		{
			name:     "只有空格",
			videoID:  "   ",
			expected: 5000,
		},
		{
			name:     "制表符和空格",
			videoID:  "\t  \n  ",
			expected: 5000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &api.VideoDeleteRequest{
				VideoID: tc.videoID,
			}

			resp, err := service.DeleteVideo(ctx, req)

			if err != nil {
				t.Fatalf("不应该返回错误: %v", err)
			}

			if resp.Base.Code != tc.expected {
				t.Errorf("期望响应码为%d，实际为%d", tc.expected, resp.Base.Code)
			}

			expectedMessage := "视频ID不能为空"
			if resp.Base.Message != expectedMessage {
				t.Errorf("期望消息为'%s'，实际为'%s'", expectedMessage, resp.Base.Message)
			}
		})
	}
}

// TestVideoService_DeleteVideo_StorageDeleteFailure 测试存储删除失败的情况
func TestVideoService_DeleteVideo_StorageDeleteFailure(t *testing.T) {
	// 这个测试需要mock存储接口来模拟删除失败
	// 由于当前没有mock框架，我们先跳过这个测试
	// 在实际实现中，应该使用mock来测试各种失败场景
	t.Skip("需要mock存储接口来测试删除失败场景")
}

// TestVideoService_DeleteVideo_MetadataDeleteFailure 测试元数据删除失败的情况
func TestVideoService_DeleteVideo_MetadataDeleteFailure(t *testing.T) {
	// 这个测试需要mock元数据服务来模拟删除失败
	// 由于当前没有mock框架，我们先跳过这个测试
	// 在实际实现中，应该使用mock来测试各种失败场景
	t.Skip("需要mock元数据服务来测试删除失败场景")
}

