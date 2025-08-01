package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
	
	"github.com/manteia/zhulong/pkg/storage"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	MinIO  MinIOConfig  `yaml:"minio"`
	App    AppConfig    `yaml:"app"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Debug   bool   `yaml:"debug"`
}

// ConfigWatcher 配置文件监听器
type ConfigWatcher struct {
	configFile string
	watcher    *fsnotify.Watcher
	stopCh     chan struct{}
}

// LoadFromFile 从文件加载配置
func LoadFromFile(filePath string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", filePath)
	}
	
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	// 解析YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}
	
	// 应用默认值
	config.applyDefaults()
	
	// 应用环境变量覆盖
	config.applyEnvironmentOverrides()
	
	return config, nil
}

// LoadEnvironmentConfig 加载环境特定配置
func LoadEnvironmentConfig(configDir, environment string) (*Config, error) {
	configFile := filepath.Join(configDir, environment+".yml")
	return LoadFromFile(configFile)
}

// applyDefaults 应用默认值
func (c *Config) applyDefaults() {
	// 服务器默认值
	if c.Server.Host == "" {
		c.Server.Host = "localhost"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8888
	}
	
	// MinIO默认值
	if c.MinIO.Region == "" {
		c.MinIO.Region = "us-east-1"
	}
	if c.MinIO.Bucket == "" {
		c.MinIO.Bucket = "zhulong-videos"
	}
	
	// 应用默认值
	if c.App.Name == "" {
		c.App.Name = "Zhulong Video Server"
	}
	if c.App.Version == "" {
		c.App.Version = "v1.0.0"
	}
}

// applyEnvironmentOverrides 应用环境变量覆盖
func (c *Config) applyEnvironmentOverrides() {
	// 服务器配置环境变量覆盖
	if port := os.Getenv("ZHULONG_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}
	if host := os.Getenv("ZHULONG_SERVER_HOST"); host != "" {
		c.Server.Host = host
	}
	
	// MinIO配置环境变量覆盖
	if endpoint := os.Getenv("ZHULONG_MINIO_ENDPOINT"); endpoint != "" {
		c.MinIO.Endpoint = endpoint
	}
	if accessKey := os.Getenv("ZHULONG_MINIO_ACCESS_KEY"); accessKey != "" {
		c.MinIO.AccessKey = accessKey
	}
	if secretKey := os.Getenv("ZHULONG_MINIO_SECRET_KEY"); secretKey != "" {
		c.MinIO.SecretKey = secretKey
	}
	if bucket := os.Getenv("ZHULONG_MINIO_BUCKET"); bucket != "" {
		c.MinIO.Bucket = bucket
	}
	if region := os.Getenv("ZHULONG_MINIO_REGION"); region != "" {
		c.MinIO.Region = region
	}
	if useSSL := os.Getenv("ZHULONG_MINIO_USE_SSL"); useSSL != "" {
		if ssl, err := strconv.ParseBool(useSSL); err == nil {
			c.MinIO.UseSSL = ssl
		}
	}
	
	// 应用配置环境变量覆盖
	if debug := os.Getenv("ZHULONG_APP_DEBUG"); debug != "" {
		if d, err := strconv.ParseBool(debug); err == nil {
			c.App.Debug = d
		}
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	var errors []string
	
	// 验证服务器配置
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, "服务器端口必须在1-65535范围内")
	}
	if c.Server.Host == "" {
		errors = append(errors, "服务器主机不能为空")
	}
	
	// 验证MinIO配置
	if c.MinIO.Endpoint == "" {
		errors = append(errors, "MinIO端点不能为空")
	}
	if c.MinIO.AccessKey == "" {
		errors = append(errors, "MinIO访问密钥不能为空")
	}
	if c.MinIO.SecretKey == "" {
		errors = append(errors, "MinIO秘密密钥不能为空")
	}
	if c.MinIO.Bucket == "" {
		errors = append(errors, "MinIO存储桶不能为空")
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("配置验证失败: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// GetStorageConfig 获取存储配置
func (c *Config) GetStorageConfig() storage.Config {
	return &StorageConfigAdapter{
		endpoint:  c.MinIO.Endpoint,
		accessKey: c.MinIO.AccessKey,
		secretKey: c.MinIO.SecretKey,
		useSSL:    c.MinIO.UseSSL,
		region:    c.MinIO.Region,
	}
}

// StorageConfigAdapter 存储配置适配器
type StorageConfigAdapter struct {
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
	region    string
}

// GetEndpoint 获取端点
func (s *StorageConfigAdapter) GetEndpoint() string {
	return s.endpoint
}

// GetAccessKey 获取访问密钥
func (s *StorageConfigAdapter) GetAccessKey() string {
	return s.accessKey
}

// GetSecretKey 获取秘密密钥
func (s *StorageConfigAdapter) GetSecretKey() string {
	return s.secretKey
}

// IsSSLEnabled 是否启用SSL
func (s *StorageConfigAdapter) IsSSLEnabled() bool {
	return s.useSSL
}

// GetRegion 获取区域
func (s *StorageConfigAdapter) GetRegion() string {
	return s.region
}

// NewConfigWatcher 创建配置文件监听器
func NewConfigWatcher(configFile string) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监听器失败: %w", err)
	}
	
	return &ConfigWatcher{
		configFile: configFile,
		watcher:    watcher,
		stopCh:     make(chan struct{}),
	}, nil
}

// Watch 启动配置监听
func (w *ConfigWatcher) Watch(changes chan<- *Config) error {
	// 添加配置文件到监听列表
	if err := w.watcher.Add(w.configFile); err != nil {
		return fmt.Errorf("添加文件监听失败: %w", err)
	}
	
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				
				// 只处理写入事件
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 重新加载配置
					config, err := LoadFromFile(w.configFile)
					if err == nil {
						select {
						case changes <- config:
						case <-w.stopCh:
							return
						}
					}
				}
				
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				// 在实际应用中，这里应该记录错误日志
				_ = err
				
			case <-w.stopCh:
				return
			}
		}
	}()
	
	return nil
}

// Stop 停止监听
func (w *ConfigWatcher) Stop() error {
	close(w.stopCh)
	return w.watcher.Close()
}