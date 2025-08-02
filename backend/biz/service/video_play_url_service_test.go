package service

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/metadata"
	"github.com/manteia/zhulong/pkg/storage"
)

// mockStorageClient 模拟存储客户端，用于测试
type mockStorageClient struct{}

func (m *mockStorageClient) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockStorageClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return true, nil
}

func (m *mockStorageClient) CreateBucket(ctx context.Context, bucketName string) error {
	return nil
}

func (m *mockStorageClient) RemoveBucket(ctx context.Context, bucketName string) error {
	return nil
}

func (m *mockStorageClient) UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) (*storage.UploadResult, error) {
	return &storage.UploadResult{}, nil
}

func (m *mockStorageClient) DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	return []byte{}, nil
}

func (m *mockStorageClient) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	return true, nil
}

func (m *mockStorageClient) GetFileInfo(ctx context.Context, bucketName, objectName string) (*storage.FileInfo, error) {
	return &storage.FileInfo{}, nil
}

func (m *mockStorageClient) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	return nil
}

func (m *mockStorageClient) ListFiles(ctx context.Context, bucketName, prefix string) ([]*storage.FileInfo, error) {
	return []*storage.FileInfo{}, nil
}

func (m *mockStorageClient) GetPresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	// 返回模拟的预签名URL
	return "https://test-minio.example.com/" + bucketName + "/" + objectName + "?expiry=" + expiry.String(), nil
}

func (m *mockStorageClient) GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration, method string) (string, error) {
	return m.GetPresignedURL(ctx, bucketName, objectName, expiry)
}

// createTestVideoServiceWithStorage 创建测试用的视频服务（包含存储客户端）
func createTestVideoServiceWithStorage(t *testing.T) *VideoService {
	return &VideoService{
		metadataService: metadata.NewMetadataService(),
		storageClient:   &mockStorageClient{},
	}
}

// TestGetVideoPlayURL_Success 测试成功获取播放URL
func TestGetVideoPlayURL_Success(t *testing.T) {
	// 准备测试环境
	service := createTestVideoServiceWithStorage(t)

	// 创建测试视频元数据
	testVideoID := "test-video-id-001"
	testMetadata := &metadata.FileMetadata{
		FileID:      testVideoID,
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/08/test-video-001.mp4",
		FileName:    "test-video.mp4",
		Title:       "测试视频",
		ContentType: "video/mp4",
		FileSize:    1024000, // 1MB
		Duration:    120,     // 2分钟
		Resolution:  "1920x1080",
		CreatedBy:   "system", // 添加创建者字段
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存测试元数据
	err := service.metadataService.SaveMetadata(context.Background(), testMetadata)
	if err != nil {
		t.Fatalf("保存测试元数据失败: %v", err)
	}

	// 清理函数
	defer func() {
		service.metadataService.DeleteMetadata(context.Background(), testVideoID)
	}()

	// 测试用例1: 使用默认过期时间
	req := &api.VideoPlayURLRequest{
		VideoID:       testVideoID,
		ExpireSeconds: 3600, // 1小时
	}

	resp, err := service.GetVideoPlayURL(context.Background(), req)
	
	// 验证结果
	if err != nil {
		t.Errorf("获取播放URL失败: %v", err)
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	if resp.Base.Code != 0 {
		t.Errorf("期望响应码为0，实际为%d，消息: %s", resp.Base.Code, resp.Base.Message)
	}
	
	if resp.PlayURL == "" {
		t.Error("播放URL为空")
	}
	
	if resp.ExpiresAt <= 0 {
		t.Error("过期时间戳无效")
	}
	
	// 验证URL格式和有效性
	if !strings.Contains(resp.PlayURL, testMetadata.ObjectName) {
		t.Error("播放URL不包含正确的对象名称")
	}
	
	// 验证过期时间大约为1小时后
	expectedExpiry := time.Now().Add(time.Hour).UnixMilli()
	timeDiff := resp.ExpiresAt - expectedExpiry
	if timeDiff < -60000 || timeDiff > 60000 { // 允许1分钟误差
		t.Errorf("过期时间不正确，期望约%d，实际%d", expectedExpiry, resp.ExpiresAt)
	}
}

// TestGetVideoPlayURL_CustomExpiry 测试自定义过期时间
func TestGetVideoPlayURL_CustomExpiry(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)

	// 创建测试视频元数据
	testVideoID := "test-video-id-002"
	testMetadata := &metadata.FileMetadata{
		FileID:      testVideoID,
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/08/test-video-002.mp4",
		FileName:    "test-video2.mp4",
		Title:       "测试视频2",
		ContentType: "video/mp4",
		FileSize:    2048000, // 2MB
		Duration:    180,     // 3分钟
		Resolution:  "1280x720",
		CreatedBy:   "system", // 添加创建者字段
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存测试元数据
	err := service.metadataService.SaveMetadata(context.Background(), testMetadata)
	if err != nil {
		t.Fatalf("保存测试元数据失败: %v", err)
	}

	defer func() {
		service.metadataService.DeleteMetadata(context.Background(), testVideoID)
	}()

	// 测试自定义过期时间（30分钟）
	req := &api.VideoPlayURLRequest{
		VideoID:       testVideoID,
		ExpireSeconds: 1800, // 30分钟
	}

	resp, err := service.GetVideoPlayURL(context.Background(), req)
	
	if err != nil {
		t.Errorf("获取播放URL失败: %v", err)
	}
	
	if resp.Base.Code != 0 {
		t.Errorf("期望响应码为0，实际为%d", resp.Base.Code)
	}
	
	// 验证过期时间约为30分钟后
	expectedExpiry := time.Now().Add(30 * time.Minute).UnixMilli()
	timeDiff := resp.ExpiresAt - expectedExpiry
	if timeDiff < -60000 || timeDiff > 60000 { // 允许1分钟误差
		t.Errorf("过期时间不正确，期望约%d，实际%d", expectedExpiry, resp.ExpiresAt)
	}
}

// TestGetVideoPlayURL_VideoNotFound 测试视频不存在的情况
func TestGetVideoPlayURL_VideoNotFound(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)

	// 使用不存在的视频ID
	req := &api.VideoPlayURLRequest{
		VideoID:       "non-existent-video-id",
		ExpireSeconds: 3600,
	}

	resp, err := service.GetVideoPlayURL(context.Background(), req)
	
	// 应该返回错误响应，但不应该有Go错误
	if err != nil {
		t.Errorf("不应该有Go错误: %v", err)
	}
	
	if resp == nil {
		t.Fatal("响应为空")
	}
	
	if resp.Base.Code != 4001 {
		t.Errorf("期望响应码为4001（视频不存在），实际为%d", resp.Base.Code)
	}
	
	if resp.PlayURL != "" {
		t.Error("视频不存在时播放URL应该为空")
	}
	
	if resp.ExpiresAt != 0 {
		t.Error("视频不存在时过期时间应该为0")
	}
}

// TestGetVideoPlayURL_EmptyVideoID 测试空的视频ID
func TestGetVideoPlayURL_EmptyVideoID(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)

	// 测试空的视频ID
	req := &api.VideoPlayURLRequest{
		VideoID:       "",
		ExpireSeconds: 3600,
	}

	resp, err := service.GetVideoPlayURL(context.Background(), req)
	
	if err != nil {
		t.Errorf("不应该有Go错误: %v", err)
	}
	
	if resp.Base.Code != 4000 {
		t.Errorf("期望响应码为4000（参数错误），实际为%d", resp.Base.Code)
	}
}

