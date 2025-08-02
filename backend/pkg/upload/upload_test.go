package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/manteia/zhulong/pkg/storage"
	"github.com/manteia/zhulong/testconfig"
)

// TestUploadService_SingleFileUpload 测试单文件上传
func TestUploadService_SingleFileUpload(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	uploadService := NewUploadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 准备测试文件
	testData := []byte("这是一个测试视频文件内容，用于测试单文件上传功能")
	fileName := "test-video.mp4"
	contentType := "video/mp4"

	uploadRequest := &UploadRequest{
		FileName:    fileName,
		ContentType: contentType,
		Size:        int64(len(testData)),
		Reader:      bytes.NewReader(testData),
		BucketName:  bucketName,
	}

	// 执行上传
	result, err := uploadService.UploadFile(ctx, uploadRequest)
	assert.NoError(t, err, "单文件上传应该成功")
	require.NotNil(t, result, "上传结果不应为空")

	// 验证上传结果
	assert.NotEmpty(t, result.FileID, "文件ID不应为空")
	assert.NotEmpty(t, result.ObjectName, "对象名不应为空")
	assert.Equal(t, int64(len(testData)), result.Size, "文件大小应该匹配")
	assert.NotEmpty(t, result.ETag, "ETag不应为空")
	assert.True(t, strings.Contains(result.ObjectName, fileName), "对象名应该包含文件名")

	// 验证文件确实已上传
	exists, err := storageService.FileExists(ctx, bucketName, result.ObjectName)
	assert.NoError(t, err)
	assert.True(t, exists, "上传的文件应该存在")
}

// TestUploadService_MultipartUpload 测试分片上传
func TestUploadService_MultipartUpload(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	uploadService := NewUploadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 创建大文件（10MB）用于测试分片上传
	fileSize := int64(10 * 1024 * 1024) // 10MB
	fileName := "large-video.mp4"
	contentType := "video/mp4"

	// 创建随机数据
	testData := make([]byte, fileSize)
	_, err = rand.Read(testData)
	require.NoError(t, err)

	multipartRequest := &MultipartUploadRequest{
		FileName:    fileName,
		ContentType: contentType,
		TotalSize:   fileSize,
		BucketName:  bucketName,
		ChunkSize:   1024 * 1024, // 1MB 分片
	}

	// 初始化分片上传
	session, err := uploadService.InitMultipartUpload(ctx, multipartRequest)
	assert.NoError(t, err, "初始化分片上传应该成功")
	require.NotNil(t, session, "上传会话不应为空")
	assert.NotEmpty(t, session.UploadID, "上传ID不应为空")
	assert.NotEmpty(t, session.ObjectName, "对象名不应为空")

	// 分片上传
	var parts []CompletedPart
	chunkSize := int64(1024 * 1024) // 1MB

	for i := int64(0); i < fileSize; i += chunkSize {
		end := i + chunkSize
		if end > fileSize {
			end = fileSize
		}

		partNumber := int(i/chunkSize) + 1
		chunkData := testData[i:end]

		partRequest := &UploadPartRequest{
			UploadID:   session.UploadID,
			ObjectName: session.ObjectName,
			PartNumber: partNumber,
			Data:       chunkData,
			BucketName: bucketName,
		}

		partResult, err := uploadService.UploadPart(ctx, partRequest)
		assert.NoError(t, err, "分片上传应该成功")
		require.NotNil(t, partResult, "分片结果不应为空")

		parts = append(parts, CompletedPart{
			PartNumber: partNumber,
			ETag:       partResult.ETag,
		})
	}

	// 完成分片上传
	completeRequest := &CompleteMultipartRequest{
		UploadID:   session.UploadID,
		ObjectName: session.ObjectName,
		Parts:      parts,
		BucketName: bucketName,
	}

	result, err := uploadService.CompleteMultipartUpload(ctx, completeRequest)
	assert.NoError(t, err, "完成分片上传应该成功")
	require.NotNil(t, result, "完成结果不应为空")

	// 验证上传结果
	assert.NotEmpty(t, result.FileID, "文件ID不应为空")
	assert.Equal(t, session.ObjectName, result.ObjectName, "对象名应该匹配")
	assert.Equal(t, fileSize, result.Size, "文件大小应该匹配")

	// 验证文件确实已上传
	exists, err := storageService.FileExists(ctx, bucketName, result.ObjectName)
	assert.NoError(t, err)
	assert.True(t, exists, "上传的文件应该存在")

	// 验证文件大小
	fileInfo, err := storageService.GetFileInfo(ctx, bucketName, result.ObjectName)
	assert.NoError(t, err)
	assert.Equal(t, fileSize, fileInfo.Size, "存储的文件大小应该匹配")
}

