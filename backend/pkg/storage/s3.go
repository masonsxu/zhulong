package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Config S3兼容存储配置结构
type S3Config struct {
	Endpoint  string // S3服务端点
	AccessKey string // 访问密钥
	SecretKey string // 秘密密钥
	UseSSL    bool   // 是否使用SSL
	Region    string // 区域
}

// GetEndpoint 获取服务端点
func (c *S3Config) GetEndpoint() string {
	return c.Endpoint
}

// GetAccessKey 获取访问密钥
func (c *S3Config) GetAccessKey() string {
	return c.AccessKey
}

// GetSecretKey 获取秘密密钥
func (c *S3Config) GetSecretKey() string {
	return c.SecretKey
}

// IsSSLEnabled 是否启用SSL
func (c *S3Config) IsSSLEnabled() bool {
	return c.UseSSL
}

// GetRegion 获取区域
func (c *S3Config) GetRegion() string {
	return c.Region
}

func (c *S3Config) toAWSConfig(ctx context.Context) (aws.Config, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if c.Endpoint != "" {
			return aws.Endpoint{
					URL:           c.Endpoint,
					SigningRegion: c.Region,
					Source:        aws.EndpointSourceCustom,
				},
				nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.GetRegion()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.GetAccessKey(), c.GetSecretKey(), "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("加载AWS配置失败: %w", err)
	}

	return cfg, nil
}

// S3Storage S3兼容存储服务
type S3Storage struct {
	client *s3.Client
	config Config
}

// 确保S3Storage实现了StorageInterface接口
var _ StorageInterface = (*S3Storage)(nil)

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

// NewS3Storage 创建S3兼容存储服务实例
func NewS3Storage(config *S3Config) (*S3Storage, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	cfg, err := config.toAWSConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("创建AWS配置失败: %w", err)
	}

	// TODO，强制使用路径风格
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Storage{
		client: client,
		config: config,
	}, nil
}

// TestConnection 测试连接
func (s *S3Storage) TestConnection(ctx context.Context) error {
	_, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("S3连接测试失败: %w", err)
	}
	return nil
}

// BucketExists 检查存储桶是否存在
func (s *S3Storage) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var nbf *types.NotFound
		if errors.As(err, &nbf) {
			return false, nil
		}
		return false, fmt.Errorf("检查存储桶存在性失败: %w", err)
	}
	return true, nil
}

// CreateBucket 创建存储桶
func (s *S3Storage) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.config.GetRegion()),
		},
	})
	if err != nil {
		return fmt.Errorf("创建存储桶失败: %w", err)
	}
	return nil
}

// RemoveBucket 删除存储桶
func (s *S3Storage) RemoveBucket(ctx context.Context, bucketName string) error {
	_, err := s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("删除存储桶失败: %w", err)
	}
	return nil
}

// UploadFile 上传文件
func (s *S3Storage) UploadFile(ctx context.Context, bucketName, objectName string, data []byte, contentType string) (*UploadResult, error) {
	reader := bytes.NewReader(data)

	putObjectOutput, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectName),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: int64(len(data)),
	})
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	return &UploadResult{
		ETag: *putObjectOutput.ETag,
		Size: int64(len(data)), // AWS SDK PutObject不直接返回Size，这里使用传入数据的大小
	}, nil
}

// FileExists 检查文件是否存在
func (s *S3Storage) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		var nsf *types.NotFound
		if errors.As(err, &nsf) {
			return false, nil
		}
		return false, fmt.Errorf("检查文件存在性失败: %w", err)
	}
	return true, nil
}

// GetPresignedURL 生成预签名URL
func (s *S3Storage) GetPresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(s.client)

	presignedUrl, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %w", err)
	}

	return presignedUrl.URL, nil
}

// GetFileInfo 获取文件信息
func (s *S3Storage) GetFileInfo(ctx context.Context, bucketName, objectName string) (*FileInfo, error) {
	headObjectOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &FileInfo{
		Key:          objectName,
		Size:         headObjectOutput.ContentLength,
		ContentType:  *headObjectOutput.ContentType,
		LastModified: *headObjectOutput.LastModified,
		ETag:         *headObjectOutput.ETag,
	}, nil
}

// DeleteFile 删除文件
func (s *S3Storage) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// ListFiles 列出文件
func (s *S3Storage) ListFiles(ctx context.Context, bucketName, prefix string) ([]*FileInfo, error) {
	var files []*FileInfo

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("列出文件失败: %w", err)
		}

		for _, obj := range page.Contents {
			// 跳过目录对象（以/结尾且大小为0）
			if strings.HasSuffix(*obj.Key, "/") && obj.Size == 0 {
				continue
			}

			files = append(files, &FileInfo{
				Key:          *obj.Key,
				Size:         obj.Size,
				ContentType:  "", // ListObjectsV2不直接返回ContentType
				LastModified: *obj.LastModified,
				ETag:         *obj.ETag,
			})
		}
	}

	return files, nil
}

// DownloadFile 下载文件
func (s *S3Storage) DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	getObjectOutput, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}
	defer getObjectOutput.Body.Close()

	data, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return nil, fmt.Errorf("读取文件数据失败: %w", err)
	}

	return data, nil
}

// GeneratePresignedURL 生成预签名URL（支持不同HTTP方法）
func (s *S3Storage) GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration, method string) (string, error) {
	presigner := s3.NewPresignClient(s.client)

	switch method {
	case "GET":
		presignedURL, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectName),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = expiry
		})
		if err != nil {
			return "", fmt.Errorf("生成GET预签名URL失败: %w", err)
		}
		return presignedURL.URL, nil

	case "PUT":
		presignedURL, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectName),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = expiry
		})
		if err != nil {
			return "", fmt.Errorf("生成PUT预签名URL失败: %w", err)
		}
		return presignedURL.URL, nil

	case "DELETE":
		// AWS S3 PresignClient 不直接支持 DELETE 操作的预签名 URL
		// 通常DELETE操作不通过预签名URL完成，或者需要自定义签名逻辑
		return "", fmt.Errorf("不支持的HTTP方法: %s", method)

	case "HEAD":
		// AWS S3 PresignClient 不直接支持 HEAD 操作的预签名 URL
		return "", fmt.Errorf("不支持的HTTP方法: %s", method)

	default:
		return "", fmt.Errorf("不支持的HTTP方法: %s", method)
	}
}
