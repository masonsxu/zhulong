package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// 保存原始环境变量
	originalEnvs := make(map[string]string)
	envKeys := []string{
		"ZHULONG_SERVER_HOST", "ZHULONG_SERVER_PORT",
		"ZHULONG_MINIO_ENDPOINT", "ZHULONG_MINIO_ACCESS_KEY", "ZHULONG_MINIO_SECRET_KEY",
		"ZHULONG_MINIO_BUCKET", "ZHULONG_MINIO_REGION", "ZHULONG_MINIO_USE_SSL",
		"ZHULONG_APP_NAME", "ZHULONG_APP_VERSION", "ZHULONG_APP_DEBUG",
		"JWT_SECRET", "JWT_EXPIRE", "UPLOAD_MAX_SIZE", "UPLOAD_ALLOWED_TYPES", "NODE_ENV",
	}
	
	for _, key := range envKeys {
		originalEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	
	defer func() {
		for key, value := range originalEnvs {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("默认配置加载", func(t *testing.T) {
		config, err := Load()
		require.NoError(t, err)
		
		// 验证默认值
		assert.Equal(t, "localhost", config.Server.Host)
		assert.Equal(t, 8888, config.Server.Port)
		assert.Equal(t, "Zhulong Video Server", config.App.Name)
		assert.Equal(t, "v1.0.0", config.App.Version)
		assert.Equal(t, "development", config.App.Env)
		assert.Equal(t, "us-east-1", config.MinIO.Region)
		assert.Equal(t, "zhulong-videos", config.MinIO.Bucket)
		assert.Equal(t, "7d", config.JWT.Expire)
		assert.Equal(t, "2GB", config.Upload.MaxSize)
		assert.Equal(t, []string{"video/mp4", "video/avi", "video/mov", "video/webm"}, config.Upload.AllowedTypes)
	})

	t.Run("环境变量覆盖", func(t *testing.T) {
		// 设置环境变量
		os.Setenv("ZHULONG_SERVER_HOST", "0.0.0.0")
		os.Setenv("ZHULONG_SERVER_PORT", "9999")
		os.Setenv("ZHULONG_MINIO_ENDPOINT", "minio.example.com:9000")
		os.Setenv("ZHULONG_MINIO_ACCESS_KEY", "testkey")
		os.Setenv("ZHULONG_MINIO_SECRET_KEY", "testsecret")
		os.Setenv("ZHULONG_MINIO_BUCKET", "test-bucket")
		os.Setenv("ZHULONG_MINIO_USE_SSL", "true")
		os.Setenv("ZHULONG_APP_DEBUG", "false")
		os.Setenv("JWT_SECRET", "super-secret-key-for-test")
		os.Setenv("UPLOAD_MAX_SIZE", "1GB")
		os.Setenv("UPLOAD_ALLOWED_TYPES", "video/mp4,video/avi")
		os.Setenv("NODE_ENV", "production")
		
		config, err := Load()
		require.NoError(t, err)
		
		// 验证环境变量覆盖生效
		assert.Equal(t, "0.0.0.0", config.Server.Host)
		assert.Equal(t, 9999, config.Server.Port)
		assert.Equal(t, "minio.example.com:9000", config.MinIO.Endpoint)
		assert.Equal(t, "testkey", config.MinIO.AccessKey)
		assert.Equal(t, "testsecret", config.MinIO.SecretKey)
		assert.Equal(t, "test-bucket", config.MinIO.Bucket)
		assert.True(t, config.MinIO.UseSSL)
		assert.False(t, config.App.Debug)
		assert.Equal(t, "super-secret-key-for-test", config.JWT.Secret)
		assert.Equal(t, "1GB", config.Upload.MaxSize)
		assert.Equal(t, []string{"video/mp4", "video/avi"}, config.Upload.AllowedTypes)
		assert.Equal(t, "production", config.App.Env)
	})
}

func TestValidate(t *testing.T) {
	t.Run("有效配置", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: 8888,
			},
			MinIO: MinIOConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "admin",
				SecretKey: "admin123456",
				Bucket:    "test-bucket",
			},
			JWT: JWTConfig{
				Secret: "very-long-secret-key",
			},
		}
		
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("无效端口", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: 70000, // 无效端口
			},
			MinIO: MinIOConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "admin",
				SecretKey: "admin123456",
				Bucket:    "test-bucket",
			},
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "端口必须在1-65535范围内")
	})

	t.Run("缺少MinIO配置", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: 8888,
			},
			MinIO: MinIOConfig{
				// 缺少必要的MinIO配置
			},
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MinIO端点不能为空")
	})

	t.Run("JWT密钥过短", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: 8888,
			},
			MinIO: MinIOConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "admin",
				SecretKey: "admin123456",
				Bucket:    "test-bucket",
			},
			JWT: JWTConfig{
				Secret: "short", // 密钥过短
			},
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT密钥长度不能少于16个字符")
	})
}

