package config

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	MinIO  MinIOConfig  `yaml:"minio"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host" default:"localhost"`
	Port int    `yaml:"port" default:"8080"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl" default:"false"`
	Bucket          string `yaml:"bucket" default:"zhulong-videos"`
}