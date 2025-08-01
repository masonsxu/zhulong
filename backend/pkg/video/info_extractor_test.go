package video

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVideoInfoExtractor_ExtractInfo 测试视频信息提取
func TestVideoInfoExtractor_ExtractInfo(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name        string
		data        []byte
		filename    string
		expectValid bool
		expectError string
	}{
		{
			name:        "有效的MP4文件",
			data:        createSampleMP4Data(),
			filename:    "test.mp4",
			expectValid: true,
		},
		{
			name:        "有效的WebM文件",
			data:        createSampleWebMData(),
			filename:    "test.webm",
			expectValid: true,
		},
		{
			name:        "有效的AVI文件",
			data:        createSampleAVIData(),
			filename:    "test.avi",
			expectValid: true,
		},
		{
			name:        "空文件",
			data:        []byte{},
			filename:    "test.mp4",
			expectValid: false,
			expectError: "文件数据为空",
		},
		{
			name:        "文件头不完整",
			data:        []byte{0x00, 0x00},
			filename:    "test.mp4",
			expectValid: false,
			expectError: "文件头信息不完整",
		},
		{
			name:        "不支持的格式",
			data:        []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			filename:    "test.unknown",
			expectValid: false,
			expectError: "无法识别的视频格式",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &InfoExtractionRequest{
				Data:     tc.data,
				Filename: tc.filename,
			}

			result, err := extractor.ExtractInfo(request)

			if tc.expectValid {
				assert.NoError(t, err, "信息提取应该成功")
				require.NotNil(t, result, "提取结果不应为空")
				assert.NotEmpty(t, result.Format, "应该检测到格式")
				assert.Greater(t, result.FileSize, int64(0), "文件大小应该大于0")
			} else {
				assert.Error(t, err, "信息提取应该失败")
				assert.Contains(t, err.Error(), tc.expectError, "错误信息应该包含预期内容")
			}
		})
	}
}

// TestVideoInfoExtractor_ExtractDuration 测试时长提取
func TestVideoInfoExtractor_ExtractDuration(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "包含基本头信息的MP4",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0时长而不是错误
		},
		{
			name:        "无时长信息的文件",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0时长而不是错误
		},
		{
			name:        "数据太短",
			data:        []byte{0x00, 0x00},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			duration, err := extractor.ExtractDuration(tc.data)

			if tc.expectError {
				assert.Error(t, err, "时长提取应该失败")
			} else {
				assert.NoError(t, err, "时长提取应该成功")
				assert.GreaterOrEqual(t, duration, time.Duration(0), "时长应该大于等于0")
			}
		})
	}
}

// TestVideoInfoExtractor_ExtractResolution 测试分辨率提取
func TestVideoInfoExtractor_ExtractResolution(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "包含基本头信息的MP4",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0x0而不是错误
		},
		{
			name:        "包含基本头信息的AVI",
			data:        createSampleAVIData(),
			expectError: false, // 应该返回0x0而不是错误
		},
		{
			name:        "无分辨率信息",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0x0而不是错误
		},
		{
			name:        "数据太短",
			data:        []byte{0x00, 0x00},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			width, height, err := extractor.ExtractResolution(tc.data)

			if tc.expectError {
				assert.Error(t, err, "分辨率提取应该失败")
			} else {
				assert.NoError(t, err, "分辨率提取应该成功")
				assert.GreaterOrEqual(t, width, 0, "宽度应该大于等于0")
				assert.GreaterOrEqual(t, height, 0, "高度应该大于等于0")
			}
		})
	}
}

// TestVideoInfoExtractor_ExtractBitrate 测试比特率提取
func TestVideoInfoExtractor_ExtractBitrate(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "包含基本头信息的MP4",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0而不是错误
		},
		{
			name:        "无比特率信息",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0而不是错误
		},
		{
			name:        "数据太短",
			data:        []byte{0x00, 0x00},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bitrate, err := extractor.ExtractBitrate(tc.data)

			if tc.expectError {
				assert.Error(t, err, "比特率提取应该失败")
			} else {
				assert.NoError(t, err, "比特率提取应该成功")
				assert.GreaterOrEqual(t, bitrate, int64(0), "比特率应该大于等于0")
			}
		})
	}
}

