package storage

import (
	"context"
	"time"
)

// StorageInterface 存储服务接口
type StorageInterface interface {
	// 连接测试
	TestConnection(ctx context.Context) error

	// 存储桶操作
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	CreateBucket(ctx context.Context, bucketName string) error
	RemoveBucket(ctx context.Context, bucketName string) error

	// 文件操作
	UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) (*UploadResult, error)
	DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error)
	FileExists(ctx context.Context, bucketName, objectName string) (bool, error)
	GetFileInfo(ctx context.Context, bucketName, objectName string) (*FileInfo, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	ListFiles(ctx context.Context, bucketName, prefix string) ([]*FileInfo, error)

	// URL生成
	GetPresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
	GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration, method string) (string, error)
}

// Config 存储配置接口
type Config interface {
	GetEndpoint() string
	GetAccessKey() string
	GetSecretKey() string
	IsSSLEnabled() bool
	GetRegion() string
}
