package video

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestThumbnailGenerator_GenerateFromVideo 测试从视频生成缩略图
func TestThumbnailGenerator_GenerateFromVideo(t *testing.T) {
	generator := NewThumbnailGenerator()

	testCases := []struct {
		name        string
		videoData   []byte
		options     *ThumbnailOptions
		expectValid bool
		expectError string
	}{
		{
			name:      "有效的MP4视频",
			videoData: createSampleMP4Data(),
			options: &ThumbnailOptions{
				Width:      320,
				Height:     240,
				Quality:    80,
				Format:     "jpeg",
				TimeOffset: 5, // 5秒处截取
			},
			expectValid: true,
		},
		{
			name:      "使用默认选项",
			videoData: createSampleMP4Data(),
			options:   nil, // 使用默认选项
			expectValid: true,
		},
		{
			name:      "自定义尺寸",
			videoData: createSampleMP4Data(),
			options: &ThumbnailOptions{
				Width:   640,
				Height:  480,
				Quality: 90,
				Format:  "jpeg",
			},
			expectValid: true,
		},
		{
			name:      "PNG格式输出",
			videoData: createSampleMP4Data(),
			options: &ThumbnailOptions{
				Width:  320,
				Height: 240,
				Format: "png",
			},
			expectValid: true,
		},
		{
			name:        "空视频数据",
			videoData:   []byte{},
			options:     &ThumbnailOptions{},
			expectValid: false,
			expectError: "视频数据为空",
		},
		{
			name:        "无效的视频格式",
			videoData:   []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			options:     &ThumbnailOptions{},
			expectValid: false,
			expectError: "无法识别的视频格式",
		},
		{
			name:      "WebM视频",
			videoData: createSampleWebMData(),
			options: &ThumbnailOptions{
				Width:   200,
				Height:  150,
				Format:  "jpeg",
				Quality: 80,
			},
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &ThumbnailRequest{
				VideoData: tc.videoData,
				Options:   tc.options,
			}

			result, err := generator.GenerateFromVideo(request)

			if tc.expectValid {
				assert.NoError(t, err, "缩略图生成应该成功")
				require.NotNil(t, result, "生成结果不应为空")
				assert.NotEmpty(t, result.ImageData, "图片数据不应为空")
				assert.Greater(t, result.Width, 0, "图片宽度应该大于0")
				assert.Greater(t, result.Height, 0, "图片高度应该大于0")
				assert.NotEmpty(t, result.Format, "图片格式不应为空")
				assert.Greater(t, result.FileSize, int64(0), "文件大小应该大于0")

				// 验证生成的图片是否可以正确解码
				_, err = decodeImage(result.ImageData, result.Format)
				assert.NoError(t, err, "生成的图片应该可以正确解码")
			} else {
				assert.Error(t, err, "缩略图生成应该失败")
				assert.Contains(t, err.Error(), tc.expectError, "错误信息应该包含预期内容")
			}
		})
	}
}

// TestThumbnailGenerator_GenerateMultiple 测试生成多个缩略图
func TestThumbnailGenerator_GenerateMultiple(t *testing.T) {
	generator := NewThumbnailGenerator()

	videoData := createSampleMP4Data()
	timeOffsets := []float64{1.0, 5.0, 10.0, 15.0} // 在1s, 5s, 10s, 15s处截取

	request := &MultipleThumbnailRequest{
		VideoData:   videoData,
		TimeOffsets: timeOffsets,
		Options: &ThumbnailOptions{
			Width:   160,
			Height:  120,
			Quality: 75,
			Format:  "jpeg",
		},
	}

	results, err := generator.GenerateMultiple(request)
	assert.NoError(t, err, "多个缩略图生成应该成功")
	require.NotNil(t, results, "生成结果不应为空")
	assert.Len(t, results, len(timeOffsets), "应该生成指定数量的缩略图")

	for i, result := range results {
		assert.NotEmpty(t, result.ImageData, "第%d个缩略图数据不应为空", i+1)
		assert.Equal(t, 160, result.Width, "第%d个缩略图宽度应该正确", i+1)
		assert.Equal(t, 120, result.Height, "第%d个缩略图高度应该正确", i+1)
		assert.Equal(t, timeOffsets[i], result.TimeOffset, "第%d个缩略图时间偏移应该正确", i+1)
	}
}

