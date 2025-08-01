package metadata

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MetadataService 文件元数据管理服务
type MetadataService struct {
	// 使用内存存储作为简单实现，实际项目中应该使用数据库
	storage map[string]*FileMetadata
	mutex   sync.RWMutex
}

// FileMetadata 文件元数据结构
type FileMetadata struct {
	FileID      string    `json:"file_id"`      // 文件唯一标识
	BucketName  string    `json:"bucket_name"`  // 存储桶名
	ObjectName  string    `json:"object_name"`  // 对象名（存储路径）
	FileName    string    `json:"file_name"`    // 原始文件名
	FileSize    int64     `json:"file_size"`    // 文件大小（字节）
	ContentType string    `json:"content_type"` // 文件类型
	Title       string    `json:"title"`        // 文件标题
	Description string    `json:"description"`  // 文件描述
	Tags        []string  `json:"tags"`         // 文件标签
	Duration    int64     `json:"duration"`     // 视频时长（秒）
	Resolution  string    `json:"resolution"`   // 分辨率
	Bitrate     int64     `json:"bitrate"`      // 比特率
	Thumbnail   string    `json:"thumbnail"`    // 缩略图路径
	CreatedBy   string    `json:"created_by"`   // 创建者
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
}

// UpdateMetadataRequest 更新元数据请求
type UpdateMetadataRequest struct {
	FileID      string    `json:"file_id"`      // 文件ID
	Title       *string   `json:"title"`        // 标题（可选）
	Description *string   `json:"description"`  // 描述（可选）
	Tags        *[]string `json:"tags"`         // 标签（可选）
	Duration    *int64    `json:"duration"`     // 时长（可选）
	Resolution  *string   `json:"resolution"`   // 分辨率（可选）
	Bitrate     *int64    `json:"bitrate"`      // 比特率（可选）
	Thumbnail   *string   `json:"thumbnail"`    // 缩略图（可选）
}

// SearchMetadataRequest 搜索元数据请求
type SearchMetadataRequest struct {
	Query     string   `json:"query"`      // 搜索关键词（标题、描述）
	Tags      []string `json:"tags"`       // 标签过滤
	CreatedBy string   `json:"created_by"` // 创建者过滤
	Limit     int      `json:"limit"`      // 返回数量限制
	Offset    int      `json:"offset"`     // 偏移量
}

// SearchMetadataResponse 搜索元数据响应
type SearchMetadataResponse struct {
	Items []*FileMetadata `json:"items"` // 搜索结果
	Total int             `json:"total"` // 总数
}

// ListMetadataRequest 列表元数据请求
type ListMetadataRequest struct {
	Offset int    `json:"offset"` // 偏移量
	Limit  int    `json:"limit"`  // 数量限制
	SortBy string `json:"sort_by"` // 排序字段
	Order  string `json:"order"`  // 排序方向 (asc/desc)
}

// ListMetadataResponse 列表元数据响应
type ListMetadataResponse struct {
	Items []*FileMetadata `json:"items"` // 列表结果
	Total int             `json:"total"` // 总数
}

// NewMetadataService 创建元数据服务
func NewMetadataService() *MetadataService {
	return &MetadataService{
		storage: make(map[string]*FileMetadata),
		mutex:   sync.RWMutex{},
	}
}

// SaveMetadata 保存文件元数据
func (s *MetadataService) SaveMetadata(ctx context.Context, metadata *FileMetadata) error {
	// 验证元数据
	if err := s.ValidateMetadata(metadata); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 设置时间戳
	now := time.Now()
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = now
	}
	metadata.UpdatedAt = now

	// 去重标签
	metadata.Tags = s.deduplicateTags(metadata.Tags)

	// 保存到存储
	s.storage[metadata.FileID] = metadata

	return nil
}

// GetMetadata 获取文件元数据
func (s *MetadataService) GetMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	metadata, exists := s.storage[fileID]
	if !exists {
		return nil, fmt.Errorf("元数据不存在: %s", fileID)
	}

	// 返回副本以避免并发修改
	return s.copyMetadata(metadata), nil
}

