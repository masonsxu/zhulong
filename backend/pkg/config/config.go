package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	Port   string       `yaml:"-"` // 保持兼容性
	MinIO  MinIOConfig  `yaml:"minio"`
	JWT    JWTConfig    `yaml:"jwt"`
	Upload UploadConfig `yaml:"upload"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint   string                 `yaml:"endpoint"`
	Host       string                 `yaml:"-"` // 保持兼容性
	Port       int                    `yaml:"-"` // 保持兼容性
	AccessKey  string                 `yaml:"access_key"`
	SecretKey  string                 `yaml:"secret_key"`
	Bucket     string                 `yaml:"bucket"`
	UseSSL     bool                   `yaml:"use_ssl"`
	Region     string                 `yaml:"region"`
	Upload     MinIOUploadConfig      `yaml:"upload"`
	Presigned  MinIOPresignedConfig   `yaml:"presigned"`
}

// MinIOUploadConfig MinIO上传配置
type MinIOUploadConfig struct {
	ChunkSize  string `yaml:"chunk_size"`
	MaxRetries int    `yaml:"max_retries"`
	Timeout    string `yaml:"timeout"`
}

// MinIOPresignedConfig MinIO预签名URL配置
type MinIOPresignedConfig struct {
	ExpireTime string `yaml:"expire_time"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire string `yaml:"expire"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxSize      string `yaml:"max_size"`
	AllowedTypes string `yaml:"allowed_types"`
}

// StorageConfig 存储策略配置
type StorageConfig struct {
	PathPattern      string   `yaml:"path_pattern"`
	ThumbnailPattern string   `yaml:"thumbnail_pattern"`
	AllowedFormats   []string `yaml:"allowed_formats"`
	MaxFileSize      string   `yaml:"max_file_size"`
}

// MinIOFullConfig MinIO完整配置结构
type MinIOFullConfig struct {
	MinIO   MinIOConfig   `yaml:"minio"`
	Storage StorageConfig `yaml:"storage"`
}

var globalConfig *Config

// LoadConfig 加载配置
func LoadConfig() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	// 查找.env文件，从当前目录开始向上查找到项目根目录
	envFile := findEnvFile()
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("Warning: failed to load .env file from %s: %v", envFile, err)
		} else {
			log.Printf("Loaded .env file from: %s", envFile)
		}
	} else {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	port, _ := strconv.Atoi(getEnv("MINIO_PORT", "9000"))

	globalConfig = &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getIntEnv("PORT", 8080),
		},
		Port: getEnv("PORT", "8080"), // 保持兼容性
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_HOST", "localhost") + ":" + getEnv("MINIO_PORT", "9000"),
			Host:      getEnv("MINIO_HOST", "localhost"),      // 保持兼容性
			Port:      port,                                   // 保持兼容性
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "zhulong-videos"),
			UseSSL:    getBoolEnv("MINIO_USE_SSL", false),
			Region:    getEnv("MINIO_REGION", "us-east-1"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
			Expire: getEnv("JWT_EXPIRE", "7d"),
		},
		Upload: UploadConfig{
			MaxSize:      getEnv("UPLOAD_MAX_SIZE", "10MB"),
			AllowedTypes: getEnv("UPLOAD_ALLOWED_TYPES", "video/mp4,video/avi"),
		},
	}

	return globalConfig
}

// LoadFromYAML 从YAML文件加载配置
func LoadFromYAML(environment string) (*Config, error) {
	configPath := findConfigFile(environment + ".yml")
	if configPath == "" {
		return nil, fmt.Errorf("config file not found for environment: %s", environment)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 先进行环境变量替换
	content := expandEnvVars(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置兼容性字段
	config.Port = fmt.Sprintf("%d", config.Server.Port)
	if config.MinIO.Endpoint != "" {
		parts := strings.Split(config.MinIO.Endpoint, ":")
		if len(parts) == 2 {
			config.MinIO.Host = parts[0]
			if port, err := strconv.Atoi(parts[1]); err == nil {
				config.MinIO.Port = port
			}
		}
	}

	return &config, nil
}

// LoadMinIOConfig 加载MinIO专用配置
func LoadMinIOConfig() (*MinIOFullConfig, error) {
	configPath := findConfigFile("minio.yml")
	if configPath == "" {
		return nil, fmt.Errorf("minio.yml config file not found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read minio config file: %w", err)
	}

	// 环境变量替换
	content := expandEnvVars(string(data))

	var config MinIOFullConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse minio config file: %w", err)
	}

	return &config, nil
}

// findConfigFile 查找配置文件
func findConfigFile(filename string) string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// 向上查找配置文件，最多查找5级
	for i := 0; i < 5; i++ {
		configPath := filepath.Join(dir, "config", filename)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	
	return ""
}

// expandEnvVars 展开环境变量 ${VAR_NAME}
func expandEnvVars(content string) string {
	return os.Expand(content, func(key string) string {
		return os.Getenv(key)
	})
}

// findEnvFile 查找.env文件，从当前目录向上查找
func findEnvFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// 向上查找.env文件，最多查找5级
	for i := 0; i < 5; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已到达根目录
			break
		}
		dir = parent
	}
	
	return ""
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig
}

// 环境变量获取辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}