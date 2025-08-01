package video

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
)

// ThumbnailGenerator 缩略图生成器
type ThumbnailGenerator struct {
	validator  *VideoValidator
	extractor  *VideoInfoExtractor
	maxWidth   int
	maxHeight  int
	minWidth   int
	minHeight  int
}

// ThumbnailOptions 缩略图选项
type ThumbnailOptions struct {
	Width      int     `json:"width"`       // 宽度
	Height     int     `json:"height"`      // 高度
	Quality    int     `json:"quality"`     // JPEG质量 (1-100)
	Format     string  `json:"format"`      // 输出格式 (jpeg/png)
	TimeOffset float64 `json:"time_offset"` // 时间偏移（秒）
	KeepAspect bool    `json:"keep_aspect"` // 保持宽高比
}

// ThumbnailRequest 缩略图生成请求
type ThumbnailRequest struct {
	VideoData []byte            `json:"video_data"` // 视频数据
	Options   *ThumbnailOptions `json:"options"`    // 生成选项
}

// MultipleThumbnailRequest 多个缩略图生成请求
type MultipleThumbnailRequest struct {
	VideoData   []byte            `json:"video_data"`   // 视频数据
	TimeOffsets []float64         `json:"time_offsets"` // 时间偏移列表
	Options     *ThumbnailOptions `json:"options"`      // 生成选项
}

// ThumbnailResult 缩略图生成结果
type ThumbnailResult struct {
	ImageData  []byte  `json:"image_data"`  // 图片数据
	Width      int     `json:"width"`       // 实际宽度
	Height     int     `json:"height"`      // 实际高度
	Format     string  `json:"format"`      // 图片格式
	FileSize   int64   `json:"file_size"`   // 文件大小
	TimeOffset float64 `json:"time_offset"` // 时间偏移
}

// NewThumbnailGenerator 创建缩略图生成器
func NewThumbnailGenerator() *ThumbnailGenerator {
	return &ThumbnailGenerator{
		validator:  NewVideoValidator(),
		extractor:  NewVideoInfoExtractor(),
		maxWidth:   1920,
		maxHeight:  1080,
		minWidth:   64,
		minHeight:  64,
	}
}

// GenerateFromVideo 从视频生成缩略图
func (g *ThumbnailGenerator) GenerateFromVideo(request *ThumbnailRequest) (*ThumbnailResult, error) {
	// 验证视频数据
	if len(request.VideoData) == 0 {
		return nil, fmt.Errorf("视频数据为空")
	}

	// 检测视频格式
	format, err := g.validator.DetectFormatByMagicNumber(request.VideoData)
	if err != nil {
		return nil, fmt.Errorf("无法识别的视频格式: %v", err)
	}

	// 使用默认选项（如果未提供）
	options := request.Options
	if options == nil {
		options = g.GetDefaultOptions()
	}

	// 验证选项
	if err := g.ValidateOptions(options); err != nil {
		return nil, err
	}

	// 由于这是一个简化实现，我们创建一个模拟的缩略图
	// 在实际项目中，这里需要使用FFmpeg或类似的视频处理库
	return g.generateMockThumbnail(request.VideoData, options, format)
}

// generateMockThumbnail 生成模拟缩略图（用于演示）
func (g *ThumbnailGenerator) generateMockThumbnail(videoData []byte, options *ThumbnailOptions, format string) (*ThumbnailResult, error) {
	// 创建一个简单的彩色缩略图
	img := image.NewRGBA(image.Rect(0, 0, options.Width, options.Height))
	
	// 根据视频格式使用不同的背景色
	var bgColor color.RGBA
	switch format {
	case "mp4":
		bgColor = color.RGBA{100, 149, 237, 255} // 蓝色
	case "webm":
		bgColor = color.RGBA{144, 238, 144, 255} // 浅绿色
	case "avi":
		bgColor = color.RGBA{255, 182, 193, 255} // 浅粉色
	case "mov":
		bgColor = color.RGBA{255, 215, 0, 255}   // 金色
	default:
		bgColor = color.RGBA{128, 128, 128, 255} // 灰色
	}

	// 填充背景
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 添加一些简单的图案（模拟视频帧）
	g.drawVideoPattern(img, options.Width, options.Height)

	// 编码图片
	var buf bytes.Buffer
	var fileSize int64

	switch options.Format {
	case "jpeg":
		jpegOptions := &jpeg.Options{Quality: options.Quality}
		if err := jpeg.Encode(&buf, img, jpegOptions); err != nil {
			return nil, fmt.Errorf("JPEG编码失败: %v", err)
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("PNG编码失败: %v", err)
		}
	default:
		return nil, fmt.Errorf("不支持的输出格式: %s", options.Format)
	}

	fileSize = int64(buf.Len())

	return &ThumbnailResult{
		ImageData:  buf.Bytes(),
		Width:      options.Width,
		Height:     options.Height,
		Format:     options.Format,
		FileSize:   fileSize,
		TimeOffset: options.TimeOffset,
	}, nil
}

