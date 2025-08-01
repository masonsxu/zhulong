package download

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/manteia/zhulong/pkg/storage"
)

// TestDownloadService_DownloadFile 测试文件下载
func TestDownloadService_DownloadFile(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	downloadService := NewDownloadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "test-video.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传测试文件
	testData := []byte("这是一个测试视频文件内容，用于测试文件下载功能")
	_, err = storageService.UploadFile(ctx, bucketName, objectName, testData, "video/mp4")
	require.NoError(t, err)
	defer func() {
		_ = storageService.DeleteFile(ctx, bucketName, objectName)
	}()

	// 测试下载文件
	downloadRequest := &DownloadRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}

	result, err := downloadService.DownloadFile(ctx, downloadRequest)
	assert.NoError(t, err, "文件下载应该成功")
	require.NotNil(t, result, "下载结果不应为空")

	// 验证下载结果
	assert.Equal(t, testData, result.Data, "下载的数据应该匹配")
	assert.Equal(t, int64(len(testData)), result.Size, "文件大小应该匹配")
	assert.Equal(t, "video/mp4", result.ContentType, "内容类型应该匹配")
	assert.Equal(t, objectName, result.ObjectName, "对象名应该匹配")
	assert.NotEmpty(t, result.ETag, "ETag不应为空")
}

// TestDownloadService_GeneratePresignedURL 测试预签名URL生成
func TestDownloadService_GeneratePresignedURL(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	downloadService := NewDownloadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "test-video.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传测试文件
	testData := []byte("测试视频内容")
	_, err = storageService.UploadFile(ctx, bucketName, objectName, testData, "video/mp4")
	require.NoError(t, err)
	defer func() {
		_ = storageService.DeleteFile(ctx, bucketName, objectName)
	}()

	// 测试生成预签名URL
	presignRequest := &PresignedURLRequest{
		BucketName: bucketName,
		ObjectName: objectName,
		ExpiresIn:  24 * time.Hour, // 24小时过期
		Method:     "GET",
	}

	result, err := downloadService.GeneratePresignedURL(ctx, presignRequest)
	assert.NoError(t, err, "生成预签名URL应该成功")
	require.NotNil(t, result, "预签名URL结果不应为空")

	// 验证预签名URL结果
	assert.NotEmpty(t, result.URL, "预签名URL不应为空")
	assert.True(t, strings.HasPrefix(result.URL, "http"), "URL应该是有效的HTTP URL")
	assert.True(t, result.ExpiresAt.After(time.Now()), "URL应该在未来过期")
	assert.True(t, result.ExpiresAt.Before(time.Now().Add(25*time.Hour)), "过期时间应该在预期范围内")
	assert.Equal(t, bucketName, result.BucketName, "存储桶名应该匹配")
	assert.Equal(t, objectName, result.ObjectName, "对象名应该匹配")
}

// TestDownloadService_GeneratePresignedURL_CustomExpiration 测试自定义过期时间
func TestDownloadService_GeneratePresignedURL_CustomExpiration(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	downloadService := NewDownloadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "test-video.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传测试文件
	testData := []byte("测试视频内容")
	_, err = storageService.UploadFile(ctx, bucketName, objectName, testData, "video/mp4")
	require.NoError(t, err)
	defer func() {
		_ = storageService.DeleteFile(ctx, bucketName, objectName)
	}()

	testCases := []struct {
		name      string
		expiresIn time.Duration
	}{
		{"1小时过期", 1 * time.Hour},
		{"30分钟过期", 30 * time.Minute},
		{"1天过期", 24 * time.Hour},
		{"7天过期", 7 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			presignRequest := &PresignedURLRequest{
				BucketName: bucketName,
				ObjectName: objectName,
				ExpiresIn:  tc.expiresIn,
				Method:     "GET",
			}

			result, err := downloadService.GeneratePresignedURL(ctx, presignRequest)
			assert.NoError(t, err, "生成预签名URL应该成功")
			require.NotNil(t, result, "预签名URL结果不应为空")

			// 验证过期时间在预期范围内
			expectedExpiry := time.Now().Add(tc.expiresIn)
			assert.True(t, result.ExpiresAt.Before(expectedExpiry.Add(1*time.Minute)), "过期时间应该在预期范围内")
		})
	}
}

// TestDownloadService_GeneratePresignedURL_DifferentMethods 测试不同HTTP方法
func TestDownloadService_GeneratePresignedURL_DifferentMethods(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	downloadService := NewDownloadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "test-video.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传测试文件
	testData := []byte("测试视频内容")
	_, err = storageService.UploadFile(ctx, bucketName, objectName, testData, "video/mp4")
	require.NoError(t, err)
	defer func() {
		_ = storageService.DeleteFile(ctx, bucketName, objectName)
	}()

	methods := []string{"GET", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run("方法_"+method, func(t *testing.T) {
			presignRequest := &PresignedURLRequest{
				BucketName: bucketName,
				ObjectName: objectName,
				ExpiresIn:  1 * time.Hour,
				Method:     method,
			}

			result, err := downloadService.GeneratePresignedURL(ctx, presignRequest)
			assert.NoError(t, err, "生成预签名URL应该成功")
			require.NotNil(t, result, "预签名URL结果不应为空")
			assert.NotEmpty(t, result.URL, "预签名URL不应为空")
		})
	}
}