// TestUploadService_AbortMultipartUpload 测试中止分片上传
func TestUploadService_AbortMultipartUpload(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	uploadService := NewUploadService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	multipartRequest := &MultipartUploadRequest{
		FileName:    "abort-test.mp4",
		ContentType: "video/mp4",
		TotalSize:   5 * 1024 * 1024, // 5MB
		BucketName:  bucketName,
		ChunkSize:   1024 * 1024, // 1MB
	}

	// 初始化分片上传
	session, err := uploadService.InitMultipartUpload(ctx, multipartRequest)
	require.NoError(t, err)

	// 中止分片上传
	abortRequest := &AbortMultipartRequest{
		UploadID:   session.UploadID,
		ObjectName: session.ObjectName,
		BucketName: bucketName,
	}

	err = uploadService.AbortMultipartUpload(ctx, abortRequest)
	assert.NoError(t, err, "中止分片上传应该成功")

	// 验证文件不存在
	exists, err := storageService.FileExists(ctx, bucketName, session.ObjectName)
	assert.NoError(t, err)
	assert.False(t, exists, "中止后文件不应存在")
}

// TestUploadService_GenerateObjectName 测试对象名生成
func TestUploadService_GenerateObjectName(t *testing.T) {
	uploadService := NewUploadService(nil)

	testCases := []struct {
		fileName     string
		expectedPath string
	}{
		{"test.mp4", "videos"},
		{"movie.avi", "videos"},
		{"video.mov", "videos"},
		{"document.pdf", "videos"}, // 即使不是视频文件，也应该放在videos目录
	}

	for _, tc := range testCases {
		t.Run(tc.fileName, func(t *testing.T) {
			objectName := uploadService.GenerateObjectName(tc.fileName)

			assert.NotEmpty(t, objectName, "对象名不应为空")
			assert.True(t, strings.HasPrefix(objectName, tc.expectedPath), "对象名应该以正确路径开头")
			assert.True(t, strings.Contains(objectName, tc.fileName), "对象名应该包含原文件名")

			// 验证路径格式：videos/{year}/{month}/{uuid}-{filename}
			parts := strings.Split(objectName, "/")
			assert.Len(t, parts, 4, "对象名应该有4个路径部分")
			assert.Equal(t, "videos", parts[0], "第一部分应该是videos")
			assert.Len(t, parts[1], 4, "年份应该是4位数")
			assert.Len(t, parts[2], 2, "月份应该是2位数")
			assert.True(t, strings.Contains(parts[3], tc.fileName), "文件名部分应该包含原文件名")
		})
	}
}

