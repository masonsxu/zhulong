package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	Port     string
	MinIO    MinIOConfig
	JWT      JWTConfig
	Upload   UploadConfig
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Host      string
	Port      int
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string
	Expire string
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxSize      string
	AllowedTypes string
}

var globalConfig *Config

// LoadConfig 加载配置
func LoadConfig() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	// 加载.env文件，尝试多个位置
	envPaths := []string{".env", "../.env", "../../.env"}
	var envErr error
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			break
		} else {
			envErr = err
		}
	}
	if envErr != nil {
		log.Printf("Warning: .env file not found in any location, using environment variables: %v", envErr)
	}

	port, _ := strconv.Atoi(getEnv("MINIO_PORT", "9000"))

	globalConfig = &Config{
		Port: getEnv("PORT", "8080"),
		MinIO: MinIOConfig{
			Host:      getEnv("MINIO_HOST", "localhost"),
			Port:      port,
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "zhulong-videos"),
			UseSSL:    false,
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

// GetConfig 获取全局配置
func GetConfig() *Config {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}