// drawVideoPattern 绘制视频图案
func (g *ThumbnailGenerator) drawVideoPattern(img *image.RGBA, width, height int) {
	bounds := img.Bounds()
	
	// 绘制播放按钮样式的三角形
	centerX := width / 2
	centerY := height / 2
	size := min(width, height) / 6

	// 三角形顶点
	points := []image.Point{
		{centerX - size/2, centerY - size/2},
		{centerX - size/2, centerY + size/2},
		{centerX + size/2, centerY},
	}

	// 填充三角形（简单实现）
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if g.pointInTriangle(x, y, points) {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	// 绘制边框
	borderColor := color.RGBA{255, 255, 255, 128}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		img.Set(x, bounds.Min.Y, borderColor)
		img.Set(x, bounds.Max.Y-1, borderColor)
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		img.Set(bounds.Min.X, y, borderColor)
		img.Set(bounds.Max.X-1, y, borderColor)
	}
}

// pointInTriangle 判断点是否在三角形内
func (g *ThumbnailGenerator) pointInTriangle(px, py int, points []image.Point) bool {
	if len(points) != 3 {
		return false
	}

	x1, y1 := points[0].X, points[0].Y
	x2, y2 := points[1].X, points[1].Y
	x3, y3 := points[2].X, points[2].Y

	// 使用重心坐标法
	denominator := ((y2-y3)*(x1-x3) + (x3-x2)*(y1-y3))
	if denominator == 0 {
		return false
	}

	a := float64((y2-y3)*(px-x3)+(x3-x2)*(py-y3)) / float64(denominator)
	b := float64((y3-y1)*(px-x3)+(x1-x3)*(py-y3)) / float64(denominator)
	c := 1 - a - b

	return a >= 0 && b >= 0 && c >= 0
}

