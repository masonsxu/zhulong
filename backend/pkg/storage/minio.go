package storage

import (
	"context"

	"github.com/minio/minio-go/v7"
)

// StorageInterface 存储接口
type StorageInterface interface {
	UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error
	GetPresignedURL(ctx context.Context, bucketName, objectName string) (string, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	ListFiles(ctx context.Context, bucketName, prefix string) ([]minio.ObjectInfo, error)
}

// MinIOStorage MinIO存储实现
type MinIOStorage struct {
	client *minio.Client
}

// NewMinIOStorage 创建MinIO存储实例
func NewMinIOStorage(client *minio.Client) *MinIOStorage {
	return &MinIOStorage{
		client: client,
	}
}

// UploadFile 上传文件
func (s *MinIOStorage) UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error {
	// TODO: 实现文件上传逻辑
	return nil
}

// GetPresignedURL 获取预签名URL
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, bucketName, objectName string) (string, error) {
	// TODO: 实现预签名URL生成
	return "", nil
}

// DeleteFile 删除文件
func (s *MinIOStorage) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	// TODO: 实现文件删除逻辑
	return nil
}

// ListFiles 列出文件
func (s *MinIOStorage) ListFiles(ctx context.Context, bucketName, prefix string) ([]minio.ObjectInfo, error) {
	// TODO: 实现文件列表逻辑
	return nil, nil
}