// TestDownloadService_ValidateDownloadRequest 测试下载请求验证
func TestDownloadService_ValidateDownloadRequest(t *testing.T) {
	downloadService := NewDownloadService(nil)

	// 测试有效请求
	validRequest := &DownloadRequest{
		BucketName: "test-bucket",
		ObjectName: "test-video.mp4",
	}

	err := downloadService.ValidateDownloadRequest(validRequest)
	assert.NoError(t, err, "有效请求应该通过验证")

	// 测试无效请求
	invalidCases := []struct {
		name    string
		request *DownloadRequest
		errMsg  string
	}{
		{
			name: "空存储桶名",
			request: &DownloadRequest{
				BucketName: "",
				ObjectName: "test-video.mp4",
			},
			errMsg: "存储桶名不能为空",
		},
		{
			name: "空对象名",
			request: &DownloadRequest{
				BucketName: "test-bucket",
				ObjectName: "",
			},
			errMsg: "对象名不能为空",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := downloadService.ValidateDownloadRequest(tc.request)
			assert.Error(t, err, "无效请求应该验证失败")
			assert.Contains(t, err.Error(), tc.errMsg, "错误信息应该包含预期内容")
		})
	}
}

// TestDownloadService_ValidatePresignedURLRequest 测试预签名URL请求验证
func TestDownloadService_ValidatePresignedURLRequest(t *testing.T) {
	downloadService := NewDownloadService(nil)

	// 测试有效请求
	validRequest := &PresignedURLRequest{
		BucketName: "test-bucket",
		ObjectName: "test-video.mp4",
		ExpiresIn:  24 * time.Hour,
		Method:     "GET",
	}

	err := downloadService.ValidatePresignedURLRequest(validRequest)
	assert.NoError(t, err, "有效请求应该通过验证")

	// 测试无效请求
	invalidCases := []struct {
		name    string
		request *PresignedURLRequest
		errMsg  string
	}{
		{
			name: "空存储桶名",
			request: &PresignedURLRequest{
				BucketName: "",
				ObjectName: "test-video.mp4",
				ExpiresIn:  24 * time.Hour,
				Method:     "GET",
			},
			errMsg: "存储桶名不能为空",
		},
		{
			name: "空对象名",
			request: &PresignedURLRequest{
				BucketName: "test-bucket",
				ObjectName: "",
				ExpiresIn:  24 * time.Hour,
				Method:     "GET",
			},
			errMsg: "对象名不能为空",
		},
		{
			name: "无效过期时间",
			request: &PresignedURLRequest{
				BucketName: "test-bucket",
				ObjectName: "test-video.mp4",
				ExpiresIn:  0,
				Method:     "GET",
			},
			errMsg: "过期时间必须大于0",
		},
		{
			name: "过期时间过长",
			request: &PresignedURLRequest{
				BucketName: "test-bucket",
				ObjectName: "test-video.mp4",
				ExpiresIn:  8 * 24 * time.Hour, // 8天
				Method:     "GET",
			},
			errMsg: "过期时间不能超过7天",
		},
		{
			name: "无效HTTP方法",
			request: &PresignedURLRequest{
				BucketName: "test-bucket",
				ObjectName: "test-video.mp4",
				ExpiresIn:  24 * time.Hour,
				Method:     "INVALID",
			},
			errMsg: "不支持的HTTP方法",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := downloadService.ValidatePresignedURLRequest(tc.request)
			assert.Error(t, err, "无效请求应该验证失败")
			assert.Contains(t, err.Error(), tc.errMsg, "错误信息应该包含预期内容")
		})
	}
}

// TestDownloadService_FileNotFound 测试文件不存在的情况
func TestDownloadService_FileNotFound(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	downloadService := NewDownloadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "non-existent-file.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 测试下载不存在的文件
	downloadRequest := &DownloadRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}

	result, err := downloadService.DownloadFile(ctx, downloadRequest)
	assert.Error(t, err, "下载不存在的文件应该失败")
	assert.Nil(t, result, "失败时结果应为空")
	assert.Contains(t, err.Error(), "文件不存在", "错误信息应该表明文件不存在")
}

// TestDownloadService_GenerateDownloadURL 测试生成下载URL
func TestDownloadService_GenerateDownloadURL(t *testing.T) {
	downloadService := NewDownloadService(nil)

	// 测试生成下载URL
	request := &DownloadURLRequest{
		BucketName: "test-bucket",
		ObjectName: "videos/2025/08/uuid-test.mp4",
		ExpiresIn:  24 * time.Hour,
	}

	result, err := downloadService.GenerateDownloadURL(request)
	assert.NoError(t, err, "生成下载URL应该成功")
	require.NotNil(t, result, "下载URL结果不应为空")

	// 验证结果
	assert.NotEmpty(t, result.DownloadURL, "下载URL不应为空")
	assert.Equal(t, request.BucketName, result.BucketName, "存储桶名应该匹配")
	assert.Equal(t, request.ObjectName, result.ObjectName, "对象名应该匹配")
	assert.True(t, result.ExpiresAt.After(time.Now()), "URL应该在未来过期")
}

// isStorageAvailable 检查存储服务是否可用
func isStorageAvailable() bool {
	storageConfig := &storage.MinIOConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		UseSSL:    false,
		Region:    "us-east-1",
	}

	storageService, err := storage.NewMinIOStorage(storageConfig)
	if err != nil {
		return false
	}

	ctx := context.Background()
	err = storageService.TestConnection(ctx)
	return err == nil
}

// setupTestStorage 设置测试存储服务
func setupTestStorage(t *testing.T) storage.StorageInterface {
	storageConfig := &storage.MinIOConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		UseSSL:    false,
		Region:    "us-east-1",
	}

	storageService, err := storage.NewMinIOStorage(storageConfig)
	require.NoError(t, err)

	return storageService
}
