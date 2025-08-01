package config

import (
	"log"
	"os"
	"path/filepath"
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

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}