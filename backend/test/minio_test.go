package test

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/manteia/zhulong/pkg/storage"
)

// TestMinIOClientInitialization 测试MinIO客户端初始化
func TestMinIOClientInitialization(t *testing.T) {
	// 测试正常初始化
	client, err := storage.NewMinIOClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

// TestMinIOConnection 测试MinIO连接
func TestMinIOConnection(t *testing.T) {
	client, err := storage.NewMinIOClient()
	assert.Nil(t, err)
	
	// 测试连接是否正常
	err = client.TestConnection()
	assert.Nil(t, err)
}

// TestMinIOBucketOperations 测试MinIO存储桶操作
func TestMinIOBucketOperations(t *testing.T) {
	client, err := storage.NewMinIOClient()
	assert.Nil(t, err)
	
	bucketName := "zhulong-videos"
	
	// 测试存储桶是否存在
	exists, err := client.BucketExists(bucketName)
	assert.Nil(t, err)
	
	// 如果不存在则创建
	if !exists {
		err = client.CreateBucket(bucketName)
		assert.Nil(t, err)
	}
	
	// 验证存储桶现在存在
	exists, err = client.BucketExists(bucketName)
	assert.Nil(t, err)
	assert.True(t, exists)
}

// TestMinIOConfiguration 测试MinIO配置加载
func TestMinIOConfiguration(t *testing.T) {
	config := storage.GetMinIOConfig()
	
	// 验证配置项不为空
	assert.NotNil(t, config.Host)
	assert.NotNil(t, config.AccessKey)
	assert.NotNil(t, config.SecretKey)
	assert.NotNil(t, config.Bucket)
	assert.True(t, config.Port > 0)
}