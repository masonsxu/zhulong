package db

import (
	"time"

	"gorm.io/gorm"
)

// VideoMetadata 是视频元数据的GORM模型
type VideoMetadata struct {
	gorm.Model
	FileID      string    `gorm:"uniqueIndex;not null"`
	BucketName  string
	ObjectName  string
	FileName    string
	Title       string
	Description string
	ContentType string
	FileSize    int64
	Duration    int64
	Resolution  string
	Thumbnail   string
	Tags        string // 使用字符串存储标签，以逗号分隔
	CreatedBy   string
	UploadedAt  time.Time
}