// GenerateMultiple 生成多个缩略图
func (g *ThumbnailGenerator) GenerateMultiple(request *MultipleThumbnailRequest) ([]*ThumbnailResult, error) {
	if len(request.VideoData) == 0 {
		return nil, fmt.Errorf("视频数据为空")
	}

	if len(request.TimeOffsets) == 0 {
		return nil, fmt.Errorf("时间偏移列表不能为空")
	}

	results := make([]*ThumbnailResult, 0, len(request.TimeOffsets))

	for _, timeOffset := range request.TimeOffsets {
		// 复制选项并设置时间偏移
		options := *request.Options
		options.TimeOffset = timeOffset

		thumbnailRequest := &ThumbnailRequest{
			VideoData: request.VideoData,
			Options:   &options,
		}

		result, err := g.GenerateFromVideo(thumbnailRequest)
		if err != nil {
			return nil, fmt.Errorf("生成时间偏移 %.1fs 的缩略图失败: %v", timeOffset, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// ValidateOptions 验证缩略图选项
func (g *ThumbnailGenerator) ValidateOptions(options *ThumbnailOptions) error {
	if options == nil {
		return fmt.Errorf("选项不能为空")
	}

	// 验证尺寸
	if options.Width < g.minWidth || options.Width > g.maxWidth {
		return fmt.Errorf("宽度必须在%d到%d之间", g.minWidth, g.maxWidth)
	}

	if options.Height < g.minHeight || options.Height > g.maxHeight {
		return fmt.Errorf("高度必须在%d到%d之间", g.minHeight, g.maxHeight)
	}

	// 验证格式
	supportedFormats := []string{"jpeg", "png"}
	formatSupported := false
	for _, format := range supportedFormats {
		if options.Format == format {
			formatSupported = true
			break
		}
	}
	if !formatSupported {
		return fmt.Errorf("不支持的图片格式: %s，支持的格式: %v", options.Format, supportedFormats)
	}

	// 验证JPEG质量
	if options.Format == "jpeg" {
		if options.Quality < 1 || options.Quality > 100 {
			return fmt.Errorf("JPEG质量必须在1到100之间")
		}
	}

	// 验证时间偏移
	if options.TimeOffset < 0 {
		return fmt.Errorf("时间偏移不能为负数")
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (g *ThumbnailGenerator) GetDefaultOptions() *ThumbnailOptions {
	return &ThumbnailOptions{
		Width:      320,
		Height:     240,
		Quality:    80,
		Format:     "jpeg",
		TimeOffset: 0.0,
		KeepAspect: true,
	}
}

// CalculateAspectRatio 计算保持宽高比的尺寸
func (g *ThumbnailGenerator) CalculateAspectRatio(originalWidth, originalHeight, targetWidth, targetHeight int) (int, int) {
	originalAspect := float64(originalWidth) / float64(originalHeight)
	targetAspect := float64(targetWidth) / float64(targetHeight)

	var newWidth, newHeight int

	if originalAspect > targetAspect {
		// 原始视频更宽，以宽度为准
		newWidth = targetWidth
		newHeight = int(float64(targetWidth) / originalAspect)
	} else {
		// 原始视频更高，以高度为准
		newHeight = targetHeight
		newWidth = int(float64(targetHeight) * originalAspect)
	}

	return newWidth, newHeight
}

// GetSupportedFormats 获取支持的输出格式
func (g *ThumbnailGenerator) GetSupportedFormats() []string {
	return []string{"jpeg", "png"}
}

// EstimateFileSize 估算缩略图文件大小
func (g *ThumbnailGenerator) EstimateFileSize(width, height int, format string, quality int) int64 {
	pixels := int64(width * height)
	
	switch format {
	case "jpeg":
		// JPEG大小估算：基于质量和像素数
		qualityFactor := float64(quality) / 100.0
		bytesPerPixel := 0.5 + (qualityFactor * 2.0) // 0.5-2.5 bytes per pixel
		return int64(float64(pixels) * bytesPerPixel)
		
	case "png":
		// PNG大小估算：通常比JPEG大
		bytesPerPixel := 3.0 // 大约3 bytes per pixel for PNG
		return int64(float64(pixels) * bytesPerPixel)
		
	default:
		// 默认估算
		return pixels * 2
	}
}

// CreatePlaceholder 创建占位图片
func (g *ThumbnailGenerator) CreatePlaceholder(options *ThumbnailOptions, text string) (*ThumbnailResult, error) {
	if err := g.ValidateOptions(options); err != nil {
		return nil, err
	}

	// 创建占位图片
	img := image.NewRGBA(image.Rect(0, 0, options.Width, options.Height))
	
	// 填充背景（浅灰色）
	bgColor := color.RGBA{240, 240, 240, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 绘制边框
	borderColor := color.RGBA{200, 200, 200, 255}
	bounds := img.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		img.Set(x, bounds.Min.Y, borderColor)
		img.Set(x, bounds.Max.Y-1, borderColor)
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		img.Set(bounds.Min.X, y, borderColor)
		img.Set(bounds.Max.X-1, y, borderColor)
	}

	// 绘制简单的相机图标
	g.drawCameraIcon(img, options.Width, options.Height)

	// 编码图片
	var buf bytes.Buffer
	switch options.Format {
	case "jpeg":
		jpegOptions := &jpeg.Options{Quality: options.Quality}
		if err := jpeg.Encode(&buf, img, jpegOptions); err != nil {
			return nil, fmt.Errorf("JPEG编码失败: %v", err)
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("PNG编码失败: %v", err)
		}
	}

	return &ThumbnailResult{
		ImageData:  buf.Bytes(),
		Width:      options.Width,
		Height:     options.Height,
		Format:     options.Format,
		FileSize:   int64(buf.Len()),
		TimeOffset: options.TimeOffset,
	}, nil
}

// drawCameraIcon 绘制相机图标
func (g *ThumbnailGenerator) drawCameraIcon(img *image.RGBA, width, height int) {
	centerX := width / 2
	centerY := height / 2
	iconSize := min(width, height) / 4

	iconColor := color.RGBA{150, 150, 150, 255}

	// 绘制相机主体（矩形）
	for y := centerY - iconSize/2; y <= centerY + iconSize/2; y++ {
		for x := centerX - iconSize/2; x <= centerX + iconSize/2; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				// 只绘制边框
				if y == centerY - iconSize/2 || y == centerY + iconSize/2 ||
				   x == centerX - iconSize/2 || x == centerX + iconSize/2 {
					img.Set(x, y, iconColor)
				}
			}
		}
	}

	// 绘制镜头（圆形）
	lensRadius := iconSize / 4
	for y := centerY - lensRadius; y <= centerY + lensRadius; y++ {
		for x := centerX - lensRadius; x <= centerX + lensRadius; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				dx := x - centerX
				dy := y - centerY
				if dx*dx + dy*dy <= lensRadius*lensRadius {
					img.Set(x, y, iconColor)
				}
			}
		}
	}
}

// GetMaxDimensions 获取最大尺寸限制
func (g *ThumbnailGenerator) GetMaxDimensions() (int, int) {
	return g.maxWidth, g.maxHeight
}

// GetMinDimensions 获取最小尺寸限制
func (g *ThumbnailGenerator) GetMinDimensions() (int, int) {
	return g.minWidth, g.minHeight
}

// SetMaxDimensions 设置最大尺寸限制
func (g *ThumbnailGenerator) SetMaxDimensions(width, height int) {
	if width > 0 {
		g.maxWidth = width
	}
	if height > 0 {
		g.maxHeight = height
	}
}

// SetMinDimensions 设置最小尺寸限制
func (g *ThumbnailGenerator) SetMinDimensions(width, height int) {
	if width > 0 {
		g.minWidth = width
	}
	if height > 0 {
		g.minHeight = height
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetQualityDescription 获取质量描述
func (g *ThumbnailGenerator) GetQualityDescription(quality int) string {
	switch {
	case quality >= 90:
		return "极高质量"
	case quality >= 80:
		return "高质量"
	case quality >= 60:
		return "中等质量"
	case quality >= 40:
		return "低质量"
	default:
		return "极低质量"
	}
}

// GetFormatDescription 获取格式描述
func (g *ThumbnailGenerator) GetFormatDescription(format string) string {
	descriptions := map[string]string{
		"jpeg": "JPEG格式，适合照片，文件较小",
		"png":  "PNG格式，支持透明，文件较大",
	}
	
	if desc, exists := descriptions[format]; exists {
		return desc
	}
	return "未知格式"
}