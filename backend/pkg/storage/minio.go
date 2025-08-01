package storage

import (
	"context"
	"fmt"

	"github.com/manteia/zhulong/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageInterface 存储接口
type StorageInterface interface {
	UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error
	GetPresignedURL(ctx context.Context, bucketName, objectName string) (string, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	ListFiles(ctx context.Context, bucketName, prefix string) ([]minio.ObjectInfo, error)
}

// MinIOClient MinIO客户端封装
type MinIOClient struct {
	client *minio.Client
	config config.MinIOConfig
}

// MinIOStorage MinIO存储实现
type MinIOStorage struct {
	client *minio.Client
}

// NewMinIOClient 创建新的MinIO客户端
func NewMinIOClient() (*MinIOClient, error) {
	cfg := config.GetConfig()
	
	endpoint := fmt.Sprintf("%s:%d", cfg.MinIO.Host, cfg.MinIO.Port)
	
	// 初始化MinIO客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}
	
	return &MinIOClient{
		client: minioClient,
		config: cfg.MinIO,
	}, nil
}

// TestConnection 测试MinIO连接
func (c *MinIOClient) TestConnection() error {
	ctx := context.Background()
	
	// 通过列出存储桶来测试连接
	_, err := c.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to MinIO: %w", err)
	}
	
	return nil
}

// BucketExists 检查存储桶是否存在
func (c *MinIOClient) BucketExists(bucketName string) (bool, error) {
	ctx := context.Background()
	
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	
	return exists, nil
}

// CreateBucket 创建存储桶
func (c *MinIOClient) CreateBucket(bucketName string) error {
	ctx := context.Background()
	
	err := c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	
	return nil
}

// GetMinIOConfig 获取MinIO配置 (用于测试)
func GetMinIOConfig() *config.MinIOConfig {
	cfg := config.GetConfig()
	return &cfg.MinIO
}

// GetClient 获取底层MinIO客户端 (用于高级操作)
func (c *MinIOClient) GetClient() *minio.Client {
	return c.client
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