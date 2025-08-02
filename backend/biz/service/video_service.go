package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	api "github.com/manteia/zhulong/biz/model/zhulong/api"
	"github.com/manteia/zhulong/pkg/config"
	"github.com/manteia/zhulong/pkg/metadata"
	"github.com/manteia/zhulong/pkg/storage"
	"github.com/manteia/zhulong/pkg/upload"
	"github.com/manteia/zhulong/pkg/video"
)

// VideoService 视频服务
type VideoService struct {
	config            *config.Config
	storageClient     storage.StorageInterface
	uploadService     *upload.UploadService
	metadataService   *metadata.MetadataService
	videoValidator    *video.VideoValidator
	videoExtractor    *video.VideoInfoExtractor
	thumbnailGenerator *video.ThumbnailGenerator
	sizeLimitManager  *video.SizeLimitManager
}

// NewVideoService 创建视频服务
func NewVideoService() (*VideoService, error) {
	// 加载配置
	cfg, err := config.LoadFromFile("../config/development.yml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 初始化存储客户端 
	storageClient, err := storage.NewMinIOStorage(&storage.MinIOConfig{
		Endpoint:  cfg.MinIO.Endpoint,
		AccessKey: cfg.MinIO.AccessKey,
		SecretKey: cfg.MinIO.SecretKey,
		UseSSL:    cfg.MinIO.UseSSL,
		Region:    cfg.MinIO.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化存储客户端失败: %v", err)
	}

	// 初始化各种服务
	uploadService := upload.NewUploadService(storageClient)
	metadataService := metadata.NewMetadataService()
	videoValidator := video.NewVideoValidator()
	videoExtractor := video.NewVideoInfoExtractor()
	thumbnailGenerator := video.NewThumbnailGenerator()
	sizeLimitManager := video.NewSizeLimitManager()

	return &VideoService{
		config:            cfg,
		storageClient:     storageClient,
		uploadService:     uploadService,
		metadataService:   metadataService,
		videoValidator:    videoValidator,
		videoExtractor:    videoExtractor,
		thumbnailGenerator: thumbnailGenerator,
		sizeLimitManager:  sizeLimitManager,
	}, nil
}

// UploadVideo 上传视频
func (s *VideoService) UploadVideo(ctx context.Context, req *api.VideoUploadRequest, fileHeader *multipart.FileHeader) (*api.VideoUploadResponse, error) {
	// 生成视频ID
	videoID := uuid.New().String()

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return s.errorResponse(1001, "无法打开上传文件"), nil
	}
	defer file.Close()

	// 读取文件数据进行验证
	fileData := make([]byte, fileHeader.Size)
	_, err = file.Read(fileData)
	if err != nil {
		return s.errorResponse(1002, "读取文件数据失败"), nil
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 验证文件大小
	if err := s.sizeLimitManager.ValidateSize(fileHeader.Size); err != nil {
		return s.errorResponse(1003, fmt.Sprintf("文件大小验证失败: %v", err)), nil
	}

	// 验证文件格式
	validationRequest := &video.ValidationRequest{
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Data:        fileData[:min(len(fileData), 512)], // 只取前512字节用于验证
	}

	validationResult, err := s.videoValidator.ValidateFormat(validationRequest)
	if err != nil {
		return s.errorResponse(1004, fmt.Sprintf("文件格式验证失败: %v", err)), nil
	}

	if !validationResult.IsValid {
		return s.errorResponse(1005, fmt.Sprintf("不支持的文件格式: %s", validationResult.ErrorMessage)), nil
	}

	// 提取视频信息
	infoRequest := &video.InfoExtractionRequest{
		Data:     fileData[:min(len(fileData), 1024*1024)], // 取前1MB用于信息提取
		Filename: fileHeader.Filename,
	}

	videoInfo, err := s.videoExtractor.ExtractInfo(infoRequest)
	if err != nil {
		// 信息提取失败不阻断上传，使用默认值
		videoInfo = &video.VideoInfo{
			Filename: fileHeader.Filename,
			Format:   validationResult.DetectedFormat,
			FileSize: fileHeader.Size,
		}
	}

	// 生成存储路径
	now := time.Now()
	objectName := fmt.Sprintf("videos/%d/%02d/%s%s",
		now.Year(), now.Month(), videoID, filepath.Ext(fileHeader.Filename))

	// 上传文件到存储
	uploadRequest := &upload.UploadRequest{
		BucketName:  "zhulong-videos", // 暂时硬编码，后续从配置获取
		FileName:    objectName,
		Reader:      file,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
	}

	_, err = s.uploadService.UploadFile(ctx, uploadRequest)
	if err != nil {
		return s.errorResponse(1006, fmt.Sprintf("文件上传失败: %v", err)), nil
	}

	// 生成缩略图
	thumbnailPath := ""
	thumbnailRequest := &video.ThumbnailRequest{
		VideoData: fileData,
		Options: &video.ThumbnailOptions{
			Width:      320,
			Height:     240,
			Quality:    80,
			Format:     "jpeg",
			TimeOffset: 0.0,
		},
	}

	thumbnailResult, err := s.thumbnailGenerator.GenerateFromVideo(thumbnailRequest)
	if err == nil && thumbnailResult != nil {
		// 上传缩略图
		thumbnailObjectName := fmt.Sprintf("thumbnails/%d/%02d/%s.jpg", now.Year(), now.Month(), videoID)
		thumbnailUploadRequest := &upload.UploadRequest{
			BucketName:  "zhulong-videos",
			FileName:    thumbnailObjectName,
			Reader:      bytes.NewReader(thumbnailResult.ImageData),
			Size:        thumbnailResult.FileSize,
			ContentType: "image/jpeg",
		}

		_, thumbnailUploadErr := s.uploadService.UploadFile(ctx, thumbnailUploadRequest)
		if thumbnailUploadErr == nil {
			thumbnailPath = thumbnailObjectName
		}
	}

	// 保存元数据
	metadataRequest := &metadata.FileMetadata{
		FileID:      videoID,
		BucketName:  "zhulong-videos",
		ObjectName:  objectName,
		FileName:    fileHeader.Filename,
		Title:       getValueOrDefaultFromString(req.Title, fileHeader.Filename),
		Description: getValueOrDefaultFromString(req.Description, ""),
		ContentType: fileHeader.Header.Get("Content-Type"),
		FileSize:    fileHeader.Size,
		Duration:    int64(videoInfo.Duration.Seconds()),
		Resolution:  fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height),
		Thumbnail:   thumbnailPath,
		Tags:        []string{},
		CreatedBy:   "system", // 暂时使用system，后续可以从上下文中获取用户信息
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.metadataService.SaveMetadata(ctx, metadataRequest)
	if err != nil {
		// 元数据保存失败，但不影响上传流程，记录日志即可
		fmt.Printf("保存元数据失败: %v\n", err)
	}

	// 构造响应
	videoResponse := &api.Video{
		ID:            videoID,
		Title:         getValueOrDefaultFromString(req.Title, fileHeader.Filename),
		Filename:      fileHeader.Filename,
		ContentType:   fileHeader.Header.Get("Content-Type"),
		Size:          fileHeader.Size,
		Duration:      int64(videoInfo.Duration.Seconds()),
		Width:         int32(videoInfo.Width),
		Height:        int32(videoInfo.Height),
		StoragePath:   objectName,
		ThumbnailPath: thumbnailPath,
		UploadedAt:    time.Now().UnixMilli(),
		UpdatedAt:     time.Now().UnixMilli(),
	}

	return &api.VideoUploadResponse{
		Base: &api.BaseResponse{
			Code:    0,
			Message: "上传成功",
		},
		Video: videoResponse,
	}, nil
}

// errorResponse 创建错误响应
func (s *VideoService) errorResponse(code int32, message string) *api.VideoUploadResponse {
	return &api.VideoUploadResponse{
		Base: &api.BaseResponse{
			Code:    code,
			Message: message,
		},
	}
}

// getValueOrDefault 获取值或默认值
func getValueOrDefault(ptr *string, defaultValue string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return defaultValue
}

// getValueOrDefaultFromString 从字符串获取值或默认值
func getValueOrDefaultFromString(value string, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetVideoList 获取视频列表
func (s *VideoService) GetVideoList(ctx context.Context, req *api.VideoListRequest) (*api.VideoListResponse, error) {
	// 参数验证
	if err := s.validateVideoListRequest(req); err != nil {
		return s.videoListErrorResponse(2001, err.Error()), nil
	}

	// 设置默认值
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	// 构建查询参数
	listRequest := &metadata.ListMetadataRequest{
		Offset: int((page - 1) * pageSize),
		Limit:  int(pageSize),
		SortBy: req.SortBy,
		Order:  "desc", // 默认降序
	}

	// 根据请求设置排序方向
	if req.SortBy != "" {
		// 默认降序，如果没有特别指定
		listRequest.Order = "desc"
	}

	// 如果没有指定排序字段，默认按创建时间排序
	if listRequest.SortBy == "" {
		listRequest.SortBy = "created_at"
	}

	// 查询数据
	listResponse, err := s.metadataService.ListMetadata(ctx, listRequest)
	if err != nil {
		return s.videoListErrorResponse(2002, fmt.Sprintf("查询视频列表失败: %v", err)), nil
	}

	// 转换为API响应格式
	var videos []*api.Video
	for _, metadata := range listResponse.Items {
		video := &api.Video{
			ID:            metadata.FileID,
			Title:         metadata.Title,
			Filename:      metadata.FileName,
			ContentType:   metadata.ContentType,
			Size:          metadata.FileSize,
			Duration:      metadata.Duration,
			Width:         0, // 从分辨率字符串解析
			Height:        0, // 从分辨率字符串解析
			StoragePath:   metadata.ObjectName,
			ThumbnailPath: metadata.Thumbnail,
			UploadedAt:    metadata.CreatedAt.UnixMilli(),
			UpdatedAt:     metadata.UpdatedAt.UnixMilli(),
		}

		// 解析分辨率
		if metadata.Resolution != "" {
			fmt.Sscanf(metadata.Resolution, "%dx%d", &video.Width, &video.Height)
		}

		videos = append(videos, video)
	}

	return &api.VideoListResponse{
		Base: &api.BaseResponse{
			Code:    0,
			Message: "获取成功",
		},
		Videos: videos,
		Total:  int32(listResponse.Total),
	}, nil
}

// validateVideoListRequest 验证视频列表请求
func (s *VideoService) validateVideoListRequest(req *api.VideoListRequest) error {
	if req.Page < 0 {
		return fmt.Errorf("页码必须大于等于0")
	}
	if req.PageSize < 0 {
		return fmt.Errorf("页面大小必须大于等于0")
	}
	if req.PageSize > 100 {
		return fmt.Errorf("页面大小不能超过100")
	}
	return nil
}

// videoListErrorResponse 创建视频列表错误响应
func (s *VideoService) videoListErrorResponse(code int32, message string) *api.VideoListResponse {
	return &api.VideoListResponse{
		Base: &api.BaseResponse{
			Code:    code,
			Message: message,
		},
		Videos: []*api.Video{},
		Total:  0,
	}
}

// GetVideoDetail 获取视频详情
func (s *VideoService) GetVideoDetail(ctx context.Context, req *api.VideoDetailRequest) (*api.VideoDetailResponse, error) {
	// 参数验证
	if err := s.validateVideoDetailRequest(req); err != nil {
		return s.videoDetailErrorResponse(3000, err.Error()), nil
	}

	// 查询视频元数据
	metadata, err := s.metadataService.GetMetadata(ctx, req.VideoID)
	if err != nil {
		return s.videoDetailErrorResponse(3001, "视频不存在"), nil
	}

	// 转换为API响应格式
	video := s.convertMetadataToVideo(metadata)

	return &api.VideoDetailResponse{
		Base: &api.BaseResponse{
			Code:    0,
			Message: "获取成功",
		},
		Video: video,
	}, nil
}

// validateVideoDetailRequest 验证视频详情请求
func (s *VideoService) validateVideoDetailRequest(req *api.VideoDetailRequest) error {
	if req.VideoID == "" {
		return fmt.Errorf("视频ID不能为空")
	}

	// 验证视频ID格式（去除空格）
	videoID := strings.TrimSpace(req.VideoID)
	if videoID == "" {
		return fmt.Errorf("视频ID不能为空或只包含空格")
	}

	// 更新请求中的videoID（去除空格）
	req.VideoID = videoID

	return nil
}

// convertMetadataToVideo 将元数据转换为Video结构
func (s *VideoService) convertMetadataToVideo(metadata *metadata.FileMetadata) *api.Video {
	video := &api.Video{
		ID:            metadata.FileID,
		Title:         metadata.Title,
		Filename:      metadata.FileName,
		ContentType:   metadata.ContentType,
		Size:          metadata.FileSize,
		Duration:      metadata.Duration,
		Width:         0, // 默认值
		Height:        0, // 默认值
		StoragePath:   metadata.ObjectName,
		ThumbnailPath: metadata.Thumbnail,
		UploadedAt:    metadata.CreatedAt.UnixMilli(),
		UpdatedAt:     metadata.UpdatedAt.UnixMilli(),
	}

	// 解析分辨率
	if metadata.Resolution != "" {
		s.parseResolution(metadata.Resolution, video)
	}

	return video
}

// parseResolution 解析分辨率字符串
func (s *VideoService) parseResolution(resolution string, video *api.Video) {
	// 尝试解析分辨率格式：如 "1920x1080"
	var width, height int32
	n, err := fmt.Sscanf(resolution, "%dx%d", &width, &height)
	if err == nil && n == 2 && width > 0 && height > 0 {
		video.Width = width
		video.Height = height
	}
	// 解析失败时保持默认值0
}

// videoDetailErrorResponse 创建视频详情错误响应
func (s *VideoService) videoDetailErrorResponse(code int32, message string) *api.VideoDetailResponse {
	return &api.VideoDetailResponse{
		Base: &api.BaseResponse{
			Code:    code,
			Message: message,
		},
		Video: nil,
	}
}

// GetVideoPlayURL 获取视频播放URL
func (s *VideoService) GetVideoPlayURL(ctx context.Context, req *api.VideoPlayURLRequest) (*api.VideoPlayURLResponse, error) {
	// 参数验证
	if err := s.validateVideoPlayURLRequest(req); err != nil {
		return s.videoPlayURLErrorResponse(4000, err.Error()), nil
	}

	// 查询视频元数据确认视频存在
	metadata, err := s.metadataService.GetMetadata(ctx, req.VideoID)
	if err != nil {
		return s.videoPlayURLErrorResponse(4001, "视频不存在"), nil
	}

	// 设置过期时间
	expireSeconds := req.ExpireSeconds
	if expireSeconds == 0 {
		expireSeconds = 3600 // 默认1小时
	}

	// 生成预签名URL
	expiry := time.Duration(expireSeconds) * time.Second
	playURL, err := s.storageClient.GetPresignedURL(ctx, metadata.BucketName, metadata.ObjectName, expiry)
	if err != nil {
		return s.videoPlayURLErrorResponse(5000, fmt.Sprintf("生成播放URL失败: %v", err)), nil
	}

	// 计算过期时间戳
	expiresAt := time.Now().Add(expiry).UnixMilli()

	return &api.VideoPlayURLResponse{
		Base: &api.BaseResponse{
			Code:    0,
			Message: "获取成功",
		},
		PlayURL:   playURL,
		ExpiresAt: expiresAt,
	}, nil
}

// validateVideoPlayURLRequest 验证获取播放URL请求
func (s *VideoService) validateVideoPlayURLRequest(req *api.VideoPlayURLRequest) error {
	if req.VideoID == "" {
		return fmt.Errorf("视频ID不能为空")
	}

	// 验证视频ID格式（去除空格）
	videoID := strings.TrimSpace(req.VideoID)
	if videoID == "" {
		return fmt.Errorf("视频ID不能为空或只包含空格")
	}

	// 更新请求中的videoID（去除空格）
	req.VideoID = videoID

	// 验证过期时间
	if req.ExpireSeconds < 0 {
		return fmt.Errorf("过期时间不能为负数")
	}

	// 最大过期时间限制为7天
	maxExpireSeconds := int32(7 * 24 * 3600) // 7天
	if req.ExpireSeconds > maxExpireSeconds {
		return fmt.Errorf("过期时间不能超过7天")
	}

	return nil
}

// videoPlayURLErrorResponse 创建播放URL错误响应
func (s *VideoService) videoPlayURLErrorResponse(code int32, message string) *api.VideoPlayURLResponse {
	return &api.VideoPlayURLResponse{
		Base: &api.BaseResponse{
			Code:    code,
			Message: message,
		},
		PlayURL:   "",
		ExpiresAt: 0,
	}
}