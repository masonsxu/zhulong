package video

import (
	"fmt"
)

// SizeLimitManager 文件大小限制管理器
type SizeLimitManager struct {
	maxFileSize   int64            // 全局最大文件大小
	minFileSize   int64            // 全局最小文件大小
	formatLimits  map[string]int64 // 按格式的大小限制
}

// SizeLimits 大小限制信息
type SizeLimits struct {
	MaxFileSize          int64  `json:"max_file_size"`           // 最大文件大小（字节）
	MinFileSize          int64  `json:"min_file_size"`           // 最小文件大小（字节）
	MaxFileSizeFormatted string `json:"max_file_size_formatted"` // 格式化的最大文件大小
	MinFileSizeFormatted string `json:"min_file_size_formatted"` // 格式化的最小文件大小
}

// NewSizeLimitManager 创建文件大小限制管理器
func NewSizeLimitManager() *SizeLimitManager {
	return &SizeLimitManager{
		maxFileSize:  2 * 1024 * 1024 * 1024, // 2GB
		minFileSize:  1,                      // 1字节
		formatLimits: make(map[string]int64),
	}
}

// GetMaxFileSize 获取最大文件大小（字节）
func (s *SizeLimitManager) GetMaxFileSize() int64 {
	return s.maxFileSize
}

// GetMinFileSize 获取最小文件大小（字节）
func (s *SizeLimitManager) GetMinFileSize() int64 {
	return s.minFileSize
}

// SetMaxFileSize 设置最大文件大小
func (s *SizeLimitManager) SetMaxFileSize(size int64) {
	if size > 0 {
		s.maxFileSize = size
	}
}

// SetMinFileSize 设置最小文件大小
func (s *SizeLimitManager) SetMinFileSize(size int64) {
	if size >= 0 {
		s.minFileSize = size
	}
}

// ValidateSize 验证文件大小
func (s *SizeLimitManager) ValidateSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("文件大小无效：%d", size)
	}

	if size < s.minFileSize {
		return fmt.Errorf("文件不能为空")
	}

	if size > s.maxFileSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %s，当前文件 %s", 
			s.FormatSize(s.maxFileSize), s.FormatSize(size))
	}

	return nil
}

// ValidateSizeForFormat 针对特定格式验证文件大小
func (s *SizeLimitManager) ValidateSizeForFormat(format string, size int64) error {
	// 先进行基本验证
	if size < 0 {
		return fmt.Errorf("文件大小无效：%d", size)
	}

	if size < s.minFileSize {
		return fmt.Errorf("文件不能为空")
	}

	// 检查格式特定的限制
	if formatLimit, exists := s.formatLimits[format]; exists {
		if size > formatLimit {
			return fmt.Errorf("%s格式文件大小超过限制，最大允许 %s，当前文件 %s", 
				format, s.FormatSize(formatLimit), s.FormatSize(size))
		}
	} else {
		// 使用全局限制
		if size > s.maxFileSize {
			return fmt.Errorf("文件大小超过限制，最大允许 %s，当前文件 %s", 
				s.FormatSize(s.maxFileSize), s.FormatSize(size))
		}
	}

	return nil
}

// GetMaxFileSizeInMB 获取以MB为单位的最大文件大小
func (s *SizeLimitManager) GetMaxFileSizeInMB() int64 {
	return s.maxFileSize / (1024 * 1024)
}

// GetMaxFileSizeInKB 获取以KB为单位的最大文件大小
func (s *SizeLimitManager) GetMaxFileSizeInKB() int64 {
	return s.maxFileSize / 1024
}

// FormatSize 格式化文件大小显示
func (s *SizeLimitManager) FormatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// GetLimits 获取当前的大小限制信息
func (s *SizeLimitManager) GetLimits() *SizeLimits {
	return &SizeLimits{
		MaxFileSize:          s.maxFileSize,
		MinFileSize:          s.minFileSize,
		MaxFileSizeFormatted: s.FormatSize(s.maxFileSize),
		MinFileSizeFormatted: s.FormatSize(s.minFileSize),
	}
}

// UpdateLimits 更新大小限制
func (s *SizeLimitManager) UpdateLimits(limits *SizeLimits) error {
	if limits.MaxFileSize <= 0 {
		return fmt.Errorf("最大文件大小必须大于0")
	}

	if limits.MinFileSize < 0 {
		return fmt.Errorf("最小文件大小不能小于0")
	}

	if limits.MinFileSize >= limits.MaxFileSize {
		return fmt.Errorf("最小文件大小不能大于或等于最大文件大小")
	}

	s.maxFileSize = limits.MaxFileSize
	s.minFileSize = limits.MinFileSize

	return nil
}

// SetFormatLimits 设置按格式的大小限制
func (s *SizeLimitManager) SetFormatLimits(limits map[string]int64) {
	s.formatLimits = make(map[string]int64)
	for format, limit := range limits {
		if limit > 0 {
			s.formatLimits[format] = limit
		}
	}
}

// GetFormatLimit 获取特定格式的大小限制
func (s *SizeLimitManager) GetFormatLimit(format string) int64 {
	if limit, exists := s.formatLimits[format]; exists {
		return limit
	}
	return s.maxFileSize // 返回默认限制
}

// GetSupportedSizeRange 获取支持的文件大小范围
func (s *SizeLimitManager) GetSupportedSizeRange() (min, max int64) {
	return s.minFileSize, s.maxFileSize
}

// IsValidSize 检查文件大小是否在有效范围内
func (s *SizeLimitManager) IsValidSize(size int64) bool {
	return size >= s.minFileSize && size <= s.maxFileSize
}

// GetSizePercentage 获取文件大小相对于最大限制的百分比
func (s *SizeLimitManager) GetSizePercentage(size int64) float64 {
	if s.maxFileSize == 0 {
		return 0.0
	}
	percentage := float64(size) / float64(s.maxFileSize) * 100
	if percentage > 100 {
		return 100.0
	}
	return percentage
}

// GetRemainingSpace 获取剩余可用空间
func (s *SizeLimitManager) GetRemainingSpace(currentSize int64) int64 {
	remaining := s.maxFileSize - currentSize
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ValidateBatchSizes 验证批量文件的总大小
func (s *SizeLimitManager) ValidateBatchSizes(sizes []int64) error {
	var totalSize int64
	
	for i, size := range sizes {
		// 验证单个文件大小
		if err := s.ValidateSize(size); err != nil {
			return fmt.Errorf("文件 %d: %v", i+1, err)
		}
		
		totalSize += size
	}

	// 可以在这里添加批量上传的总大小限制
	batchLimit := s.maxFileSize * 10 // 例如：批量上传总大小不超过单文件限制的10倍
	if totalSize > batchLimit {
		return fmt.Errorf("批量上传总大小超过限制，最大允许 %s，当前总大小 %s", 
			s.FormatSize(batchLimit), s.FormatSize(totalSize))
	}

	return nil
}