func TestGetServerAddr(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 9999,
		},
	}
	
	addr := config.GetServerAddr()
	assert.Equal(t, "0.0.0.0:9999", addr)
}

func TestEnvironmentCheckers(t *testing.T) {
	t.Run("IsProduction", func(t *testing.T) {
		config := &Config{App: AppConfig{Env: "production"}}
		assert.True(t, config.IsProduction())
		
		config.App.Env = "PRODUCTION"
		assert.True(t, config.IsProduction())
		
		config.App.Env = "development"
		assert.False(t, config.IsProduction())
	})

	t.Run("IsDevelopment", func(t *testing.T) {
		config := &Config{App: AppConfig{Env: "development"}}
		assert.True(t, config.IsDevelopment())
		
		config.App.Env = "DEVELOPMENT"
		assert.True(t, config.IsDevelopment())
		
		config.App.Env = "production"
		assert.False(t, config.IsDevelopment())
	})
}

func TestGetStorageConfig(t *testing.T) {
	config := &Config{
		MinIO: MinIOConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "admin",
			SecretKey: "admin123456",
			UseSSL:    true,
			Region:    "us-west-2",
		},
	}
	
	storageConfig := config.GetStorageConfig()
	assert.Equal(t, "localhost:9000", storageConfig.GetEndpoint())
	assert.Equal(t, "admin", storageConfig.GetAccessKey())
	assert.Equal(t, "admin123456", storageConfig.GetSecretKey())
	assert.True(t, storageConfig.IsSSLEnabled())
	assert.Equal(t, "us-west-2", storageConfig.GetRegion())
}

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "未设置"},
		{"short", "****"},
		{"123456", "****"},
		{"1234567", "123****567"},
		{"verylongsecretkey123", "ver****123"},
	}

	for _, test := range tests {
		result := maskSensitive(test.input)
		assert.Equal(t, test.expected, result, "输入: %s", test.input)
	}
}

func TestUploadTypeParsing(t *testing.T) {
	// 保存原始环境变量
	original := os.Getenv("UPLOAD_ALLOWED_TYPES")
	defer func() {
		if original != "" {
			os.Setenv("UPLOAD_ALLOWED_TYPES", original)
		} else {
			os.Unsetenv("UPLOAD_ALLOWED_TYPES")
		}
	}()

	t.Run("解析上传类型列表", func(t *testing.T) {
		os.Setenv("UPLOAD_ALLOWED_TYPES", "video/mp4, video/avi , video/mov")
		
		config, err := Load()
		require.NoError(t, err)
		
		expected := []string{"video/mp4", "video/avi", "video/mov"}
		assert.Equal(t, expected, config.Upload.AllowedTypes)
	})

	t.Run("空上传类型使用默认值", func(t *testing.T) {
		os.Unsetenv("UPLOAD_ALLOWED_TYPES")
		
		config, err := Load()
		require.NoError(t, err)
		
		expected := []string{"video/mp4", "video/avi", "video/mov", "video/webm"}
		assert.Equal(t, expected, config.Upload.AllowedTypes)
	})
}