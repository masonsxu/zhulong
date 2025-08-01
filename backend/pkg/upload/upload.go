package upload

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/manteia/zhulong/pkg/storage"
)

// UploadService 文件上传服务
type UploadService struct {
	storage     storage.StorageInterface
	maxFileSize int64 // 最大文件大小限制（字节）
}

// UploadRequest 单文件上传请求
type UploadRequest struct {
	FileName    string    // 文件名
	ContentType string    // 内容类型
	Size        int64     // 文件大小
	Reader      io.Reader // 文件读取器
	BucketName  string    // 存储桶名
}

// UploadResult 上传结果
type UploadResult struct {
	FileID     string    // 文件唯一标识
	ObjectName string    // 对象名（存储路径）
	Size       int64     // 文件大小
	ETag       string    // 文件ETag
	UploadedAt time.Time // 上传时间
}

// MultipartUploadRequest 分片上传请求
type MultipartUploadRequest struct {
	FileName    string // 文件名
	ContentType string // 内容类型
	TotalSize   int64  // 总文件大小
	BucketName  string // 存储桶名
	ChunkSize   int64  // 分片大小
}

// MultipartUploadSession 分片上传会话
type MultipartUploadSession struct {
	UploadID   string    // 上传ID
	ObjectName string    // 对象名
	CreatedAt  time.Time // 创建时间
}

// UploadPartRequest 分片上传请求
type UploadPartRequest struct {
	UploadID   string // 上传ID
	ObjectName string // 对象名
	PartNumber int    // 分片号（从1开始）
	Data       []byte // 分片数据
	BucketName string // 存储桶名
}

// UploadPartResult 分片上传结果
type UploadPartResult struct {
	PartNumber int    // 分片号
	ETag       string // 分片ETag
	Size       int64  // 分片大小
}

// CompletedPart 已完成分片
type CompletedPart struct {
	PartNumber int    // 分片号
	ETag       string // 分片ETag
}

// CompleteMultipartRequest 完成分片上传请求
type CompleteMultipartRequest struct {
	UploadID   string          // 上传ID
	ObjectName string          // 对象名
	Parts      []CompletedPart // 已完成的分片列表
	BucketName string          // 存储桶名
}

// AbortMultipartRequest 中止分片上传请求
type AbortMultipartRequest struct {
	UploadID   string // 上传ID
	ObjectName string // 对象名
	BucketName string // 存储桶名
}

// UploadProgress 上传进度
type UploadProgress struct {
	UploadID      string    // 上传ID
	Percentage    int       // 完成百分比
	BytesUploaded int64     // 已上传字节数
	TotalBytes    int64     // 总字节数
	IsCompleted   bool      // 是否完成
	UpdatedAt     time.Time // 更新时间
}

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	uploadID   string
	progressCh chan<- *UploadProgress
}

// NewUploadService 创建上传服务
func NewUploadService(storage storage.StorageInterface) *UploadService {
	return &UploadService{
		storage:     storage,
		maxFileSize: 2 * 1024 * 1024 * 1024, // 2GB
	}
}