// TestUploadService_ValidateRequest 测试请求验证
func TestUploadService_ValidateRequest(t *testing.T) {
	uploadService := NewUploadService(nil)

	// 测试有效请求
	validRequest := &UploadRequest{
		FileName:    "test.mp4",
		ContentType: "video/mp4",
		Size:        1024,
		Reader:      strings.NewReader("test data"),
		BucketName:  "test-bucket",
	}

	err := uploadService.ValidateUploadRequest(validRequest)
	assert.NoError(t, err, "有效请求应该通过验证")

	// 测试无效请求
	invalidCases := []struct {
		name    string
		request *UploadRequest
		errMsg  string
	}{
		{
			name: "空文件名",
			request: &UploadRequest{
				FileName:    "",
				ContentType: "video/mp4",
				Size:        1024,
				Reader:      strings.NewReader("test"),
				BucketName:  "test-bucket",
			},
			errMsg: "文件名不能为空",
		},
		{
			name: "空内容类型",
			request: &UploadRequest{
				FileName:    "test.mp4",
				ContentType: "",
				Size:        1024,
				Reader:      strings.NewReader("test"),
				BucketName:  "test-bucket",
			},
			errMsg: "内容类型不能为空",
		},
		{
			name: "无效文件大小",
			request: &UploadRequest{
				FileName:    "test.mp4",
				ContentType: "video/mp4",
				Size:        0,
				Reader:      strings.NewReader("test"),
				BucketName:  "test-bucket",
			},
			errMsg: "文件大小必须大于0",
		},
		{
			name: "文件过大",
			request: &UploadRequest{
				FileName:    "test.mp4",
				ContentType: "video/mp4",
				Size:        3 * 1024 * 1024 * 1024, // 3GB
				Reader:      strings.NewReader("test"),
				BucketName:  "test-bucket",
			},
			errMsg: "文件大小超过限制",
		},
		{
			name: "空Reader",
			request: &UploadRequest{
				FileName:    "test.mp4",
				ContentType: "video/mp4",
				Size:        1024,
				Reader:      nil,
				BucketName:  "test-bucket",
			},
			errMsg: "文件读取器不能为空",
		},
		{
			name: "空存储桶名",
			request: &UploadRequest{
				FileName:    "test.mp4",
				ContentType: "video/mp4",
				Size:        1024,
				Reader:      strings.NewReader("test"),
				BucketName:  "",
			},
			errMsg: "存储桶名不能为空",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := uploadService.ValidateUploadRequest(tc.request)
			assert.Error(t, err, "无效请求应该验证失败")
			assert.Contains(t, err.Error(), tc.errMsg, "错误信息应该包含预期内容")
		})
	}
}

// TestUploadService_UploadProgress 测试上传进度跟踪
func TestUploadService_UploadProgress(t *testing.T) {
	uploadService := NewUploadService(nil)

	// 创建进度跟踪器
	progressCh := make(chan *UploadProgress, 10)
	tracker := uploadService.CreateProgressTracker("test-upload-id", progressCh)

	// 模拟上传进度
	go func() {
		defer close(progressCh) // 由调用方关闭通道
		tracker.UpdateProgress(25)
		time.Sleep(10 * time.Millisecond)
		tracker.UpdateProgress(50)
		time.Sleep(10 * time.Millisecond)
		tracker.UpdateProgress(75)
		time.Sleep(10 * time.Millisecond)
		tracker.UpdateProgress(100)
		tracker.Complete()
	}()

	// 验证进度更新
	var progressUpdates []*UploadProgress
	for progress := range progressCh {
		progressUpdates = append(progressUpdates, progress)
		if progress.IsCompleted {
			break
		}
	}

	assert.Len(t, progressUpdates, 5, "应该收到5个进度更新")
	assert.Equal(t, 25, progressUpdates[0].Percentage, "第一个进度应该是25%")
	assert.Equal(t, 50, progressUpdates[1].Percentage, "第二个进度应该是50%")
	assert.Equal(t, 75, progressUpdates[2].Percentage, "第三个进度应该是75%")
	assert.Equal(t, 100, progressUpdates[3].Percentage, "第四个进度应该是100%")
	assert.Equal(t, 100, progressUpdates[4].Percentage, "最后进度应该是100%")
	assert.True(t, progressUpdates[4].IsCompleted, "最后应该标记为完成")
}

// isStorageAvailable 检查存储服务是否可用
func isStorageAvailable() bool {
	testConfig := testconfig.GetMinIOTestConfig()
	storageConfig := &storage.MinIOConfig{
		Endpoint:  testConfig.GetEndpoint(),
		AccessKey: testConfig.AccessKey,
		SecretKey: testConfig.SecretKey,
		UseSSL:    testConfig.UseSSL,
		Region:    testConfig.Region,
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
	testConfig := testconfig.GetMinIOTestConfig()
	storageConfig := &storage.MinIOConfig{
		Endpoint:  testConfig.GetEndpoint(),
		AccessKey: testConfig.AccessKey,
		SecretKey: testConfig.SecretKey,
		UseSSL:    testConfig.UseSSL,
		Region:    testConfig.Region,
	}

	storageService, err := storage.NewMinIOStorage(storageConfig)
	require.NoError(t, err)

	return storageService
}

// generateTestID 生成测试ID
func generateTestID() string {
	return strings.ReplaceAll(time.Now().Format("20060102-150405.000"), ".", "")
}
