package delete

import (
	"context"
	"fmt"
	"time"

	"github.com/manteia/zhulong/pkg/storage"
)

// DeleteService 文件删除服务
type DeleteService struct {
	storage       storage.StorageInterface
	maxBatchSize  int           // 批量删除最大文件数
	deleteTimeout time.Duration // 删除操作超时时间
}

// DeleteRequest 单文件删除请求
type DeleteRequest struct {
	BucketName string // 存储桶名
	ObjectName string // 对象名
}

// DeleteResult 单文件删除结果
type DeleteResult struct {
	BucketName   string    // 存储桶名
	ObjectName   string    // 对象名
	Success      bool      // 是否成功
	ErrorMessage string    // 错误信息（失败时）
	DeletedAt    time.Time // 删除时间
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	BucketName  string   // 存储桶名
	ObjectNames []string // 对象名列表
}

// BatchDeleteResult 批量删除结果
type BatchDeleteResult struct {
	Results      []*DeleteResult // 每个文件的删除结果
	TotalCount   int             // 总文件数
	SuccessCount int             // 成功删除数
	FailureCount int             // 删除失败数
	ProcessedAt  time.Time       // 处理时间
}

// PrefixDeleteRequest 按前缀删除请求
type PrefixDeleteRequest struct {
	BucketName string // 存储桶名
	Prefix     string // 文件前缀
}

// PrefixDeleteResult 按前缀删除结果
type PrefixDeleteResult struct {
	DeletedCount int       // 删除的文件数量
	DeletedFiles []string  // 删除的文件列表
	ProcessedAt  time.Time // 处理时间
}

// NewDeleteService 创建删除服务
func NewDeleteService(storage storage.StorageInterface) *DeleteService {
	return &DeleteService{
		storage:       storage,
		maxBatchSize:  1000,             // 一次最多删除1000个文件
		deleteTimeout: 30 * time.Second, // 30秒超时
	}
}