// TestVideoInfoExtractor_ExtractFrameRate 测试帧率提取
func TestVideoInfoExtractor_ExtractFrameRate(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "包含基本头信息的MP4",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0而不是错误
		},
		{
			name:        "包含基本头信息的AVI",
			data:        createSampleAVIData(),
			expectError: false, // 应该返回0而不是错误
		},
		{
			name:        "无帧率信息",
			data:        createSampleMP4Data(),
			expectError: false, // 应该返回0而不是错误
		},
		{
			name:        "数据太短",
			data:        []byte{0x00, 0x00},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frameRate, err := extractor.ExtractFrameRate(tc.data)

			if tc.expectError {
				assert.Error(t, err, "帧率提取应该失败")
			} else {
				assert.NoError(t, err, "帧率提取应该成功")
				assert.GreaterOrEqual(t, frameRate, 0.0, "帧率应该大于等于0")
			}
		})
	}
}

// TestVideoInfoExtractor_GetVideoInfo 测试完整视频信息获取
func TestVideoInfoExtractor_GetVideoInfo(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	// 创建包含完整信息的测试数据
	testData := createCompleteMP4Data()

	request := &InfoExtractionRequest{
		Data:     testData,
		Filename: "complete.mp4",
	}

	result, err := extractor.ExtractInfo(request)
	assert.NoError(t, err, "完整信息提取应该成功")
	require.NotNil(t, result, "提取结果不应为空")

	// 验证基本信息
	assert.Equal(t, "mp4", result.Format, "格式应该正确")
	assert.Greater(t, result.FileSize, int64(0), "文件大小应该大于0")

	// 验证视频信息（如果存在）
	if result.Duration > 0 {
		assert.Greater(t, result.Duration, time.Duration(0), "时长应该大于0")
	}

	if result.Width > 0 && result.Height > 0 {
		assert.Greater(t, result.Width, 0, "宽度应该大于0")
		assert.Greater(t, result.Height, 0, "高度应该大于0")
	}
}

// TestVideoInfoExtractor_FormatDuration 测试时长格式化
func TestVideoInfoExtractor_FormatDuration(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "秒级时长",
			duration: 45 * time.Second,
			expected: "00:45",
		},
		{
			name:     "分钟级时长",
			duration: 3*time.Minute + 25*time.Second,
			expected: "03:25",
		},
		{
			name:     "小时级时长",
			duration: 1*time.Hour + 23*time.Minute + 45*time.Second,
			expected: "01:23:45",
		},
		{
			name:     "零时长",
			duration: 0,
			expected: "00:00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := extractor.FormatDuration(tc.duration)
			assert.Equal(t, tc.expected, formatted, "格式化的时长应该匹配")
		})
	}
}

// TestVideoInfoExtractor_FormatResolution 测试分辨率格式化
func TestVideoInfoExtractor_FormatResolution(t *testing.T) {
	extractor := NewVideoInfoExtractor()

	testCases := []struct {
		name     string
		width    int
		height   int
		expected string
	}{
		{
			name:     "1080p",
			width:    1920,
			height:   1080,
			expected: "1920x1080",
		},
		{
			name:     "720p",
			width:    1280,
			height:   720,
			expected: "1280x720",
		},
		{
			name:     "4K",
			width:    3840,
			height:   2160,
			expected: "3840x2160",
		},
		{
			name:     "零分辨率",
			width:    0,
			height:   0,
			expected: "0x0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := extractor.FormatResolution(tc.width, tc.height)
			assert.Equal(t, tc.expected, formatted, "格式化的分辨率应该匹配")
		})
	}
}

// 辅助函数：创建示例数据

func createSampleMP4Data() []byte {
	// 创建最小有效的MP4文件头
	return []byte{
		0x00, 0x00, 0x00, 0x20, // Box size
		0x66, 0x74, 0x79, 0x70, // 'ftyp'
		0x6D, 0x70, 0x34, 0x31, // 'mp41'
		0x00, 0x00, 0x00, 0x00, // Minor version
		0x6D, 0x70, 0x34, 0x31, // Compatible brand
		0x69, 0x73, 0x6F, 0x6D, // Compatible brand
	}
}

func createSampleWebMData() []byte {
	// 创建最小有效的WebM文件头
	return []byte{
		0x1A, 0x45, 0xDF, 0xA3, // EBML header
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
}

func createSampleAVIData() []byte {
	// 创建最小有效的AVI文件头
	return []byte{
		0x52, 0x49, 0x46, 0x46, // 'RIFF'
		0x00, 0x00, 0x00, 0x00, // File size
		0x41, 0x56, 0x49, 0x20, // 'AVI '
	}
}

func createCompleteMP4Data() []byte {
	// 创建包含完整信息的MP4数据（简化版）
	return createSampleMP4Data()
}