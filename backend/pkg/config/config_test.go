package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_LoadFromYAML 测试从YAML文件加载配置
func TestConfig_LoadFromYAML(t *testing.T) {
	// 创建临时配置文件
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
  use_ssl: false
  region: "us-east-1"
  bucket: "zhulong-videos"
  
app:
  name: "Zhulong Video Server"
  version: "v1.0.0"
  debug: true
`
	
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	// 红阶段：测试加载配置（应该失败，因为还没有实现）
	config, err := LoadFromFile(configFile)
	assert.NoError(t, err, "加载配置文件应该成功")
	require.NotNil(t, config, "配置对象不应为空")
	
	// 验证服务器配置
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8888, config.Server.Port)
	
	// 验证MinIO配置
	assert.Equal(t, "localhost:9000", config.MinIO.Endpoint)
	assert.Equal(t, "minioadmin", config.MinIO.AccessKey)
	assert.Equal(t, "minioadmin", config.MinIO.SecretKey)
	assert.False(t, config.MinIO.UseSSL)
	assert.Equal(t, "us-east-1", config.MinIO.Region)
	assert.Equal(t, "zhulong-videos", config.MinIO.Bucket)
	
	// 验证应用配置
	assert.Equal(t, "Zhulong Video Server", config.App.Name)
	assert.Equal(t, "v1.0.0", config.App.Version)
	assert.True(t, config.App.Debug)
}

// TestConfig_LoadFromFile_NotFound 测试加载不存在的配置文件
func TestConfig_LoadFromFile_NotFound(t *testing.T) {
	config, err := LoadFromFile("non-existent-file.yml")
	assert.Error(t, err, "加载不存在的文件应该返回错误")
	assert.Nil(t, config, "配置对象应为空")
	assert.Contains(t, err.Error(), "配置文件不存在", "错误信息应该包含文件不存在的提示")
}

// TestConfig_LoadFromFile_InvalidYAML 测试加载无效的YAML文件
func TestConfig_LoadFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid.yml")
	
	invalidYAML := `
server:
  host: "localhost"
  port: invalid_port
  missing_quote: some value
`
	
	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)
	
	config, err := LoadFromFile(configFile)
	assert.Error(t, err, "加载无效YAML应该返回错误")
	assert.Nil(t, config, "配置对象应为空")
	assert.Contains(t, err.Error(), "解析YAML失败", "错误信息应该包含解析失败的提示")
}

// TestConfig_LoadWithEnvironmentOverride 测试环境变量覆盖
func TestConfig_LoadWithEnvironmentOverride(t *testing.T) {
	// 设置环境变量
	os.Setenv("ZHULONG_SERVER_PORT", "9999")
	os.Setenv("ZHULONG_MINIO_ENDPOINT", "minio.example.com:9000")
	defer func() {
		os.Unsetenv("ZHULONG_SERVER_PORT")
		os.Unsetenv("ZHULONG_MINIO_ENDPOINT")
	}()
	
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test.yml")
	
	yamlContent := `
server:
  host: "localhost"
  port: 8888
  
minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
`
	
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	config, err := LoadFromFile(configFile)
	assert.NoError(t, err)
	require.NotNil(t, config)
	
	// 验证环境变量覆盖了配置文件的值
	assert.Equal(t, 9999, config.Server.Port, "环境变量应该覆盖配置文件中的端口")
	assert.Equal(t, "minio.example.com:9000", config.MinIO.Endpoint, "环境变量应该覆盖MinIO端点")
}

// TestConfig_Validation 测试配置验证
func TestConfig_Validation(t *testing.T) {
	// 测试无效配置
	config := &Config{
		Server: ServerConfig{
			Host: "",
			Port: 0,
		},
		MinIO: MinIOConfig{
			Endpoint:  "",
			AccessKey: "",
			SecretKey: "",
		},
	}
	
	err := config.Validate()
	assert.Error(t, err, "无效配置应该验证失败")
	assert.Contains(t, err.Error(), "服务器端口", "错误信息应该包含端口验证")
	assert.Contains(t, err.Error(), "MinIO端点", "错误信息应该包含MinIO端点验证")
	assert.Contains(t, err.Error(), "访问密钥", "错误信息应该包含访问密钥验证")
}

// TestConfig_DefaultValues 测试默认值
func TestConfig_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "minimal.yml")
	
	// 最小配置文件，测试默认值
	minimalYAML := `
server:
  port: 8888
  
minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
`
	
	err := os.WriteFile(configFile, []byte(minimalYAML), 0644)
	require.NoError(t, err)
	
	config, err := LoadFromFile(configFile)
	assert.NoError(t, err)
	require.NotNil(t, config)
	
	// 验证默认值
	assert.Equal(t, "localhost", config.Server.Host, "应该使用默认主机")
	assert.Equal(t, "us-east-1", config.MinIO.Region, "应该使用默认区域")
	assert.Equal(t, "zhulong-videos", config.MinIO.Bucket, "应该使用默认存储桶")
	assert.False(t, config.MinIO.UseSSL, "应该默认不使用SSL")
}

// TestConfig_GetStorageConfig 测试获取存储配置
func TestConfig_GetStorageConfig(t *testing.T) {
	config := &Config{
		MinIO: MinIOConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSSL:    false,
			Region:    "us-east-1",
			Bucket:    "test-bucket",
		},
	}
	
	storageConfig := config.GetStorageConfig()
	assert.NotNil(t, storageConfig, "存储配置不应为空")
	assert.Equal(t, "localhost:9000", storageConfig.GetEndpoint())
	assert.Equal(t, "minioadmin", storageConfig.GetAccessKey())
	assert.Equal(t, "minioadmin", storageConfig.GetSecretKey())
	assert.False(t, storageConfig.IsSSLEnabled())
	assert.Equal(t, "us-east-1", storageConfig.GetRegion())
}

// TestConfig_LoadEnvironmentSpecific 测试加载环境特定配置
func TestConfig_LoadEnvironmentSpecific(t *testing.T) {
	tempDir := t.TempDir()
	
	// 创建开发环境配置
	devConfig := filepath.Join(tempDir, "development.yml")
	devYAML := `
server:
  port: 8888
  
app:
  debug: true
  
minio:
  endpoint: "localhost:9000"
  access_key: "dev-key"
  secret_key: "dev-secret"
`
	
	err := os.WriteFile(devConfig, []byte(devYAML), 0644)
	require.NoError(t, err)
	
	// 创建生产环境配置
	prodConfig := filepath.Join(tempDir, "production.yml")
	prodYAML := `
server:
  port: 80
  
app:
  debug: false
  
minio:
  endpoint: "minio.prod.com:9000"
  access_key: "prod-key"
  secret_key: "prod-secret"
  use_ssl: true
`
	
	err = os.WriteFile(prodConfig, []byte(prodYAML), 0644)
	require.NoError(t, err)
	
	// 测试加载开发环境配置
	devConf, err := LoadEnvironmentConfig(tempDir, "development")
	assert.NoError(t, err)
	require.NotNil(t, devConf)
	assert.True(t, devConf.App.Debug)
	assert.Equal(t, "dev-key", devConf.MinIO.AccessKey)
	
	// 测试加载生产环境配置  
	prodConf, err := LoadEnvironmentConfig(tempDir, "production")
	assert.NoError(t, err)
	require.NotNil(t, prodConf)
	assert.False(t, prodConf.App.Debug)
	assert.Equal(t, "prod-key", prodConf.MinIO.AccessKey)
	assert.True(t, prodConf.MinIO.UseSSL)
}

// TestConfig_Watch 测试配置文件监听（热重载）
func TestConfig_Watch(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "watch_test.yml")
	
	initialYAML := `
server:
  port: 8888
app:
  debug: false
`
	
	err := os.WriteFile(configFile, []byte(initialYAML), 0644)
	require.NoError(t, err)
	
	// 创建配置监听器
	watcher, err := NewConfigWatcher(configFile)
	assert.NoError(t, err, "创建配置监听器应该成功")
	require.NotNil(t, watcher, "监听器不应为空")
	
	// 启动监听
	changes := make(chan *Config, 1)
	err = watcher.Watch(changes)
	assert.NoError(t, err, "启动监听应该成功")
	
	// 修改配置文件
	updatedYAML := `
server:
  port: 9999
app:
  debug: true
`
	
	err = os.WriteFile(configFile, []byte(updatedYAML), 0644)
	require.NoError(t, err)
	
	// 等待配置变更通知
	select {
	case newConfig := <-changes:
		assert.NotNil(t, newConfig, "应该接收到新配置")
		assert.Equal(t, 9999, newConfig.Server.Port, "端口应该更新")
		assert.True(t, newConfig.App.Debug, "调试模式应该启用")
	case <-time.After(2 * time.Second):
		t.Fatal("超时：未接收到配置变更通知")
	}
	
	// 停止监听
	err = watcher.Stop()
	assert.NoError(t, err, "停止监听应该成功")
}