// TestThumbnailGenerator_ValidateOptions 测试选项验证
func TestThumbnailGenerator_ValidateOptions(t *testing.T) {
	generator := NewThumbnailGenerator()

	testCases := []struct {
		name        string
		options     *ThumbnailOptions
		expectValid bool
		expectError string
	}{
		{
			name: "有效选项",
			options: &ThumbnailOptions{
				Width:   320,
				Height:  240,
				Quality: 80,
				Format:  "jpeg",
			},
			expectValid: true,
		},
		{
			name: "最小尺寸",
			options: &ThumbnailOptions{
				Width:   64,
				Height:  64,
				Format:  "jpeg",
				Quality: 80,
			},
			expectValid: true,
		},
		{
			name: "最大尺寸",
			options: &ThumbnailOptions{
				Width:   1920,
				Height:  1080,
				Format:  "jpeg",
				Quality: 80,
			},
			expectValid: true,
		},
		{
			name: "PNG格式",
			options: &ThumbnailOptions{
				Width:  200,
				Height: 150,
				Format: "png",
			},
			expectValid: true,
		},
		{
			name: "宽度过小",
			options: &ThumbnailOptions{
				Width:  32,
				Height: 240,
				Format: "jpeg",
			},
			expectValid: false,
			expectError: "宽度必须在64到1920之间",
		},
		{
			name: "高度过小",
			options: &ThumbnailOptions{
				Width:  320,
				Height: 32,
				Format: "jpeg",
			},
			expectValid: false,
			expectError: "高度必须在64到1080之间",
		},
		{
			name: "宽度过大",
			options: &ThumbnailOptions{
				Width:  2048,
				Height: 240,
				Format: "jpeg",
			},
			expectValid: false,
			expectError: "宽度必须在64到1920之间",
		},
		{
			name: "不支持的格式",
			options: &ThumbnailOptions{
				Width:  320,
				Height: 240,
				Format: "gif",
			},
			expectValid: false,
			expectError: "不支持的图片格式",
		},
		{
			name: "质量过低",
			options: &ThumbnailOptions{
				Width:   320,
				Height:  240,
				Quality: 0,
				Format:  "jpeg",
			},
			expectValid: false,
			expectError: "JPEG质量必须在1到100之间",
		},
		{
			name: "质量过高",
			options: &ThumbnailOptions{
				Width:   320,
				Height:  240,
				Quality: 101,
				Format:  "jpeg",
			},
			expectValid: false,
			expectError: "JPEG质量必须在1到100之间",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := generator.ValidateOptions(tc.options)

			if tc.expectValid {
				assert.NoError(t, err, "选项验证应该成功")
			} else {
				assert.Error(t, err, "选项验证应该失败")
				assert.Contains(t, err.Error(), tc.expectError, "错误信息应该包含预期内容")
			}
		})
	}
}

// TestThumbnailGenerator_GetDefaultOptions 测试获取默认选项
func TestThumbnailGenerator_GetDefaultOptions(t *testing.T) {
	generator := NewThumbnailGenerator()

	options := generator.GetDefaultOptions()
	require.NotNil(t, options, "默认选项不应为空")

	assert.Equal(t, 320, options.Width, "默认宽度应该是320")
	assert.Equal(t, 240, options.Height, "默认高度应该是240")
	assert.Equal(t, 80, options.Quality, "默认质量应该是80")
	assert.Equal(t, "jpeg", options.Format, "默认格式应该是jpeg")
	assert.Equal(t, 0.0, options.TimeOffset, "默认时间偏移应该是0")
}