// DeleteFile 删除单个文件
func (s *DeleteService) DeleteFile(ctx context.Context, req *DeleteRequest) (*DeleteResult, error) {
	// 验证请求
	if err := s.ValidateDeleteRequest(req); err != nil {
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

	// 删除文件
	err = s.storage.DeleteFile(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return &DeleteResult{
			BucketName:   req.BucketName,
			ObjectName:   req.ObjectName,
			Success:      false,
			ErrorMessage: err.Error(),
			DeletedAt:    time.Now(),
		}, fmt.Errorf("删除文件失败: %w", err)
	}

	return &DeleteResult{
		BucketName:   req.BucketName,
		ObjectName:   req.ObjectName,
		Success:      true,
		ErrorMessage: "",
		DeletedAt:    time.Now(),
	}, nil
}

// DeleteMultipleFiles 批量删除文件
func (s *DeleteService) DeleteMultipleFiles(ctx context.Context, req *BatchDeleteRequest) (*BatchDeleteResult, error) {
	// 验证请求
	if err := s.ValidateBatchDeleteRequest(req); err != nil {
		return nil, err
	}

	results := make([]*DeleteResult, len(req.ObjectNames))
	successCount := 0
	failureCount := 0

	// 逐个删除文件
	for i, objectName := range req.ObjectNames {
		deleteReq := &DeleteRequest{
			BucketName: req.BucketName,
			ObjectName: objectName,
		}

		// 尝试删除单个文件
		result, err := s.deleteSingleFile(ctx, deleteReq)
		results[i] = result

		if result.Success {
			successCount++
		} else {
			failureCount++
		}

		// 如果是验证错误，记录但继续处理其他文件
		if err != nil && result.Success == false {
			// 已经在result中记录了错误信息
			continue
		}
	}

	return &BatchDeleteResult{
		Results:      results,
		TotalCount:   len(req.ObjectNames),
		SuccessCount: successCount,
		FailureCount: failureCount,
		ProcessedAt:  time.Now(),
	}, nil
}

// deleteSingleFile 删除单个文件（内部方法，不进行请求验证）
func (s *DeleteService) deleteSingleFile(ctx context.Context, req *DeleteRequest) (*DeleteResult, error) {
	// 检查文件是否存在
	exists, err := s.storage.FileExists(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return &DeleteResult{
			BucketName:   req.BucketName,
			ObjectName:   req.ObjectName,
			Success:      false,
			ErrorMessage: fmt.Sprintf("检查文件存在性失败: %v", err),
			DeletedAt:    time.Now(),
		}, nil // 返回nil错误，因为错误已记录在result中
	}

	if !exists {
		return &DeleteResult{
			BucketName:   req.BucketName,
			ObjectName:   req.ObjectName,
			Success:      false,
			ErrorMessage: "文件不存在",
			DeletedAt:    time.Now(),
		}, nil
	}

	// 删除文件
	err = s.storage.DeleteFile(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		return &DeleteResult{
			BucketName:   req.BucketName,
			ObjectName:   req.ObjectName,
			Success:      false,
			ErrorMessage: fmt.Sprintf("删除文件失败: %v", err),
			DeletedAt:    time.Now(),
		}, nil
	}

	return &DeleteResult{
		BucketName:   req.BucketName,
		ObjectName:   req.ObjectName,
		Success:      true,
		ErrorMessage: "",
		DeletedAt:    time.Now(),
	}, nil
}

// DeleteFilesByPrefix 按前缀删除文件
func (s *DeleteService) DeleteFilesByPrefix(ctx context.Context, req *PrefixDeleteRequest) (*PrefixDeleteResult, error) {
	// 验证请求
	if err := s.validatePrefixDeleteRequest(req); err != nil {
		return nil, err
	}

	// 列出匹配前缀的文件
	files, err := s.storage.ListFiles(ctx, req.BucketName, req.Prefix)
	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %w", err)
	}

	if len(files) == 0 {
		return &PrefixDeleteResult{
			DeletedCount: 0,
			DeletedFiles: []string{},
			ProcessedAt:  time.Now(),
		}, nil
	}

	// 批量删除匹配的文件
	objectNames := make([]string, len(files))
	for i, file := range files {
		objectNames[i] = file.Key
	}

	batchRequest := &BatchDeleteRequest{
		BucketName:  req.BucketName,
		ObjectNames: objectNames,
	}

	batchResult, err := s.DeleteMultipleFiles(ctx, batchRequest)
	if err != nil {
		return nil, fmt.Errorf("批量删除失败: %w", err)
	}

	// 统计成功删除的文件
	deletedFiles := make([]string, 0, batchResult.SuccessCount)
	for _, result := range batchResult.Results {
		if result.Success {
			deletedFiles = append(deletedFiles, result.ObjectName)
		}
	}

	return &PrefixDeleteResult{
		DeletedCount: batchResult.SuccessCount,
		DeletedFiles: deletedFiles,
		ProcessedAt:  time.Now(),
	}, nil
}

// ValidateDeleteRequest 验证删除请求
func (s *DeleteService) ValidateDeleteRequest(req *DeleteRequest) error {
	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	if req.ObjectName == "" {
		return fmt.Errorf("对象名不能为空")
	}

	return nil
}

// ValidateBatchDeleteRequest 验证批量删除请求
func (s *DeleteService) ValidateBatchDeleteRequest(req *BatchDeleteRequest) error {
	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	if len(req.ObjectNames) == 0 {
		return fmt.Errorf("对象名列表不能为空")
	}

	if len(req.ObjectNames) > s.maxBatchSize {
		return fmt.Errorf("一次最多删除%d个文件", s.maxBatchSize)
	}

	// 检查对象名是否有效
	for i, objectName := range req.ObjectNames {
		if objectName == "" {
			return fmt.Errorf("对象名不能为空 (索引: %d)", i)
		}
	}

	return nil
}

// validatePrefixDeleteRequest 验证前缀删除请求
func (s *DeleteService) validatePrefixDeleteRequest(req *PrefixDeleteRequest) error {
	if req.BucketName == "" {
		return fmt.Errorf("存储桶名不能为空")
	}

	if req.Prefix == "" {
		return fmt.Errorf("文件前缀不能为空")
	}

	return nil
}
