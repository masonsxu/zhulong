package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/manteia/zhulong/pkg/storage"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig
	MinIO  MinIOConfig
	App    AppConfig
	JWT    JWTConfig
	Upload UploadConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port int
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
	Bucket    string
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string
	Version string
	Debug   bool
	Env     string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string
	Expire string
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxSize      string
	AllowedTypes []string
}

// Load 从.env文件加载配置
func Load() (*Config, error) {
	// 尝试从多个位置加载.env文件
	envPaths := []string{
		"config/.env",      // 从backend目录相对路径
		"../config/.env",   // 从backend目录的上级目录
		".env",             // 当前目录
	}
	
	var envLoaded bool
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			break
		}
	}
	
	if !envLoaded {
		// 如果所有路径都加载失败，不返回错误，因为环境变量可能通过其他方式设置
		fmt.Println("Warning: 无法从任何路径加载.env文件，将使用系统环境变量")
	}
	
	config := &Config{}
	config.loadFromEnv()
	config.applyDefaults()
	
	return config, nil
}

// LoadFromPath 从指定路径加载配置
func LoadFromPath(envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("加载.env文件失败 [%s]: %w", envPath, err)
	}
	
	config := &Config{}
	config.loadFromEnv()
	config.applyDefaults()
	
	return config, nil
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() {
	// 服务器配置
	c.Server.Host = getEnvString("ZHULONG_SERVER_HOST", "")
	c.Server.Port = getEnvInt("ZHULONG_SERVER_PORT", 0)
	
	// MinIO配置
	c.MinIO.Endpoint = getEnvString("ZHULONG_MINIO_ENDPOINT", "")
	c.MinIO.AccessKey = getEnvString("ZHULONG_MINIO_ACCESS_KEY", "")
	c.MinIO.SecretKey = getEnvString("ZHULONG_MINIO_SECRET_KEY", "")
	c.MinIO.UseSSL = getEnvBool("ZHULONG_MINIO_USE_SSL", false)
	c.MinIO.Region = getEnvString("ZHULONG_MINIO_REGION", "")
	c.MinIO.Bucket = getEnvString("ZHULONG_MINIO_BUCKET", "")
	
	// 应用配置
	c.App.Name = getEnvString("ZHULONG_APP_NAME", "")
	c.App.Version = getEnvString("ZHULONG_APP_VERSION", "")
	c.App.Debug = getEnvBool("ZHULONG_APP_DEBUG", false)
	c.App.Env = getEnvString("NODE_ENV", "")
	
	// JWT配置
	c.JWT.Secret = getEnvString("JWT_SECRET", "")
	c.JWT.Expire = getEnvString("JWT_EXPIRE", "")
	
	// 上传配置
	c.Upload.MaxSize = getEnvString("UPLOAD_MAX_SIZE", "")
	allowedTypes := getEnvString("UPLOAD_ALLOWED_TYPES", "")
	if allowedTypes != "" {
		c.Upload.AllowedTypes = strings.Split(allowedTypes, ",")
		// 去除空格
		for i, t := range c.Upload.AllowedTypes {
			c.Upload.AllowedTypes[i] = strings.TrimSpace(t)
		}
	}
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
	if c.App.Env == "" {
		c.App.Env = "development"
	}
	
	// JWT默认值
	if c.JWT.Expire == "" {
		c.JWT.Expire = "7d"
	}
	
	// 上传默认值
	if c.Upload.MaxSize == "" {
		c.Upload.MaxSize = "2GB"
	}
	if len(c.Upload.AllowedTypes) == 0 {
		c.Upload.AllowedTypes = []string{"video/mp4", "video/avi", "video/mov", "video/webm"}
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
	
	// 验证JWT配置（如果设置了）
	if c.JWT.Secret != "" && len(c.JWT.Secret) < 16 {
		errors = append(errors, "JWT密钥长度不能少于16个字符")
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

// GetServerAddr 获取服务器监听地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Env) == "production"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Env) == "development"
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

// 辅助函数

// getEnvString 获取字符串环境变量
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// PrintConfig 打印配置信息（用于调试，隐藏敏感信息）
func (c *Config) PrintConfig() {
	fmt.Println("==================== 配置信息 ====================")
	fmt.Printf("服务器: %s:%d\n", c.Server.Host, c.Server.Port)
	fmt.Printf("应用: %s %s (env: %s, debug: %v)\n", c.App.Name, c.App.Version, c.App.Env, c.App.Debug)
	fmt.Printf("MinIO: %s (bucket: %s, ssl: %v)\n", c.MinIO.Endpoint, c.MinIO.Bucket, c.MinIO.UseSSL)
	fmt.Printf("MinIO访问密钥: %s\n", maskSensitive(c.MinIO.AccessKey))
	fmt.Printf("MinIO秘密密钥: %s\n", maskSensitive(c.MinIO.SecretKey))
	fmt.Printf("JWT密钥: %s\n", maskSensitive(c.JWT.Secret))
	fmt.Printf("上传限制: %s (%v)\n", c.Upload.MaxSize, c.Upload.AllowedTypes)
	fmt.Println("====================================================")
}

// maskSensitive 遮蔽敏感信息
func maskSensitive(value string) string {
	if len(value) == 0 {
		return "未设置"
	}
	if len(value) <= 6 {
		return "****"
	}
	return value[:3] + "****" + value[len(value)-3:]
}