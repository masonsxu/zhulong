package delete

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/manteia/zhulong/pkg/storage"
	"github.com/manteia/zhulong/testconfig"
)

// TestDeleteService_DeleteFile 测试单文件删除
func TestDeleteService_DeleteFile(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	deleteService := NewDeleteService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()
	objectName := "test-video.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传测试文件
	testData := []byte("这是一个测试视频文件内容，用于测试文件删除功能")
	_, err = storageService.UploadFile(ctx, bucketName, objectName, testData, "video/mp4")
	require.NoError(t, err)

	// 验证文件存在
	exists, err := storageService.FileExists(ctx, bucketName, objectName)
	require.NoError(t, err)
	require.True(t, exists, "文件应该存在")

	// 测试删除文件
	deleteRequest := &DeleteRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}

	result, err := deleteService.DeleteFile(ctx, deleteRequest)
	assert.NoError(t, err, "文件删除应该成功")
	require.NotNil(t, result, "删除结果不应为空")

	// 验证删除结果
	assert.Equal(t, bucketName, result.BucketName, "存储桶名应该匹配")
	assert.Equal(t, objectName, result.ObjectName, "对象名应该匹配")
	assert.True(t, result.Success, "删除应该成功")
	assert.Empty(t, result.ErrorMessage, "成功时错误信息应为空")

	// 验证文件已被删除
	exists, err = storageService.FileExists(ctx, bucketName, objectName)
	assert.NoError(t, err)
	assert.False(t, exists, "删除后文件不应存在")
}

// TestDeleteService_DeleteMultipleFiles 测试批量文件删除
func TestDeleteService_DeleteMultipleFiles(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	deleteService := NewDeleteService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传多个测试文件
	testFiles := []struct {
		name string
		data []byte
	}{
		{"video1.mp4", []byte("video1 content")},
		{"video2.mp4", []byte("video2 content")},
		{"video3.mp4", []byte("video3 content")},
	}

	for _, file := range testFiles {
		_, err := storageService.UploadFile(ctx, bucketName, file.name, file.data, "video/mp4")
		require.NoError(t, err)
	}

	// 验证所有文件都存在
	for _, file := range testFiles {
		exists, err := storageService.FileExists(ctx, bucketName, file.name)
		require.NoError(t, err)
		require.True(t, exists, "文件 %s 应该存在", file.name)
	}

	// 测试批量删除
	objectNames := []string{"video1.mp4", "video2.mp4", "video3.mp4"}
	batchDeleteRequest := &BatchDeleteRequest{
		BucketName:  bucketName,
		ObjectNames: objectNames,
	}

	results, err := deleteService.DeleteMultipleFiles(ctx, batchDeleteRequest)
	assert.NoError(t, err, "批量删除应该成功")
	require.NotNil(t, results, "删除结果不应为空")
	assert.Len(t, results.Results, 3, "应该有3个删除结果")

	// 验证每个删除结果
	for i, result := range results.Results {
		assert.Equal(t, bucketName, result.BucketName, "存储桶名应该匹配")
		assert.Equal(t, objectNames[i], result.ObjectName, "对象名应该匹配")
		assert.True(t, result.Success, "删除应该成功")
		assert.Empty(t, result.ErrorMessage, "成功时错误信息应为空")
	}

	// 验证所有文件都已被删除
	for _, file := range testFiles {
		exists, err := storageService.FileExists(ctx, bucketName, file.name)
		assert.NoError(t, err)
		assert.False(t, exists, "删除后文件 %s 不应存在", file.name)
	}

	// 验证批量删除结果统计
	assert.Equal(t, 3, results.TotalCount, "总数应该是3")
	assert.Equal(t, 3, results.SuccessCount, "成功数应该是3")
	assert.Equal(t, 0, results.FailureCount, "失败数应该是0")
}