// TestGetVideoPlayURL_InvalidExpiry 测试无效的过期时间
func TestGetVideoPlayURL_InvalidExpiry(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)

	// 创建测试视频元数据
	testVideoID := "test-video-id-003"
	testMetadata := &metadata.FileMetadata{
		FileID:      testVideoID,
		BucketName:  "zhulong-videos",
		ObjectName:  "videos/2024/08/test-video-003.mp4",
		FileName:    "test-video3.mp4",
		Title:       "测试视频3",
		ContentType: "video/mp4",
		FileSize:    1024000,
		Duration:    60,
		CreatedBy:   "system", // 添加创建者字段
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := service.metadataService.SaveMetadata(context.Background(), testMetadata)
	if err != nil {
		t.Fatalf("保存测试元数据失败: %v", err)
	}

	defer func() {
		service.metadataService.DeleteMetadata(context.Background(), testVideoID)
	}()

	// 测试用例：负数过期时间
	testCases := []struct {
		name          string
		expireSeconds int32
		expectedCode  int32
	}{
		{"负数过期时间", -1, 4000},
		{"零过期时间", 0, 0},                       // 0会被设为默认值3600，应该成功
		{"过长过期时间", 8*24*3600, 4000}, // 8天，超过7天限制
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &api.VideoPlayURLRequest{
				VideoID:       testVideoID,
				ExpireSeconds: tc.expireSeconds,
			}

			resp, err := service.GetVideoPlayURL(context.Background(), req)
			
			if err != nil {
				t.Errorf("不应该有Go错误: %v", err)
			}
			
			if resp.Base.Code != tc.expectedCode {
				t.Errorf("期望响应码为%d，实际为%d", tc.expectedCode, resp.Base.Code)
			}
		})
	}
}

// TestGetVideoPlayURL_WhitespaceVideoID 测试只包含空格的视频ID
func TestGetVideoPlayURL_WhitespaceVideoID(t *testing.T) {
	service := createTestVideoServiceWithStorage(t)

	testCases := []string{
		"   ",     // 只有空格
		"\t\t\t", // 只有制表符
		"\n\n",   // 只有换行符
		" \t \n ", // 混合空白字符
	}

	for i, videoID := range testCases {
		t.Run("空白字符测试_"+strconv.Itoa(i+1), func(t *testing.T) {
			req := &api.VideoPlayURLRequest{
				VideoID:       videoID,
				ExpireSeconds: 3600,
			}

			resp, err := service.GetVideoPlayURL(context.Background(), req)
			
			if err != nil {
				t.Errorf("不应该有Go错误: %v", err)
			}
			
			if resp.Base.Code != 4000 {
				t.Errorf("期望响应码为4000（参数错误），实际为%d", resp.Base.Code)
			}
		})
	}
}