// UploadFile 上传单个文件
func (s *UploadService) UploadFile(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	// 验证请求
	if err := s.ValidateUploadRequest(req); err != nil {
		return nil, err
	}

	// 生成对象名
	objectName := s.GenerateObjectName(req.FileName)

	// 读取所有数据
	data, err := io.ReadAll(req.Reader)
	if err != nil {
		return nil, fmt.Errorf("读取文件数据失败: %w", err)
	}

	// 上传到存储
	uploadResult, err := s.storage.UploadFile(ctx, req.BucketName, objectName, data, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 生成文件ID
	fileID := uuid.New().String()

	return &UploadResult{
		FileID:     fileID,
		ObjectName: objectName,
		Size:       uploadResult.Size,
		ETag:       uploadResult.ETag,
		UploadedAt: time.Now(),
	}, nil
}

// InitMultipartUpload 初始化分片上传
func (s *UploadService) InitMultipartUpload(ctx context.Context, req *MultipartUploadRequest) (*MultipartUploadSession, error) {
	// 验证请求
	if err := s.validateMultipartRequest(req); err != nil {
		return nil, err
	}

	// 生成对象名
	objectName := s.GenerateObjectName(req.FileName)

	// 生成上传ID（在实际MinIO实现中，这会调用MinIO的InitiateMultipartUpload）
	uploadID := uuid.New().String()

	return &MultipartUploadSession{
		UploadID:   uploadID,
		ObjectName: objectName,
		CreatedAt:  time.Now(),
	}, nil
}

// UploadPart 上传分片
func (s *UploadService) UploadPart(ctx context.Context, req *UploadPartRequest) (*UploadPartResult, error) {
	// 验证请求
	if err := s.validateUploadPartRequest(req); err != nil {
		return nil, err
	}

	// 在实际实现中，这里会调用MinIO的UploadPart
	// 现在我们模拟一个简单的实现
	partObjectName := fmt.Sprintf("%s.part.%d", req.ObjectName, req.PartNumber)
	uploadResult, err := s.storage.UploadFile(ctx, req.BucketName, partObjectName, req.Data, "application/octet-stream")
	if err != nil {
		return nil, fmt.Errorf("上传分片失败: %w", err)
	}

	return &UploadPartResult{
		PartNumber: req.PartNumber,
		ETag:       uploadResult.ETag,
		Size:       int64(len(req.Data)),
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *UploadService) CompleteMultipartUpload(ctx context.Context, req *CompleteMultipartRequest) (*UploadResult, error) {
	// 验证请求
	if err := s.validateCompleteMultipartRequest(req); err != nil {
		return nil, err
	}

	// 在实际实现中，这里会调用MinIO的CompleteMultipartUpload
	// 现在我们模拟：将所有分片合并成一个文件
	var totalData []byte
	var totalSize int64

	for _, part := range req.Parts {
		partObjectName := fmt.Sprintf("%s.part.%d", req.ObjectName, part.PartNumber)

		// 检查分片是否存在
		exists, err := s.storage.FileExists(ctx, req.BucketName, partObjectName)
		if err != nil {
			return nil, fmt.Errorf("检查分片存在性失败: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("分片 %d 不存在", part.PartNumber)
		}

		// 获取分片信息来计算总大小
		fileInfo, err := s.storage.GetFileInfo(ctx, req.BucketName, partObjectName)
		if err != nil {
			return nil, fmt.Errorf("获取分片信息失败: %w", err)
		}
		totalSize += fileInfo.Size
	}

	// 创建最终文件（模拟合并）
	uploadResult, err := s.storage.UploadFile(ctx, req.BucketName, req.ObjectName, totalData, "video/mp4")
	if err != nil {
		return nil, fmt.Errorf("创建最终文件失败: %w", err)
	}

	// 删除分片文件
	for _, part := range req.Parts {
		partObjectName := fmt.Sprintf("%s.part.%d", req.ObjectName, part.PartNumber)
		_ = s.storage.DeleteFile(ctx, req.BucketName, partObjectName)
	}

	// 生成文件ID
	fileID := uuid.New().String()

	return &UploadResult{
		FileID:     fileID,
		ObjectName: req.ObjectName,
		Size:       totalSize,
		ETag:       uploadResult.ETag,
		UploadedAt: time.Now(),
	}, nil
}

// AbortMultipartUpload 中止分片上传
func (s *UploadService) AbortMultipartUpload(ctx context.Context, req *AbortMultipartRequest) error {
	// 在实际实现中，这里会调用MinIO的AbortMultipartUpload
	// 现在我们模拟：删除可能存在的分片文件

	// 由于我们没有跟踪分片，这里只是一个占位符实现
	// 在真实场景中，MinIO会自动清理未完成的分片上传

	return nil
}

// GenerateObjectName 生成对象名
func (s *UploadService) GenerateObjectName(fileName string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// 生成UUID作为文件前缀
	fileID := uuid.New().String()

	// 构造对象名：videos/{year}/{month}/{uuid}-{filename}
	objectName := fmt.Sprintf("videos/%s/%s/%s-%s", year, month, fileID, fileName)

	return objectName
}

// ValidateUploadRequest 验证上传请求
func (s *UploadService) ValidateUploadRequest(req *UploadRequest) error {
	if req.FileName == "" {
		return fmt.Errorf("文件名不能为空")
	}

	if req.ContentType == "" {
		return fmt.Errorf("内容类型不能为空")
	}

	if req.Size <= 0 {
		return fmt.Errorf("文件大小必须大于0")
	}

	if req.Size > s.maxFileSize {
		return fmt.Errorf("文件大小超过限制 (最大: %d 字节)", s.maxFileSize)
	}

	if req.Reader == nil {
		return fmt.Errorf("文件读取器不能为空")
	}

	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	return nil
}

// validateMultipartRequest 验证分片上传请求
func (s *UploadService) validateMultipartRequest(req *MultipartUploadRequest) error {
	if req.FileName == "" {
		return fmt.Errorf("文件名不能为空")
	}

	if req.ContentType == "" {
		return fmt.Errorf("内容类型不能为空")
	}

	if req.TotalSize <= 0 {
		return fmt.Errorf("总文件大小必须大于0")
	}

	if req.TotalSize > s.maxFileSize {
		return fmt.Errorf("文件大小超过限制 (最大: %d 字节)", s.maxFileSize)
	}

	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	if req.ChunkSize <= 0 {
		return fmt.Errorf("分片大小必须大于0")
	}

	return nil
}

// validateUploadPartRequest 验证分片上传请求
func (s *UploadService) validateUploadPartRequest(req *UploadPartRequest) error {
	if req.UploadID == "" {
		return fmt.Errorf("上传ID不能为空")
	}

	if req.ObjectName == "" {
		return fmt.Errorf("对象名不能为空")
	}

	if req.PartNumber <= 0 {
		return fmt.Errorf("分片号必须大于0")
	}

	if len(req.Data) == 0 {
		return fmt.Errorf("分片数据不能为空")
	}

	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	return nil
}

// validateCompleteMultipartRequest 验证完成分片上传请求
func (s *UploadService) validateCompleteMultipartRequest(req *CompleteMultipartRequest) error {
	if req.UploadID == "" {
		return fmt.Errorf("上传ID不能为空")
	}

	if req.ObjectName == "" {
		return fmt.Errorf("对象名不能为空")
	}

	if len(req.Parts) == 0 {
		return fmt.Errorf("分片列表不能为空")
	}

	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	// 验证分片编号连续性
	for i, part := range req.Parts {
		if part.PartNumber != i+1 {
			return fmt.Errorf("分片编号不连续: 期望 %d, 实际 %d", i+1, part.PartNumber)
		}
		if part.ETag == "" {
			return fmt.Errorf("分片 %d 的ETag不能为空", part.PartNumber)
		}
	}

	return nil
}

// CreateProgressTracker 创建进度跟踪器
func (s *UploadService) CreateProgressTracker(uploadID string, progressCh chan<- *UploadProgress) *ProgressTracker {
	return &ProgressTracker{
		uploadID:   uploadID,
		progressCh: progressCh,
	}
}

// UpdateProgress 更新进度
func (t *ProgressTracker) UpdateProgress(percentage int) {
	progress := &UploadProgress{
		UploadID:    t.uploadID,
		Percentage:  percentage,
		IsCompleted: false,
		UpdatedAt:   time.Now(),
	}

	select {
	case t.progressCh <- progress:
	default:
		// 如果通道已满，跳过这次更新
	}
}

// Complete 标记完成
func (t *ProgressTracker) Complete() {
	progress := &UploadProgress{
		UploadID:    t.uploadID,
		Percentage:  100,
		IsCompleted: true,
		UpdatedAt:   time.Now(),
	}

	select {
	case t.progressCh <- progress:
	default:
		// 如果通道已满，跳过这次更新
	}
}