// TestDeleteService_DeleteFile_NotFound 测试删除不存在的文件
func TestDeleteService_DeleteFile_NotFound(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	deleteService := NewDeleteService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()
	objectName := "non-existent-file.mp4"

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 测试删除不存在的文件
	deleteRequest := &DeleteRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}

	result, err := deleteService.DeleteFile(ctx, deleteRequest)
	assert.Error(t, err, "删除不存在的文件应该失败")
	assert.Nil(t, result, "失败时结果应为空")
	assert.Contains(t, err.Error(), "文件不存在", "错误信息应该表明文件不存在")
}

// TestDeleteService_DeleteMultipleFiles_PartialFailure 测试批量删除部分失败
func TestDeleteService_DeleteMultipleFiles_PartialFailure(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	deleteService := NewDeleteService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 只上传部分文件
	testData := []byte("test video content")
	_, err = storageService.UploadFile(ctx, bucketName, "existing-video.mp4", testData, "video/mp4")
	require.NoError(t, err)

	// 测试批量删除（包含存在和不存在的文件）
	objectNames := []string{"existing-video.mp4", "non-existent1.mp4", "non-existent2.mp4"}
	batchDeleteRequest := &BatchDeleteRequest{
		BucketName:  bucketName,
		ObjectNames: objectNames,
	}

	results, err := deleteService.DeleteMultipleFiles(ctx, batchDeleteRequest)
	assert.NoError(t, err, "批量删除应该执行成功（即使部分失败）")
	require.NotNil(t, results, "删除结果不应为空")
	assert.Len(t, results.Results, 3, "应该有3个删除结果")

	// 验证删除结果
	assert.True(t, results.Results[0].Success, "存在的文件应该删除成功")
	assert.False(t, results.Results[1].Success, "不存在的文件应该删除失败")
	assert.False(t, results.Results[2].Success, "不存在的文件应该删除失败")

	// 验证失败的文件有错误信息
	assert.NotEmpty(t, results.Results[1].ErrorMessage, "失败时应该有错误信息")
	assert.NotEmpty(t, results.Results[2].ErrorMessage, "失败时应该有错误信息")

	// 验证批量删除结果统计
	assert.Equal(t, 3, results.TotalCount, "总数应该是3")
	assert.Equal(t, 1, results.SuccessCount, "成功数应该是1")
	assert.Equal(t, 2, results.FailureCount, "失败数应该是2")
}

// TestDeleteService_ValidateDeleteRequest 测试删除请求验证
func TestDeleteService_ValidateDeleteRequest(t *testing.T) {
	deleteService := NewDeleteService(nil)

	// 测试有效请求
	validRequest := &DeleteRequest{
		BucketName: "test-bucket",
		ObjectName: "test-video.mp4",
	}

	err := deleteService.ValidateDeleteRequest(validRequest)
	assert.NoError(t, err, "有效请求应该通过验证")

	// 测试无效请求
	invalidCases := []struct {
		name    string
		request *DeleteRequest
		errMsg  string
	}{
		{
			name: "空存储桶名",
			request: &DeleteRequest{
				BucketName: "",
				ObjectName: "test-video.mp4",
			},
			errMsg: "存储桶名不能为空",
		},
		{
			name: "空对象名",
			request: &DeleteRequest{
				BucketName: "test-bucket",
				ObjectName: "",
			},
			errMsg: "对象名不能为空",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := deleteService.ValidateDeleteRequest(tc.request)
			assert.Error(t, err, "无效请求应该验证失败")
			assert.Contains(t, err.Error(), tc.errMsg, "错误信息应该包含预期内容")
		})
	}
}

