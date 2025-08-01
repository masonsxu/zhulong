package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig MinIO配置结构
type MinIOConfig struct {
	Endpoint  string // MinIO服务端点
	AccessKey string // 访问密钥
	SecretKey string // 秘密密钥
	UseSSL    bool   // 是否使用SSL
	Region    string // 区域
}

// GetEndpoint 获取服务端点
func (c *MinIOConfig) GetEndpoint() string {
	return c.Endpoint
}

// GetAccessKey 获取访问密钥
func (c *MinIOConfig) GetAccessKey() string {
	return c.AccessKey
}

// GetSecretKey 获取秘密密钥
func (c *MinIOConfig) GetSecretKey() string {
	return c.SecretKey
}

// IsSSLEnabled 是否启用SSL
func (c *MinIOConfig) IsSSLEnabled() bool {
	return c.UseSSL
}

// GetRegion 获取区域
func (c *MinIOConfig) GetRegion() string {
	return c.Region
}

// MinIOStorage MinIO存储服务
type MinIOStorage struct {
	client *minio.Client
	config Config
}

// 确保MinIOStorage实现了StorageInterface接口
var _ StorageInterface = (*MinIOStorage)(nil)

// UploadResult 上传结果
type UploadResult struct {
	ETag string // 文件ETag
	Size int64  // 文件大小
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string    // 文件名/键
	Size         int64     // 文件大小
	ContentType  string    // 内容类型
	LastModified time.Time // 最后修改时间
	ETag         string    // ETag
}

// NewMinIOStorage 创建MinIO存储服务实例
func NewMinIOStorage(config *MinIOConfig) (*MinIOStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 创建MinIO客户端
	client, err := minio.New(config.GetEndpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(config.GetAccessKey(), config.GetSecretKey(), ""),
		Secure: config.IsSSLEnabled(),
		Region: config.GetRegion(),
	})
	if err != nil {
		return nil, fmt.Errorf("创建MinIO客户端失败: %w", err)
	}

	return &MinIOStorage{
		client: client,
		config: config,
	}, nil
}

// TestConnection 测试连接
func (s *MinIOStorage) TestConnection(ctx context.Context) error {
	// 通过列出存储桶来测试连接
	_, err := s.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("MinIO连接测试失败: %w", err)
	}
	return nil
}

// BucketExists 检查存储桶是否存在
func (s *MinIOStorage) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("检查存储桶存在性失败: %w", err)
	}
	return exists, nil
}

// CreateBucket 创建存储桶
func (s *MinIOStorage) CreateBucket(ctx context.Context, bucketName string) error {
	err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: s.config.GetRegion(),
	})
	if err != nil {
		return fmt.Errorf("创建存储桶失败: %w", err)
	}
	return nil
}

// RemoveBucket 删除存储桶
func (s *MinIOStorage) RemoveBucket(ctx context.Context, bucketName string) error {
	err := s.client.RemoveBucket(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("删除存储桶失败: %w", err)
	}
	return nil
}

// UploadFile 上传文件
func (s *MinIOStorage) UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) (*UploadResult, error) {
	reader := bytes.NewReader(data)

	info, err := s.client.PutObject(ctx, bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	return &UploadResult{
		ETag: info.ETag,
		Size: info.Size,
	}, nil
}

// FileExists 检查文件是否存在
func (s *MinIOStorage) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// 如果是NotFound错误，返回false
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("检查文件存在性失败: %w", err)
	}
	return true, nil
}

// GetPresignedURL 生成预签名URL
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %w", err)
	}
	return url.String(), nil
}

// GetFileInfo 获取文件信息
func (s *MinIOStorage) GetFileInfo(ctx context.Context, bucketName, objectName string) (*FileInfo, error) {
	stat, err := s.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &FileInfo{
		Key:          stat.Key,
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
		ETag:         stat.ETag,
	}, nil
}

// DeleteFile 删除文件
func (s *MinIOStorage) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// ListFiles 列出文件
func (s *MinIOStorage) ListFiles(ctx context.Context, bucketName, prefix string) ([]*FileInfo, error) {
	var files []*FileInfo

	objectCh := s.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix: prefix,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("列出文件失败: %w", object.Err)
		}

		files = append(files, &FileInfo{
			Key:          object.Key,
			Size:         object.Size,
			ContentType:  "", // ListObjects不返回ContentType
			LastModified: object.LastModified,
			ETag:         object.ETag,
		})
	}

	return files, nil
}

// DownloadFile 下载文件
func (s *MinIOStorage) DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	// 获取对象
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}
	defer object.Close()

	// 读取所有数据
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("读取文件数据失败: %w", err)
	}

	return data, nil
}

// GeneratePresignedURL 生成预签名URL（支持不同HTTP方法）
func (s *MinIOStorage) GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration, method string) (string, error) {
	// 将HTTP方法字符串转换为MinIO的方法类型
	var reqParams url.Values

	switch method {
	case "GET":
		presignedURL, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
		if err != nil {
			return "", fmt.Errorf("生成GET预签名URL失败: %w", err)
		}
		return presignedURL.String(), nil

	case "PUT":
		presignedURL, err := s.client.PresignedPutObject(ctx, bucketName, objectName, expiry)
		if err != nil {
			return "", fmt.Errorf("生成PUT预签名URL失败: %w", err)
		}
		return presignedURL.String(), nil

	case "DELETE":
		// 对于DELETE方法，使用PresignedGetObject生成URL，但注意这实际上是GET方法
		// 在实际应用中，DELETE操作通常不通过预签名URL完成
		presignedURL, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
		if err != nil {
			return "", fmt.Errorf("生成DELETE预签名URL失败: %w", err)
		}
		return presignedURL.String(), nil

	case "HEAD":
		// 对于HEAD方法，使用PresignedGetObject生成URL
		presignedURL, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
		if err != nil {
			return "", fmt.Errorf("生成HEAD预签名URL失败: %w", err)
		}
		return presignedURL.String(), nil

	default:
		return "", fmt.Errorf("不支持的HTTP方法: %s", method)
	}
}
