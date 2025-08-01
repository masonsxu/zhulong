package video

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// VideoInfoExtractor 视频信息提取器
type VideoInfoExtractor struct {
	validator *VideoValidator
}

// InfoExtractionRequest 信息提取请求
type InfoExtractionRequest struct {
	Data     []byte `json:"data"`     // 文件数据
	Filename string `json:"filename"` // 文件名
}

// VideoInfo 视频信息
type VideoInfo struct {
	// 基本信息
	Filename string `json:"filename"` // 文件名
	Format   string `json:"format"`   // 视频格式
	FileSize int64  `json:"file_size"` // 文件大小（字节）

	// 视频属性
	Duration  time.Duration `json:"duration"`   // 时长
	Width     int           `json:"width"`      // 宽度
	Height    int           `json:"height"`     // 高度
	Bitrate   int64         `json:"bitrate"`    // 比特率（bps）
	FrameRate float64       `json:"frame_rate"` // 帧率（fps）

	// 编码信息
	VideoCodec string `json:"video_codec"` // 视频编码
	AudioCodec string `json:"audio_codec"` // 音频编码

	// 格式化显示
	DurationFormatted   string `json:"duration_formatted"`   // 格式化时长
	ResolutionFormatted string `json:"resolution_formatted"` // 格式化分辨率
	FileSizeFormatted   string `json:"file_size_formatted"`  // 格式化文件大小
}

// NewVideoInfoExtractor 创建视频信息提取器
func NewVideoInfoExtractor() *VideoInfoExtractor {
	return &VideoInfoExtractor{
		validator: NewVideoValidator(),
	}
}

// ExtractInfo 提取视频信息
func (e *VideoInfoExtractor) ExtractInfo(request *InfoExtractionRequest) (*VideoInfo, error) {
	// 验证输入
	if len(request.Data) == 0 {
		return nil, fmt.Errorf("文件数据为空")
	}

	if len(request.Data) < 12 {
		return nil, fmt.Errorf("文件头信息不完整")
	}

	// 检测文件格式
	format, err := e.validator.DetectFormatByMagicNumber(request.Data)
	if err != nil {
		return nil, fmt.Errorf("无法识别的视频格式: %v", err)
	}

	// 创建基本信息
	info := &VideoInfo{
		Filename: request.Filename,
		Format:   format,
		FileSize: int64(len(request.Data)),
	}

	// 提取详细信息
	e.extractDetailedInfo(request.Data, format, info)

	// 生成格式化显示
	e.formatDisplayInfo(info)

	return info, nil
}

// extractDetailedInfo 提取详细信息
func (e *VideoInfoExtractor) extractDetailedInfo(data []byte, format string, info *VideoInfo) {
	switch format {
	case "mp4", "mov":
		e.extractMP4Info(data, info)
	case "avi":
		e.extractAVIInfo(data, info)
	case "webm":
		e.extractWebMInfo(data, info)
	}
}

// extractMP4Info 提取MP4信息
func (e *VideoInfoExtractor) extractMP4Info(data []byte, info *VideoInfo) {
	// 解析MP4 box结构
	offset := 0
	for offset < len(data)-8 {
		if offset+8 > len(data) {
			break
		}

		// 读取box头
		boxSize := binary.BigEndian.Uint32(data[offset : offset+4])
		boxType := string(data[offset+4 : offset+8])

		if boxSize == 0 || boxSize > uint32(len(data)-offset) {
			break
		}

		// 处理不同类型的box
		switch boxType {
		case "mvhd": // Movie header
			e.extractMovieHeader(data[offset:offset+int(boxSize)], info)
		case "tkhd": // Track header
			e.extractTrackHeader(data[offset:offset+int(boxSize)], info)
		case "stsd": // Sample description
			e.extractSampleDescription(data[offset:offset+int(boxSize)], info)
		}

		offset += int(boxSize)
	}
}