// TestThumbnailGenerator_CalculateAspectRatio 测试宽高比计算
func TestThumbnailGenerator_CalculateAspectRatio(t *testing.T) {
	generator := NewThumbnailGenerator()

	testCases := []struct {
		name           string
		originalWidth  int
		originalHeight int
		targetWidth    int
		targetHeight   int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "16:9视频适配4:3目标",
			originalWidth:  1920,
			originalHeight: 1080,
			targetWidth:    320,
			targetHeight:   240,
			expectedWidth:  320,
			expectedHeight: 180, // 保持16:9比例
		},
		{
			name:           "4:3视频适配16:9目标",
			originalWidth:  640,
			originalHeight: 480,
			targetWidth:    320,
			targetHeight:   180,
			expectedWidth:  240, // 保持4:3比例
			expectedHeight: 180,
		},
		{
			name:           "正方形视频",
			originalWidth:  800,
			originalHeight: 800,
			targetWidth:    200,
			targetHeight:   150,
			expectedWidth:  150, // 保持1:1比例
			expectedHeight: 150,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			width, height := generator.CalculateAspectRatio(
				tc.originalWidth, tc.originalHeight,
				tc.targetWidth, tc.targetHeight)

			assert.Equal(t, tc.expectedWidth, width, "计算的宽度应该正确")
			assert.Equal(t, tc.expectedHeight, height, "计算的高度应该正确")
		})
	}
}

// TestThumbnailGenerator_GetSupportedFormats 测试获取支持的格式
func TestThumbnailGenerator_GetSupportedFormats(t *testing.T) {
	generator := NewThumbnailGenerator()

	formats := generator.GetSupportedFormats()
	assert.NotEmpty(t, formats, "支持的格式列表不应为空")
	assert.Contains(t, formats, "jpeg", "应该支持JPEG格式")
	assert.Contains(t, formats, "png", "应该支持PNG格式")
}

// TestThumbnailGenerator_EstimateFileSize 测试文件大小估算
func TestThumbnailGenerator_EstimateFileSize(t *testing.T) {
	generator := NewThumbnailGenerator()

	testCases := []struct {
		name     string
		width    int
		height   int
		format   string
		quality  int
		expected int64 // 大概的文件大小（字节）
	}{
		{
			name:     "小尺寸JPEG高质量",
			width:    160,
			height:   120,
			format:   "jpeg",
			quality:  90,
			expected: 40000, // 约40KB
		},
		{
			name:     "标准尺寸JPEG中等质量",
			width:    320,
			height:   240,
			format:   "jpeg",
			quality:  80,
			expected: 120000, // 约120KB
		},
		{
			name:     "大尺寸PNG",
			width:    640,
			height:   480,
			format:   "png",
			quality:  0, // PNG不使用质量参数
			expected: 800000, // 约800KB
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estimatedSize := generator.EstimateFileSize(tc.width, tc.height, tc.format, tc.quality)
			
			// 允许50%的误差范围
			minSize := int64(float64(tc.expected) * 0.5)
			maxSize := int64(float64(tc.expected) * 1.5)
			
			assert.GreaterOrEqual(t, estimatedSize, minSize, "估算大小不应过小")
			assert.LessOrEqual(t, estimatedSize, maxSize, "估算大小不应过大")
		})
	}
}

// TestThumbnailGenerator_CreatePlaceholder 测试创建占位图片
func TestThumbnailGenerator_CreatePlaceholder(t *testing.T) {
	generator := NewThumbnailGenerator()

	options := &ThumbnailOptions{
		Width:   200,
		Height:  150,
		Format:  "jpeg",
		Quality: 75,
	}

	result, err := generator.CreatePlaceholder(options, "视频处理中...")
	assert.NoError(t, err, "创建占位图片应该成功")
	require.NotNil(t, result, "占位图片结果不应为空")

	assert.Equal(t, 200, result.Width, "占位图片宽度应该正确")
	assert.Equal(t, 150, result.Height, "占位图片高度应该正确")
	assert.Equal(t, "jpeg", result.Format, "占位图片格式应该正确")
	assert.NotEmpty(t, result.ImageData, "占位图片数据不应为空")

	// 验证占位图片可以正确解码
	_, err = decodeImage(result.ImageData, result.Format)
	assert.NoError(t, err, "占位图片应该可以正确解码")
}

// 辅助函数

// decodeImage 解码图片数据验证其有效性
func decodeImage(data []byte, format string) (image.Image, error) {
	reader := bytes.NewReader(data)
	
	switch format {
	case "jpeg":
		return jpeg.Decode(reader)
	case "png":
		return png.Decode(reader)
	default:
		img, _, err := image.Decode(reader)
		return img, err
	}
}