// UpdateMetadata 更新文件元数据
func (s *MetadataService) UpdateMetadata(ctx context.Context, req *UpdateMetadataRequest) error {
	if req.FileID == "" {
		return fmt.Errorf("文件ID不能为空")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	metadata, exists := s.storage[req.FileID]
	if !exists {
		return fmt.Errorf("元数据不存在: %s", req.FileID)
	}

	// 更新字段
	if req.Title != nil {
		metadata.Title = *req.Title
	}
	if req.Description != nil {
		metadata.Description = *req.Description
	}
	if req.Tags != nil {
		metadata.Tags = s.deduplicateTags(*req.Tags)
	}
	if req.Duration != nil {
		metadata.Duration = *req.Duration
	}
	if req.Resolution != nil {
		metadata.Resolution = *req.Resolution
	}
	if req.Bitrate != nil {
		metadata.Bitrate = *req.Bitrate
	}
	if req.Thumbnail != nil {
		metadata.Thumbnail = *req.Thumbnail
	}

	// 更新时间戳
	metadata.UpdatedAt = time.Now()

	return nil
}

// DeleteMetadata 删除文件元数据
func (s *MetadataService) DeleteMetadata(ctx context.Context, fileID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.storage[fileID]; !exists {
		return fmt.Errorf("元数据不存在: %s", fileID)
	}

	delete(s.storage, fileID)
	return nil
}

// GetMetadataByObjectName 根据对象名获取元数据
func (s *MetadataService) GetMetadataByObjectName(ctx context.Context, bucketName, objectName string) (*FileMetadata, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, metadata := range s.storage {
		if metadata.BucketName == bucketName && metadata.ObjectName == objectName {
			return s.copyMetadata(metadata), nil
		}
	}

	return nil, fmt.Errorf("未找到对象的元数据: %s/%s", bucketName, objectName)
}

// SearchMetadata 搜索文件元数据
func (s *MetadataService) SearchMetadata(ctx context.Context, req *SearchMetadataRequest) (*SearchMetadataResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var matches []*FileMetadata

	// 遍历所有元数据进行匹配
	for _, metadata := range s.storage {
		if s.matchesSearchCriteria(metadata, req) {
			matches = append(matches, s.copyMetadata(metadata))
		}
	}

	// 应用偏移和限制
	total := len(matches)
	start := req.Offset
	if start > total {
		start = total
	}

	end := start + req.Limit
	if end > total {
		end = total
	}

	if start >= total {
		matches = []*FileMetadata{}
	} else {
		matches = matches[start:end]
	}

	return &SearchMetadataResponse{
		Items: matches,
		Total: total,
	}, nil
}

// ListMetadata 列出文件元数据
func (s *MetadataService) ListMetadata(ctx context.Context, req *ListMetadataRequest) (*ListMetadataResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 获取所有元数据
	var items []*FileMetadata
	for _, metadata := range s.storage {
		items = append(items, s.copyMetadata(metadata))
	}

	// 排序
	s.sortMetadata(items, req.SortBy, req.Order)

	// 应用分页
	total := len(items)
	start := req.Offset
	if start > total {
		start = total
	}

	end := start + req.Limit
	if end > total {
		end = total
	}

	if start >= total {
		items = []*FileMetadata{}
	} else {
		items = items[start:end]
	}

	return &ListMetadataResponse{
		Items: items,
		Total: total,
	}, nil
}

// AddTags 添加标签
func (s *MetadataService) AddTags(ctx context.Context, fileID string, tags []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metadata, exists := s.storage[fileID]
	if !exists {
		return fmt.Errorf("元数据不存在: %s", fileID)
	}

	// 合并现有标签和新标签
	allTags := append(metadata.Tags, tags...)
	metadata.Tags = s.deduplicateTags(allTags)
	metadata.UpdatedAt = time.Now()

	return nil
}

// RemoveTags 移除标签
func (s *MetadataService) RemoveTags(ctx context.Context, fileID string, tags []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metadata, exists := s.storage[fileID]
	if !exists {
		return fmt.Errorf("元数据不存在: %s", fileID)
	}

	// 创建要移除的标签映射
	toRemove := make(map[string]bool)
	for _, tag := range tags {
		toRemove[tag] = true
	}

	// 过滤掉要移除的标签
	var remainingTags []string
	for _, tag := range metadata.Tags {
		if !toRemove[tag] {
			remainingTags = append(remainingTags, tag)
		}
	}

	metadata.Tags = remainingTags
	metadata.UpdatedAt = time.Now()

	return nil
}

// ValidateMetadata 验证元数据
func (s *MetadataService) ValidateMetadata(metadata *FileMetadata) error {
	if metadata.FileID == "" {
		return fmt.Errorf("文件ID不能为空")
	}

	if metadata.Title == "" {
		return fmt.Errorf("标题不能为空")
	}

	if len(metadata.Title) > 255 {
		return fmt.Errorf("标题长度不能超过255个字符")
	}

	if metadata.CreatedBy == "" {
		return fmt.Errorf("创建者不能为空")
	}

	if len(metadata.Description) > 1000 {
		return fmt.Errorf("描述长度不能超过1000个字符")
	}

	return nil
}

// matchesSearchCriteria 检查元数据是否匹配搜索条件
func (s *MetadataService) matchesSearchCriteria(metadata *FileMetadata, req *SearchMetadataRequest) bool {
	// 检查查询关键词
	if req.Query != "" {
		query := strings.ToLower(req.Query)
		title := strings.ToLower(metadata.Title)
		description := strings.ToLower(metadata.Description)
		
		if !strings.Contains(title, query) && !strings.Contains(description, query) {
			return false
		}
	}

	// 检查标签
	if len(req.Tags) > 0 {
		metadataTags := make(map[string]bool)
		for _, tag := range metadata.Tags {
			metadataTags[strings.ToLower(tag)] = true
		}

		for _, reqTag := range req.Tags {
			if !metadataTags[strings.ToLower(reqTag)] {
				return false
			}
		}
	}

	// 检查创建者
	if req.CreatedBy != "" && metadata.CreatedBy != req.CreatedBy {
		return false
	}

	return true
}

// sortMetadata 排序元数据
func (s *MetadataService) sortMetadata(items []*FileMetadata, sortBy, order string) {
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			less = items[i].Title < items[j].Title
		case "duration":
			less = items[i].Duration < items[j].Duration
		case "file_size":
			less = items[i].FileSize < items[j].FileSize
		case "created_at":
			less = items[i].CreatedAt.Before(items[j].CreatedAt)
		case "updated_at":
			less = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		default:
			less = items[i].CreatedAt.Before(items[j].CreatedAt)
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

// copyMetadata 复制元数据以避免并发修改
func (s *MetadataService) copyMetadata(original *FileMetadata) *FileMetadata {
	copy := *original
	// 深拷贝标签切片
	if original.Tags != nil {
		copy.Tags = make([]string, len(original.Tags))
		copySlice := copy.Tags
		for i, tag := range original.Tags {
			copySlice[i] = tag
		}
	}
	return &copy
}

// deduplicateTags 去重标签
func (s *MetadataService) deduplicateTags(tags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		if tag != "" && !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}