package download

import (
	"context"
	"fmt"
	"time"

	"github.com/manteia/zhulong/pkg/storage"
)

// DownloadService 文件下载服务
type DownloadService struct {
	storage            storage.StorageInterface
	maxPresignedExpiry time.Duration // 最大预签名URL过期时间
}

// DownloadRequest 文件下载请求
type DownloadRequest struct {
	BucketName string // 存储桶名
	ObjectName string // 对象名
}

// DownloadResult 文件下载结果
type DownloadResult struct {
	Data         []byte    // 文件数据
	Size         int64     // 文件大小
	ContentType  string    // 内容类型
	ObjectName   string    // 对象名
	ETag         string    // 文件ETag
	LastModified time.Time // 最后修改时间
}

// PresignedURLRequest 预签名URL请求
type PresignedURLRequest struct {
	BucketName string        // 存储桶名
	ObjectName string        // 对象名
	ExpiresIn  time.Duration // 过期时间
	Method     string        // HTTP方法 (GET, PUT, DELETE)
}

// PresignedURLResult 预签名URL结果
type PresignedURLResult struct {
	URL        string    // 预签名URL
	ExpiresAt  time.Time // 过期时间
	BucketName string    // 存储桶名
	ObjectName string    // 对象名
	Method     string    // HTTP方法
}

// DownloadURLRequest 下载URL请求
type DownloadURLRequest struct {
	BucketName string        // 存储桶名
	ObjectName string        // 对象名
	ExpiresIn  time.Duration // 过期时间
}

// DownloadURLResult 下载URL结果
type DownloadURLResult struct {
	DownloadURL string    // 下载URL
	ExpiresAt   time.Time // 过期时间
	BucketName  string    // 存储桶名
	ObjectName  string    // 对象名
}

// NewDownloadService 创建下载服务
func NewDownloadService(storage storage.StorageInterface) *DownloadService {
	return &DownloadService{
		storage:            storage,
		maxPresignedExpiry: 7 * 24 * time.Hour, // 最大7天
	}
}

// DownloadFile 下载文件
func (s *DownloadService) DownloadFile(ctx context.Context, req *DownloadRequest) (*DownloadResult, error) {
	// 验证请求
	if err := s.ValidateDownloadRequest(req); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	exists, err := s.storage.FileExists(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("检查文件存在性失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("文件不存在: %s/%s", req.BucketName, req.ObjectName)
	}

	// 获取文件信息
	fileInfo, err := s.storage.GetFileInfo(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 下载文件数据
	data, err := s.storage.DownloadFile(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return &DownloadResult{
		Data:         data,
		Size:         fileInfo.Size,
		ContentType:  fileInfo.ContentType,
		ObjectName:   req.ObjectName,
		ETag:         fileInfo.ETag,
		LastModified: fileInfo.LastModified,
	}, nil
}

// GeneratePresignedURL 生成预签名URL
func (s *DownloadService) GeneratePresignedURL(ctx context.Context, req *PresignedURLRequest) (*PresignedURLResult, error) {
	// 验证请求
	if err := s.ValidatePresignedURLRequest(req); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	exists, err := s.storage.FileExists(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("检查文件存在性失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("文件不存在: %s/%s", req.BucketName, req.ObjectName)
	}

	// 生成预签名URL
	url, err := s.storage.GeneratePresignedURL(ctx, req.BucketName, req.ObjectName, req.ExpiresIn, req.Method)
	if err != nil {
		return nil, fmt.Errorf("生成预签名URL失败: %w", err)
	}

	return &PresignedURLResult{
		URL:        url,
		ExpiresAt:  time.Now().Add(req.ExpiresIn),
		BucketName: req.BucketName,
		ObjectName: req.ObjectName,
		Method:     req.Method,
	}, nil
}

// GenerateDownloadURL 生成下载URL (GET方法的预签名URL)
func (s *DownloadService) GenerateDownloadURL(req *DownloadURLRequest) (*DownloadURLResult, error) {
	// 验证基本参数
	if req.BucketName == "" {
		return nil, fmt.Errorf("存储桶名不能为空")
	}
	if req.ObjectName == "" {
		return nil, fmt.Errorf("对象名不能为空")
	}
	if req.ExpiresIn <= 0 {
		return nil, fmt.Errorf("过期时间必须大于0")
	}
	if req.ExpiresIn > s.maxPresignedExpiry {
		return nil, fmt.Errorf("过期时间不能超过%v", s.maxPresignedExpiry)
	}

	// 这里简化实现，直接构造一个模拟的下载URL
	// 在实际实现中，这会调用MinIO的预签名URL生成
	downloadURL := fmt.Sprintf("http://localhost:9000/%s/%s?expires=%d",
		req.BucketName, req.ObjectName, time.Now().Add(req.ExpiresIn).Unix())

	return &DownloadURLResult{
		DownloadURL: downloadURL,
		ExpiresAt:   time.Now().Add(req.ExpiresIn),
		BucketName:  req.BucketName,
		ObjectName:  req.ObjectName,
	}, nil
}

// ValidateDownloadRequest 验证下载请求
func (s *DownloadService) ValidateDownloadRequest(req *DownloadRequest) error {
	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	if req.ObjectName == "" {
		return fmt.Errorf("对象名不能为空")
	}

	return nil
}

// ValidatePresignedURLRequest 验证预签名URL请求
func (s *DownloadService) ValidatePresignedURLRequest(req *PresignedURLRequest) error {
	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	if req.ObjectName == "" {
		return fmt.Errorf("对象名不能为空")
	}

	if req.ExpiresIn <= 0 {
		return fmt.Errorf("过期时间必须大于0")
	}

	if req.ExpiresIn > s.maxPresignedExpiry {
		return fmt.Errorf("过期时间不能超过7天")
	}

	// 验证HTTP方法
	validMethods := map[string]bool{
		"GET":    true,
		"PUT":    true,
		"DELETE": true,
		"HEAD":   true,
	}

	if !validMethods[req.Method] {
		return fmt.Errorf("不支持的HTTP方法: %s", req.Method)
	}

	return nil
}
