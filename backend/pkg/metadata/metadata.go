package metadata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manteia/zhulong/biz/model/db"
	"gorm.io/gorm"
)

// MetadataService 文件元数据管理服务
type MetadataService struct {
	db *gorm.DB
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
func NewMetadataService(database *gorm.DB) (*MetadataService, error) {
	// 自动迁移数据库表
	err := database.AutoMigrate(&db.VideoMetadata{})
	if err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return &MetadataService{
		db: database,
	}, nil
}

// SaveMetadata 保存文件元数据
func (s *MetadataService) SaveMetadata(ctx context.Context, metadata *FileMetadata) error {
	// 验证元数据
	if err := s.ValidateMetadata(metadata); err != nil {
		return err
	}

	dbMetadata := toDBMetadata(metadata)

	// 设置时间戳
	now := time.Now()
	if dbMetadata.CreatedAt.IsZero() {
		dbMetadata.CreatedAt = now
	}
	dbMetadata.UpdatedAt = now

	// 保存到数据库
	result := s.db.WithContext(ctx).Create(dbMetadata)
	if result.Error != nil {
		return fmt.Errorf("保存元数据失败: %w", result.Error)
	}

	return nil
}

// GetMetadata 获取文件元数据
func (s *MetadataService) GetMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	var dbMetadata db.VideoMetadata
	result := s.db.WithContext(ctx).Where("file_id = ?", fileID).First(&dbMetadata)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("元数据不存在: %s", fileID)
		}
		return nil, fmt.Errorf("查询元数据失败: %w", result.Error)
	}

	return fromDBMetadata(&dbMetadata), nil
}

// DeleteMetadata 删除文件元数据
func (s *MetadataService) DeleteMetadata(ctx context.Context, fileID string) error {
	result := s.db.WithContext(ctx).Where("file_id = ?", fileID).Delete(&db.VideoMetadata{})
	if result.Error != nil {
		return fmt.Errorf("删除元数据失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("元数据不存在: %s", fileID)
	}
	return nil
}

// ListMetadata 列出文件元数据
func (s *MetadataService) ListMetadata(ctx context.Context, req *ListMetadataRequest) (*ListMetadataResponse, error) {
	var dbMetadatas []db.VideoMetadata
	var total int64

	db := s.db.WithContext(ctx).Model(&db.VideoMetadata{})

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询元数据总数失败: %w", err)
	}

	// 应用排序
	order := "desc"
	if req.Order == "asc" {
		order = "asc"
	}
	db = db.Order(fmt.Sprintf("%s %s", req.SortBy, order))

	// 应用分页
	db = db.Offset(req.Offset).Limit(req.Limit)

	// 查询数据
	if err := db.Find(&dbMetadatas).Error; err != nil {
		return nil, fmt.Errorf("查询元数据列表失败: %w", err)
	}

	// 转换为FileMetadata
	var items []*FileMetadata
	for _, dbm := range dbMetadatas {
		items = append(items, fromDBMetadata(&dbm))
	}

	return &ListMetadataResponse{
		Items: items,
		Total: int(total),
	}, nil
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

// toDBMetadata 将FileMetadata转换为db.VideoMetadata
func toDBMetadata(fm *FileMetadata) *db.VideoMetadata {
	return &db.VideoMetadata{
		FileID:      fm.FileID,
		BucketName:  fm.BucketName,
		ObjectName:  fm.ObjectName,
		FileName:    fm.FileName,
		Title:       fm.Title,
		Description: fm.Description,
		ContentType: fm.ContentType,
		FileSize:    fm.FileSize,
		Duration:    fm.Duration,
		Resolution:  fm.Resolution,
		Thumbnail:   fm.Thumbnail,
		Tags:        strings.Join(fm.Tags, ","),
		CreatedBy:   fm.CreatedBy,
		UploadedAt:  fm.CreatedAt,
	}
}

// fromDBMetadata 将db.VideoMetadata转换为FileMetadata
func fromDBMetadata(dbm *db.VideoMetadata) *FileMetadata {
	return &FileMetadata{
		FileID:      dbm.FileID,
		BucketName:  dbm.BucketName,
		ObjectName:  dbm.ObjectName,
		FileName:    dbm.FileName,
		Title:       dbm.Title,
		Description: dbm.Description,
		ContentType: dbm.ContentType,
		FileSize:    dbm.FileSize,
		Duration:    dbm.Duration,
		Resolution:  dbm.Resolution,
		Thumbnail:   dbm.Thumbnail,
		Tags:        strings.Split(dbm.Tags, ","),
		CreatedBy:   dbm.CreatedBy,
		CreatedAt:   dbm.UploadedAt,
		UpdatedAt:   dbm.UpdatedAt,
	}
}
