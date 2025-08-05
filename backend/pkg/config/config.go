
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config 结构体定义了应用的所有配置项
type Config struct {
	ServerHost         string   `mapstructure:"ZHULONG_SERVER_HOST"`
	ServerPort         int      `mapstructure:"ZHULONG_SERVER_PORT"`
	S3Host             string   `mapstructure:"ZHULONG_S3_HOST"`
	S3Port             int      `mapstructure:"ZHULONG_S3_PORT"`
	S3AccessKeyID      string   `mapstructure:"ZHULONG_S3_ACCESS_KEY_ID"`
	S3SecretAccessKey  string   `mapstructure:"ZHULONG_S3_SECRET_ACCESS_KEY"`
	S3Bucket           string   `mapstructure:"ZHULONG_S3_BUCKET"`
	S3Region           string   `mapstructure:"ZHULONG_S3_REGION"`
	S3UseSSL           bool     `mapstructure:"ZHULONG_S3_USE_SSL"`
	AppName            string   `mapstructure:"ZHULONG_APP_NAME"`
	AppVersion         string   `mapstructure:"ZHULONG_APP_VERSION"`
	AppDebug           bool     `mapstructure:"ZHULONG_APP_DEBUG"`
	JWTSecret          string   `mapstructure:"JWT_SECRET"`
	JWTExpire          string   `mapstructure:"JWT_EXPIRE"`
	UploadMaxSize      string   `mapstructure:"UPLOAD_MAX_SIZE"`
	UploadAllowedTypes []string `mapstructure:"UPLOAD_ALLOWED_TYPES"`
	PostgresHost       string   `mapstructure:"POSTGRES_HOST"`
	PostgresPort       int      `mapstructure:"POSTGRES_PORT"`
	PostgresUser       string   `mapstructure:"POSTGRES_USER"`
	PostgresPassword   string   `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDBName     string   `mapstructure:"POSTGRES_DBNAME"`
	PostgresSSLMode    string   `mapstructure:"POSTGRES_SSLMODE"`
	NodeEnv            string   `mapstructure:"NODE_ENV"`
}

// LoadConfig 从环境变量和配置文件加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	viper.SetDefault("ZHULONG_SERVER_HOST", "localhost")
	viper.SetDefault("ZHULONG_SERVER_PORT", 8888)
	viper.SetDefault("ZHULONG_S3_HOST", "localhost")
	viper.SetDefault("ZHULONG_S3_PORT", 9000)
	viper.SetDefault("ZHULONG_S3_ACCESS_KEY_ID", "")
	viper.SetDefault("ZHULONG_S3_SECRET_ACCESS_KEY", "")
	viper.SetDefault("ZHULONG_S3_BUCKET", "zhulong-videos")
	viper.SetDefault("ZHULONG_S3_REGION", "us-east-1")
	viper.SetDefault("ZHULONG_S3_USE_SSL", false)
	viper.SetDefault("ZHULONG_APP_NAME", "Zhulong Video Server")
	viper.SetDefault("ZHULONG_APP_VERSION", "v1.0.0")
	viper.SetDefault("ZHULONG_APP_DEBUG", true)
	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_EXPIRE", "7d")
	viper.SetDefault("UPLOAD_MAX_SIZE", "2GB")
	viper.SetDefault("UPLOAD_ALLOWED_TYPES", "video/mp4,video/avi,video/mov,video/webm")
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_USER", "")
	viper.SetDefault("POSTGRES_PASSWORD", "")
	viper.SetDefault("POSTGRES_DBNAME", "zhulong")
	viper.SetDefault("POSTGRES_SSLMODE", "disable")
	viper.SetDefault("NODE_ENV", "development")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// `viper` 对于逗号分割的字符串不能自动解析为 a string slice
	// 所以我们需要手动处理
	if types := viper.GetString("UPLOAD_ALLOWED_TYPES"); types != "" {
		config.UploadAllowedTypes = strings.Split(types, ",")
	}

	return &config, nil
}
