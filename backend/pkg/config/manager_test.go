package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManager_Load 测试配置管理器加载
func TestManager_Load(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yml")
	
	yamlContent := `
server:
  host: "0.0.0.0"
  port: 8888
  
minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
`
	
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	manager := NewManager(configFile)
	
	// 测试加载配置
	err = manager.Load()
	assert.NoError(t, err, "加载配置应该成功")
	
	// 验证配置
	config := manager.GetConfig()
	assert.NotNil(t, config, "配置不应为空")
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8888, config.Server.Port)
}

// TestManager_LoadInvalidConfig 测试加载无效配置
func TestManager_LoadInvalidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid.yml")
	
	invalidYAML := `
server:
  port: 0  # 无效端口
  
minio:
  endpoint: ""  # 空端点
  access_key: ""  # 空访问密钥
`
	
	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)
	
	manager := NewManager(configFile)
	
	// 测试加载无效配置
	err = manager.Load()
	assert.Error(t, err, "加载无效配置应该失败")
	assert.Contains(t, err.Error(), "配置验证失败")
}

// TestManager_GetConfigs 测试获取各种配置
func TestManager_GetConfigs(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yml")
	
	yamlContent := `
server:
  host: "localhost"
  port: 9000
  
minio:
  endpoint: "minio.local:9000"
  access_key: "testkey"
  secret_key: "testsecret"
  use_ssl: true
  region: "test-region"
  
app:
  name: "Test App"
  version: "v1.2.3"
  debug: true
`
	
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	manager := NewManager(configFile)
	err = manager.Load()
	require.NoError(t, err)
	
	// 测试获取服务器配置
	serverConfig := manager.GetServerConfig()
	assert.Equal(t, "localhost", serverConfig.Host)
	assert.Equal(t, 9000, serverConfig.Port)
	
	// 测试获取应用配置
	appConfig := manager.GetAppConfig()
	assert.Equal(t, "Test App", appConfig.Name)
	assert.Equal(t, "v1.2.3", appConfig.Version)
	assert.True(t, appConfig.Debug)
	
	// 测试获取存储配置
	storageConfig := manager.GetStorageConfig()
	assert.NotNil(t, storageConfig)
	assert.Equal(t, "minio.local:9000", storageConfig.GetEndpoint())
	assert.True(t, storageConfig.IsSSLEnabled())
}

// TestManager_Reload 测试重新加载配置
func TestManager_Reload(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "reload_test.yml")
	
	initialYAML := `
server:
  port: 8000
minio:
  endpoint: "localhost:9000"  
  access_key: "minioadmin"
  secret_key: "minioadmin"
app:
  debug: false
`
	
	err := os.WriteFile(configFile, []byte(initialYAML), 0644)
	require.NoError(t, err)
	
	manager := NewManager(configFile)
	err = manager.Load()
	require.NoError(t, err)
	
	// 验证初始配置
	assert.Equal(t, 8000, manager.GetServerConfig().Port)
	assert.False(t, manager.GetAppConfig().Debug)
	
	// 修改配置文件
	updatedYAML := `
server:
  port: 9000
minio:
  endpoint: "localhost:9000"  
  access_key: "minioadmin"
  secret_key: "minioadmin"
app:
  debug: true
`
	
	err = os.WriteFile(configFile, []byte(updatedYAML), 0644)
	require.NoError(t, err)
	
	// 重新加载配置
	err = manager.Reload()
	assert.NoError(t, err, "重新加载应该成功")
	
	// 验证配置已更新
	assert.Equal(t, 9000, manager.GetServerConfig().Port)
	assert.True(t, manager.GetAppConfig().Debug)
}

// TestManager_Watching 测试配置监听
func TestManager_Watching(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "watch_test.yml")
	
	initialYAML := `
server:
  port: 8000
minio:
  endpoint: "localhost:9000"  
  access_key: "minioadmin"
  secret_key: "minioadmin"
app:
  debug: false
`
	
	err := os.WriteFile(configFile, []byte(initialYAML), 0644)
	require.NoError(t, err)
	
	manager := NewManager(configFile)
	err = manager.Load()
	require.NoError(t, err)
	
	// 启动监听
	changes := make(chan *Config, 1)
	err = manager.StartWatching(changes)
	assert.NoError(t, err, "启动监听应该成功")
	
	defer func() {
		err := manager.StopWatching()
		assert.NoError(t, err)
	}()
	
	// 修改配置文件
	updatedYAML := `
server:
  port: 9000
minio:
  endpoint: "localhost:9000"  
  access_key: "minioadmin"
  secret_key: "minioadmin"
app:
  debug: true
`
	
	err = os.WriteFile(configFile, []byte(updatedYAML), 0644)
	require.NoError(t, err)
	
	// 等待配置变更通知
	select {
	case newConfig := <-changes:
		assert.NotNil(t, newConfig)
		assert.Equal(t, 9000, newConfig.Server.Port)
		assert.True(t, newConfig.App.Debug)
	case <-time.After(2 * time.Second):
		t.Fatal("超时：未接收到配置变更通知")
	}
}

// TestManager_UnloadedConfig 测试未加载配置时的行为
func TestManager_UnloadedConfig(t *testing.T) {
	manager := NewManager("non-existent.yml")
	
	// 测试获取配置（未加载）
	config := manager.GetConfig()
	assert.Nil(t, config, "未加载时配置应为空")
	
	// 测试验证（未加载）
	err := manager.Validate()
	assert.Error(t, err, "未加载配置时验证应该失败")
	assert.Contains(t, err.Error(), "配置未加载")
	
	// 测试获取存储配置（未加载）
	storageConfig := manager.GetStorageConfig()
	assert.Nil(t, storageConfig, "未加载时存储配置应为空")
	
	// 测试获取服务器配置（未加载）
	serverConfig := manager.GetServerConfig()
	assert.Equal(t, ServerConfig{}, serverConfig, "未加载时应返回零值")
	
	// 测试获取应用配置（未加载）
	appConfig := manager.GetAppConfig()
	assert.Equal(t, AppConfig{}, appConfig, "未加载时应返回零值")
}