// extractAVIInfo 提取AVI信息
func (e *VideoInfoExtractor) extractAVIInfo(data []byte, info *VideoInfo) {
	// 简化的AVI信息提取
	// 查找avih chunk (AVI header)
	avihPos := bytes.Index(data, []byte("avih"))
	if avihPos != -1 && avihPos+56 <= len(data) {
		headerData := data[avihPos+8 : avihPos+56] // AVI header结构

		// 提取微秒每帧（frame rate）
		microSecPerFrame := binary.LittleEndian.Uint32(headerData[0:4])
		if microSecPerFrame > 0 {
			info.FrameRate = 1000000.0 / float64(microSecPerFrame)
		}

		// 提取总帧数
		totalFrames := binary.LittleEndian.Uint32(headerData[16:20])
		if totalFrames > 0 && info.FrameRate > 0 {
			info.Duration = time.Duration(float64(totalFrames)/info.FrameRate) * time.Second
		}

		// 提取分辨率
		info.Width = int(binary.LittleEndian.Uint32(headerData[32:36]))
		info.Height = int(binary.LittleEndian.Uint32(headerData[36:40]))
	}
}

// extractWebMInfo 提取WebM信息
func (e *VideoInfoExtractor) extractWebMInfo(data []byte, info *VideoInfo) {
	// 简化的WebM信息提取
	// WebM使用Matroska容器格式，这里实现基本的信息提取
	// 在实际项目中，可能需要更复杂的EBML解析

	// 查找关键元素标识符
	if e.findWebMElement(data, []byte{0x44, 0x89}) != -1 { // Duration element
		// 提取时长信息（简化版）
		info.Duration = 0 // 需要实际的EBML解析
	}

	if e.findWebMElement(data, []byte{0xB0}) != -1 { // PixelWidth
		// 提取宽度信息（简化版）
		info.Width = 0 // 需要实际的EBML解析
	}

	if e.findWebMElement(data, []byte{0xBA}) != -1 { // PixelHeight
		// 提取高度信息（简化版）
		info.Height = 0 // 需要实际的EBML解析
	}
}

// findWebMElement 查找WebM元素
func (e *VideoInfoExtractor) findWebMElement(data []byte, elementID []byte) int {
	return bytes.Index(data, elementID)
}

// extractMovieHeader 提取电影头信息
func (e *VideoInfoExtractor) extractMovieHeader(boxData []byte, info *VideoInfo) {
	if len(boxData) < 32 {
		return
	}

	// 跳过box头和版本/标志
	offset := 12

	// 时间刻度和时长
	timeScale := binary.BigEndian.Uint32(boxData[offset+8 : offset+12])
	duration := binary.BigEndian.Uint32(boxData[offset+12 : offset+16])

	if timeScale > 0 {
		info.Duration = time.Duration(duration) * time.Second / time.Duration(timeScale)
	}
}

// extractTrackHeader 提取轨道头信息
func (e *VideoInfoExtractor) extractTrackHeader(boxData []byte, info *VideoInfo) {
	if len(boxData) < 92 {
		return
	}

	// 提取宽度和高度（固定点数格式）
	widthFixed := binary.BigEndian.Uint32(boxData[len(boxData)-8 : len(boxData)-4])
	heightFixed := binary.BigEndian.Uint32(boxData[len(boxData)-4:])

	info.Width = int(widthFixed >> 16)   // 取整数部分
	info.Height = int(heightFixed >> 16) // 取整数部分
}

// extractSampleDescription 提取样本描述信息
func (e *VideoInfoExtractor) extractSampleDescription(boxData []byte, info *VideoInfo) {
	if len(boxData) < 16 {
		return
	}

	// 查找编解码器信息
	// 这里简化处理，实际需要解析完整的sample description
	if bytes.Contains(boxData, []byte("avc1")) {
		info.VideoCodec = "H.264"
	} else if bytes.Contains(boxData, []byte("hvc1")) {
		info.VideoCodec = "H.265"
	}

	if bytes.Contains(boxData, []byte("mp4a")) {
		info.AudioCodec = "AAC"
	}
}

