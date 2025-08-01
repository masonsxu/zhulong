package test

import (
	"os"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/manteia/zhulong/pkg/config"
)

// TestConfigDirectoryStructure 测试配置目录结构
func TestConfigDirectoryStructure(t *testing.T) {
	// 测试配置目录是否存在
	configDir := "../../config"
	stat, err := os.Stat(configDir)
	assert.Nil(t, err)
	assert.True(t, stat.IsDir())
	
	// 测试配置文件是否存在
	configFiles := []string{
		"../../config/README.md",
		"../../config/app.yml",
		"../../config/development.yml",
		"../../config/production.yml",
	}
	
	for _, file := range configFiles {
		_, err := os.Stat(file)
		assert.Nil(t, err)
	}
}

// TestEnvFileDiscovery 测试.env文件发现机制
func TestEnvFileDiscovery(t *testing.T) {
	// 重置全局配置以测试文件发现
	cfg := config.LoadConfig()
	assert.NotNil(t, cfg)
	
	// 验证配置值不是默认值（说明成功加载了.env文件）
	assert.NotEqual(t, "localhost", cfg.MinIO.Host)
	assert.NotEqual(t, "minioadmin", cfg.MinIO.AccessKey)
}