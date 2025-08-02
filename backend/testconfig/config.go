package testconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

// MinIOTestConfig MinIO测试配置
type MinIOTestConfig struct {
	Host      string
	Port      int
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

var (
	loadEnvOnce sync.Once
)

// loadEnvFile 加载.env文件（只执行一次）
func loadEnvFile() {
	loadEnvOnce.Do(func() {
		// 查找项目根目录的.env文件
		if envPath := findEnvFile(); envPath != "" {
			_ = godotenv.Load(envPath)
		}
	})
}

// findEnvFile 查找.env文件
func findEnvFile() string {
	dir, _ := os.Getwd()
	for {
		envFile := filepath.Join(dir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			return envFile
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// GetMinIOTestConfig 从环境变量获取MinIO测试配置
func GetMinIOTestConfig() *MinIOTestConfig {
	loadEnvFile() // 确保.env文件已加载
	
	return &MinIOTestConfig{
		Host:      getEnvOrDefault("TEST_MINIO_HOST", "localhost"),
		Port:      getEnvIntOrDefault("TEST_MINIO_PORT", 9000),
		AccessKey: getEnvOrDefault("TEST_MINIO_ACCESS_KEY", "admin"),
		SecretKey: getEnvOrDefault("TEST_MINIO_SECRET_KEY", "admin123456"),
		UseSSL:    false, // 默认不使用SSL
		Region:    getEnvOrDefault("TEST_MINIO_REGION", "us-east-1"),
	}
}

// GetEndpoint 获取MinIO端点
func (c *MinIOTestConfig) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault 获取环境变量整数值或返回默认值
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}