// ExtractDuration 提取视频时长
func (e *VideoInfoExtractor) ExtractDuration(data []byte) (time.Duration, error) {
	if len(data) < 12 {
		return 0, fmt.Errorf("数据太短，无法提取时长")
	}

	format, err := e.validator.DetectFormatByMagicNumber(data)
	if err != nil {
		return 0, err
	}

	info := &VideoInfo{}
	e.extractDetailedInfo(data, format, info)

	return info.Duration, nil
}

// ExtractResolution 提取视频分辨率
func (e *VideoInfoExtractor) ExtractResolution(data []byte) (width, height int, err error) {
	if len(data) < 12 {
		return 0, 0, fmt.Errorf("数据太短，无法提取分辨率")
	}

	format, err := e.validator.DetectFormatByMagicNumber(data)
	if err != nil {
		return 0, 0, err
	}

	info := &VideoInfo{}
	e.extractDetailedInfo(data, format, info)

	return info.Width, info.Height, nil
}

// ExtractBitrate 提取视频比特率
func (e *VideoInfoExtractor) ExtractBitrate(data []byte) (int64, error) {
	if len(data) < 12 {
		return 0, fmt.Errorf("数据太短，无法提取比特率")
	}

	format, err := e.validator.DetectFormatByMagicNumber(data)
	if err != nil {
		return 0, err
	}

	info := &VideoInfo{}
	e.extractDetailedInfo(data, format, info)

	return info.Bitrate, nil
}

// ExtractFrameRate 提取视频帧率
func (e *VideoInfoExtractor) ExtractFrameRate(data []byte) (float64, error) {
	if len(data) < 12 {
		return 0, fmt.Errorf("数据太短，无法提取帧率")
	}

	format, err := e.validator.DetectFormatByMagicNumber(data)
	if err != nil {
		return 0, err
	}

	info := &VideoInfo{}
	e.extractDetailedInfo(data, format, info)

	return info.FrameRate, nil
}

// formatDisplayInfo 格式化显示信息
func (e *VideoInfoExtractor) formatDisplayInfo(info *VideoInfo) {
	// 格式化时长
	info.DurationFormatted = e.FormatDuration(info.Duration)

	// 格式化分辨率
	info.ResolutionFormatted = e.FormatResolution(info.Width, info.Height)

	// 格式化文件大小
	sizeLimitManager := NewSizeLimitManager()
	info.FileSizeFormatted = sizeLimitManager.FormatSize(info.FileSize)
}

// FormatDuration 格式化时长显示
func (e *VideoInfoExtractor) FormatDuration(duration time.Duration) string {
	if duration == 0 {
		return "00:00"
	}

	totalSeconds := int(duration.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// FormatResolution 格式化分辨率显示
func (e *VideoInfoExtractor) FormatResolution(width, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

// GetFileExtension 获取文件扩展名
func (e *VideoInfoExtractor) GetFileExtension(filename string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
}

// GetVideoCodecDescription 获取视频编码描述
func (e *VideoInfoExtractor) GetVideoCodecDescription(codec string) string {
	descriptions := map[string]string{
		"H.264": "高效视频编码，广泛兼容",
		"H.265": "新一代高效视频编码，更小文件",
		"VP9":   "Google开发的开源编码",
		"AV1":   "下一代开源视频编码",
	}

	if desc, exists := descriptions[codec]; exists {
		return desc
	}
	return "未知编码格式"
}

// GetResolutionDescription 获取分辨率描述
func (e *VideoInfoExtractor) GetResolutionDescription(width, height int) string {
	switch {
	case width == 3840 && height == 2160:
		return "4K Ultra HD"
	case width == 1920 && height == 1080:
		return "Full HD 1080p"
	case width == 1280 && height == 720:
		return "HD 720p"
	case width == 854 && height == 480:
		return "SD 480p"
	case width == 640 && height == 360:
		return "SD 360p"
	default:
		return fmt.Sprintf("自定义分辨率 (%dx%d)", width, height)
	}
}

// IsHighDefinition 判断是否为高清视频
func (e *VideoInfoExtractor) IsHighDefinition(width, height int) bool {
	return width >= 1280 && height >= 720
}

// GetAspectRatio 计算视频宽高比
func (e *VideoInfoExtractor) GetAspectRatio(width, height int) float64 {
	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}