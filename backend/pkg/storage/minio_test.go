package storage

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/manteia/zhulong/testconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinIOStorage_Creation 测试MinIO存储实例创建
func TestMinIOStorage_Creation(t *testing.T) {
	// 从环境变量获取测试配置
	testConfig := testconfig.GetMinIOTestConfig()
	config := &MinIOConfig{
		Endpoint:  testConfig.GetEndpoint(),
		AccessKey: testConfig.AccessKey,
		SecretKey: testConfig.SecretKey,
		UseSSL:    testConfig.UseSSL,
		Region:    testConfig.Region,
	}

	// 测试正常创建
	storage, err := NewMinIOStorage(config)

	require.NoError(t, err, "创建MinIO存储实例应该成功")
	require.NotNil(t, storage, "存储实例不应为空")
	require.NotNil(t, storage.client, "MinIO客户端不应为空")
	require.NotNil(t, storage.config, "配置不应为空")
}

// TestMinIOStorage_Creation_WithNilConfig 测试使用空配置创建
func TestMinIOStorage_Creation_WithNilConfig(t *testing.T) {
	storage, err := NewMinIOStorage(nil)

	require.Error(t, err, "使用空配置应该返回错误")
	require.Nil(t, storage, "存储实例应为空")
	assert.Contains(t, err.Error(), "配置不能为空", "错误信息应该包含配置为空的提示")
}

// TestMinIOStorage_Connection 测试MinIO连接（需要真实服务）
func TestMinIOStorage_Connection(t *testing.T) {
	if !isMinIOAvailable() {
		t.Skip("跳过测试：MinIO服务不可用")
	}

	storage := setupTestStorage(t)
	ctx := context.Background()

	err := storage.TestConnection(ctx)
	assert.NoError(t, err, "MinIO连接测试应该成功")
}

// TestMinIOStorage_BucketOperations 测试存储桶操作（需要真实服务）
func TestMinIOStorage_BucketOperations(t *testing.T) {
	if !isMinIOAvailable() {
		t.Skip("跳过测试：MinIO服务不可用")
	}

	storage := setupTestStorage(t)
	ctx := context.Background()
	testBucket := "test-bucket-" + generateTestID()

	// 测试存储桶不存在
	exists, err := storage.BucketExists(ctx, testBucket)
	assert.NoError(t, err)
	assert.False(t, exists, "测试存储桶应该不存在")

	// 测试创建存储桶
	err = storage.CreateBucket(ctx, testBucket)
	assert.NoError(t, err, "创建存储桶应该成功")

	// 测试存储桶存在
	exists, err = storage.BucketExists(ctx, testBucket)
	assert.NoError(t, err)
	assert.True(t, exists, "创建后存储桶应该存在")

	// 清理
	defer func() {
		_ = storage.RemoveBucket(ctx, testBucket)
	}()
}

// TestMinIOStorage_FileOperations 测试文件操作（需要真实服务）
func TestMinIOStorage_FileOperations(t *testing.T) {
	if !isMinIOAvailable() {
		t.Skip("跳过测试：MinIO服务不可用")
	}

	storage := setupTestStorage(t)
	ctx := context.Background()
	testBucket := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storage.CreateBucket(ctx, testBucket)
	require.NoError(t, err)
	defer func() {
		_ = storage.RemoveBucket(ctx, testBucket)
	}()

	// 测试文件上传
	testData := []byte("这是测试视频文件内容")
	objectName := "videos/2025/08/test-video.mp4"
	contentType := "video/mp4"

	uploadResult, err := storage.UploadFile(ctx, testBucket, objectName, testData, contentType)
	assert.NoError(t, err, "文件上传应该成功")
	assert.NotNil(t, uploadResult, "上传结果不应为空")
	assert.Equal(t, int64(len(testData)), uploadResult.Size, "上传文件大小应该匹配")

	// 测试文件存在性检查
	exists, err := storage.FileExists(ctx, testBucket, objectName)
	assert.NoError(t, err)
	assert.True(t, exists, "上传后文件应该存在")

	// 测试生成预签名URL
	expiry := time.Hour
	presignedURL, err := storage.GetPresignedURL(ctx, testBucket, objectName, expiry)
	assert.NoError(t, err, "生成预签名URL应该成功")
	assert.NotEmpty(t, presignedURL, "预签名URL不应为空")
	assert.True(t, strings.Contains(presignedURL, objectName), "URL应该包含对象名")

	// 测试获取文件信息
	fileInfo, err := storage.GetFileInfo(ctx, testBucket, objectName)
	assert.NoError(t, err, "获取文件信息应该成功")
	assert.Equal(t, objectName, fileInfo.Key, "文件名应该匹配")
	assert.Equal(t, int64(len(testData)), fileInfo.Size, "文件大小应该匹配")
	assert.Equal(t, contentType, fileInfo.ContentType, "内容类型应该匹配")

	// 测试文件删除
	err = storage.DeleteFile(ctx, testBucket, objectName)
	assert.NoError(t, err, "删除文件应该成功")

	// 测试文件不存在
	exists, err = storage.FileExists(ctx, testBucket, objectName)
	assert.NoError(t, err)
	assert.False(t, exists, "删除后文件应该不存在")
}

