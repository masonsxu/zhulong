package test

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/manteia/zhulong/pkg/config"
)

// TestYAMLConfigLoading 测试YAML配置文件加载
func TestYAMLConfigLoading(t *testing.T) {
	// 测试加载开发环境配置
	cfg, err := config.LoadFromYAML("development")
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	
	// 验证服务器配置
	assert.DeepEqual(t, "localhost", cfg.Server.Host)
	assert.DeepEqual(t, 8080, cfg.Server.Port)
	
	// 验证MinIO配置
	assert.DeepEqual(t, "localhost:9000", cfg.MinIO.Endpoint)
	assert.DeepEqual(t, "zhulong-videos-dev", cfg.MinIO.Bucket)
	assert.False(t, cfg.MinIO.UseSSL)
}

// TestProductionConfig 测试生产环境配置
func TestProductionConfig(t *testing.T) {
	cfg, err := config.LoadFromYAML("production")
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	
	// 验证生产环境配置
	assert.DeepEqual(t, "0.0.0.0", cfg.Server.Host)
	assert.DeepEqual(t, 8080, cfg.Server.Port)
	assert.True(t, cfg.MinIO.UseSSL)
	assert.DeepEqual(t, "24h", cfg.JWT.Expire)
}

// TestMinIOConfigFile 测试MinIO专用配置文件
func TestMinIOConfigFile(t *testing.T) {
	minioConfig, err := config.LoadMinIOConfig()
	assert.Nil(t, err)
	assert.NotNil(t, minioConfig)
	
	// 验证MinIO配置结构
	assert.True(t, len(minioConfig.MinIO.Endpoint) > 0)
	assert.True(t, len(minioConfig.MinIO.AccessKey) > 0)
	assert.True(t, len(minioConfig.MinIO.SecretKey) > 0)
	assert.True(t, len(minioConfig.MinIO.Bucket) > 0)
	
	// 验证存储配置
	assert.True(t, len(minioConfig.Storage.PathPattern) > 0)
	assert.True(t, len(minioConfig.Storage.AllowedFormats) > 0)
}

// TestConfigMerging 测试配置合并（YAML + 环境变量）
func TestConfigMerging(t *testing.T) {
	// 测试环境变量覆盖YAML配置
	cfg := config.GetConfig()
	assert.NotNil(t, cfg)
	
	// 环境变量应该优先于YAML配置
	// 如果.env文件存在，MinIO配置来自环境变量
	if cfg.MinIO.Host != "localhost" {
		// 说明加载了.env文件中的配置
		assert.DeepEqual(t, "172.31.145.138", cfg.MinIO.Host)
		assert.DeepEqual(t, "admin", cfg.MinIO.AccessKey)
	}
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	cfg := config.GetConfig()
	
	// 验证必需配置项不为空
	assert.True(t, len(cfg.Port) > 0)
	assert.True(t, len(cfg.MinIO.Host) > 0)
	assert.True(t, len(cfg.MinIO.AccessKey) > 0)
	assert.True(t, len(cfg.MinIO.SecretKey) > 0)
	assert.True(t, len(cfg.MinIO.Bucket) > 0)
	assert.True(t, cfg.MinIO.Port > 0)
}