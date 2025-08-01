package models

import (
	"time"
)

// Video 视频模型
type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Duration    int64     `json:"duration"` // 秒
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	StoragePath string    `json:"storage_path"`
	ThumbnailPath string  `json:"thumbnail_path,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VideoUploadRequest 视频上传请求
type VideoUploadRequest struct {
	Title string `json:"title" binding:"required"`
}

// VideoListResponse 视频列表响应
type VideoListResponse struct {
	Videos []Video `json:"videos"`
	Total  int     `json:"total"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}