// TestMinIOStorage_ListFiles 测试文件列表（需要真实服务）
func TestMinIOStorage_ListFiles(t *testing.T) {
	if !isMinIOAvailable() {
		t.Skip("跳过测试：MinIO服务不可用")
	}

	storage := setupTestStorage(t)
	ctx := context.Background()
	testBucket := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storage.CreateBucket(ctx, testBucket)
	require.NoError(t, err)
	defer func() {
		// 删除所有文件再删除存储桶
		files, _ := storage.ListFiles(ctx, testBucket, "")
		for _, file := range files {
			_ = storage.DeleteFile(ctx, testBucket, file.Key)
		}
		_ = storage.RemoveBucket(ctx, testBucket)
	}()

	// 上传多个测试文件
	testFiles := []struct {
		name string
		data []byte
	}{
		{"videos/2025/08/video1.mp4", []byte("video1 content")},
		{"videos/2025/08/video2.mp4", []byte("video2 content")},
		{"videos/2025/07/video3.mp4", []byte("video3 content")},
	}

	for _, file := range testFiles {
		_, err := storage.UploadFile(ctx, testBucket, file.name, file.data, "video/mp4")
		require.NoError(t, err)
		
		// 验证文件确实上传成功
		exists, err := storage.FileExists(ctx, testBucket, file.name)
		require.NoError(t, err)
		require.True(t, exists, "文件应该存在: "+file.name)
	}

	// 测试列出所有文件
	files, err := storage.ListFiles(ctx, testBucket, "")
	assert.NoError(t, err, "列出文件应该成功")
	assert.Len(t, files, 3, "应该有3个文件")

	// 测试按前缀过滤
	files, err = storage.ListFiles(ctx, testBucket, "videos/2025/08/")
	assert.NoError(t, err)
	assert.Len(t, files, 2, "2025年8月应该有2个文件")
}

// TestMinIOStorage_FileExists_NotFound 测试文件不存在的情况
func TestMinIOStorage_FileExists_NotFound(t *testing.T) {
	if !isMinIOAvailable() {
		t.Skip("跳过测试：MinIO服务不可用")
	}

	storage := setupTestStorage(t)
	ctx := context.Background()
	testBucket := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storage.CreateBucket(ctx, testBucket)
	require.NoError(t, err)
	defer func() {
		_ = storage.RemoveBucket(ctx, testBucket)
	}()

	// 测试不存在的文件
	exists, err := storage.FileExists(ctx, testBucket, "non-existent-file.mp4")
	assert.NoError(t, err, "检查不存在文件应该成功（无错误）")
	assert.False(t, exists, "不存在的文件应该返回false")
}

// isMinIOAvailable 检查MinIO服务是否可用
func isMinIOAvailable() bool {
	// 尝试创建一个存储实例并测试连接
	testConfig := testconfig.GetMinIOTestConfig()
	config := &MinIOConfig{
		Endpoint:  testConfig.GetEndpoint(),
		AccessKey: testConfig.AccessKey,
		SecretKey: testConfig.SecretKey,
		UseSSL:    testConfig.UseSSL,
		Region:    testConfig.Region,
	}

	storage, err := NewMinIOStorage(config)
	if err != nil {
		return false
	}

	ctx := context.Background()
	err = storage.TestConnection(ctx)
	return err == nil
}

// setupTestStorage 设置测试存储实例
func setupTestStorage(t *testing.T) *MinIOStorage {
	testConfig := testconfig.GetMinIOTestConfig()
	config := &MinIOConfig{
		Endpoint:  testConfig.GetEndpoint(),
		AccessKey: testConfig.AccessKey,
		SecretKey: testConfig.SecretKey,
		UseSSL:    testConfig.UseSSL,
		Region:    testConfig.Region,
	}

	storage, err := NewMinIOStorage(config)
	require.NoError(t, err, "创建测试存储实例应该成功")

	return storage
}

// generateTestID 生成测试ID
func generateTestID() string {
	return strings.ReplaceAll(time.Now().Format("20060102-150405.000"), ".", "")
}