// TestDeleteService_ValidateBatchDeleteRequest 测试批量删除请求验证
func TestDeleteService_ValidateBatchDeleteRequest(t *testing.T) {
	deleteService := NewDeleteService(nil)

	// 测试有效请求
	validRequest := &BatchDeleteRequest{
		BucketName:  "test-bucket",
		ObjectNames: []string{"video1.mp4", "video2.mp4"},
	}

	err := deleteService.ValidateBatchDeleteRequest(validRequest)
	assert.NoError(t, err, "有效请求应该通过验证")

	// 测试无效请求
	invalidCases := []struct {
		name    string
		request *BatchDeleteRequest
		errMsg  string
	}{
		{
			name: "空存储桶名",
			request: &BatchDeleteRequest{
				BucketName:  "",
				ObjectNames: []string{"video1.mp4"},
			},
			errMsg: "存储桶名不能为空",
		},
		{
			name: "空对象名列表",
			request: &BatchDeleteRequest{
				BucketName:  "test-bucket",
				ObjectNames: []string{},
			},
			errMsg: "对象名列表不能为空",
		},
		{
			name: "对象名列表过长",
			request: &BatchDeleteRequest{
				BucketName:  "test-bucket",
				ObjectNames: make([]string, 1001), // 超过限制
			},
			errMsg: "一次最多删除1000个文件",
		},
		{
			name: "包含空对象名",
			request: &BatchDeleteRequest{
				BucketName:  "test-bucket",
				ObjectNames: []string{"video1.mp4", "", "video3.mp4"},
			},
			errMsg: "对象名不能为空",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := deleteService.ValidateBatchDeleteRequest(tc.request)
			assert.Error(t, err, "无效请求应该验证失败")
			assert.Contains(t, err.Error(), tc.errMsg, "错误信息应该包含预期内容")
		})
	}
}

// TestDeleteService_DeleteFilesByPrefix 测试按前缀删除文件
func TestDeleteService_DeleteFilesByPrefix(t *testing.T) {
	if !isStorageAvailable() {
		t.Skip("跳过测试：MinIO存储服务不可用")
	}

	storageService := setupTestStorage(t)
	deleteService := NewDeleteService(storageService)

	ctx := context.Background()
	bucketName := "test-bucket-" + generateTestID()

	// 创建测试存储桶
	err := storageService.CreateBucket(ctx, bucketName)
	require.NoError(t, err)
	defer func() {
		_ = storageService.RemoveBucket(ctx, bucketName)
	}()

	// 上传测试文件
	testFiles := []struct {
		name string
		data []byte
	}{
		{"videos/2025/08/video1.mp4", []byte("video1")},
		{"videos/2025/08/video2.mp4", []byte("video2")},
		{"videos/2025/07/video3.mp4", []byte("video3")},
		{"documents/doc1.pdf", []byte("doc1")},
	}

	for _, file := range testFiles {
		_, err := storageService.UploadFile(ctx, bucketName, file.name, file.data, "application/octet-stream")
		require.NoError(t, err)
	}

	// 测试按前缀删除
	prefixDeleteRequest := &PrefixDeleteRequest{
		BucketName: bucketName,
		Prefix:     "videos/2025/08/",
	}

	result, err := deleteService.DeleteFilesByPrefix(ctx, prefixDeleteRequest)
	assert.NoError(t, err, "按前缀删除应该成功")
	require.NotNil(t, result, "删除结果不应为空")

	// 验证删除结果
	assert.Equal(t, 2, result.DeletedCount, "应该删除2个文件")
	assert.Len(t, result.DeletedFiles, 2, "删除的文件列表应该有2个")

	// 验证被删除的文件
	deletedFiles := result.DeletedFiles
	assert.Contains(t, deletedFiles, "videos/2025/08/video1.mp4", "应该包含video1.mp4")
	assert.Contains(t, deletedFiles, "videos/2025/08/video2.mp4", "应该包含video2.mp4")

	// 验证匹配前缀的文件已被删除
	exists1, _ := storageService.FileExists(ctx, bucketName, "videos/2025/08/video1.mp4")
	exists2, _ := storageService.FileExists(ctx, bucketName, "videos/2025/08/video2.mp4")
	assert.False(t, exists1, "video1.mp4应该被删除")
	assert.False(t, exists2, "video2.mp4应该被删除")

	// 验证不匹配前缀的文件仍然存在
	exists3, _ := storageService.FileExists(ctx, bucketName, "videos/2025/07/video3.mp4")
	exists4, _ := storageService.FileExists(ctx, bucketName, "documents/doc1.pdf")
	assert.True(t, exists3, "video3.mp4应该仍然存在")
	assert.True(t, exists4, "doc1.pdf应该仍然存在")
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
