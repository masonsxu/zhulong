package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVideoValidator_ValidateFormat 测试视频格式验证
func TestVideoValidator_ValidateFormat(t *testing.T) {
	validator := NewVideoValidator()

	testCases := []struct {
		name        string
		filename    string
		contentType string
		data        []byte
		expectValid bool
		expectError string
	}{
		{
			name:        "有效的MP4文件",
			filename:    "test.mp4",
			contentType: "video/mp4",
			data:        []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x31}, // MP4魔数
			expectValid: true,
		},
		{
			name:        "有效的WebM文件",
			filename:    "test.webm",
			contentType: "video/webm",
			data:        []byte{0x1A, 0x45, 0xDF, 0xA3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // WebM魔数，填充到12字节
			expectValid: true,
		},
		{
			name:        "有效的AVI文件",
			filename:    "test.avi",
			contentType: "video/avi",
			data:        []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x41, 0x56, 0x49, 0x20}, // AVI魔数
			expectValid: true,
		},
		{
			name:        "有效的MOV文件",
			filename:    "test.mov",
			contentType: "video/quicktime",
			data:        []byte{0x00, 0x00, 0x00, 0x14, 0x66, 0x74, 0x79, 0x70, 0x71, 0x74, 0x20, 0x20}, // MOV魔数
			expectValid: true,
		},
		{
			name:        "不支持的格式",
			filename:    "test.wmv",
			contentType: "video/x-ms-wmv",
			data:        []byte{0x30, 0x26, 0xB2, 0x75, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // WMV魔数，填充到12字节
			expectValid: false,
			expectError: "不支持的视频格式",
		},
		{
			name:        "文件扩展名与内容不匹配",
			filename:    "test.mp4",
			contentType: "video/mp4",
			data:        []byte{0x1A, 0x45, 0xDF, 0xA3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // WebM魔数但声称是MP4，填充到12字节
			expectValid: false,
			expectError: "文件内容与扩展名不匹配",
		},
		{
			name:        "空文件",
			filename:    "test.mp4",
			contentType: "video/mp4",
			data:        []byte{},
			expectValid: false,
			expectError: "文件内容为空",
		},
		{
			name:        "文件头太短",
			filename:    "test.mp4",
			contentType: "video/mp4",
			data:        []byte{0x00, 0x00}, // 只有2字节
			expectValid: false,
			expectError: "文件头信息不完整",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &ValidationRequest{
				Filename:    tc.filename,
				ContentType: tc.contentType,
				Data:        tc.data,
			}

			result, err := validator.ValidateFormat(request)

			if tc.expectValid {
				assert.NoError(t, err, "验证应该成功")
				require.NotNil(t, result, "验证结果不应为空")
				assert.True(t, result.IsValid, "文件应该是有效的")
				assert.NotEmpty(t, result.DetectedFormat, "应该检测到文件格式")
				assert.Empty(t, result.ErrorMessage, "成功时不应有错误信息")
			} else {
				if tc.expectError != "" {
					assert.Error(t, err, "验证应该失败")
					assert.Contains(t, err.Error(), tc.expectError, "错误信息应该包含预期内容")
				} else {
					require.NotNil(t, result, "验证结果不应为空")
					assert.False(t, result.IsValid, "文件应该是无效的")
					assert.NotEmpty(t, result.ErrorMessage, "失败时应该有错误信息")
				}
			}
		})
	}
}

// TestVideoValidator_GetSupportedFormats 测试获取支持的格式列表
func TestVideoValidator_GetSupportedFormats(t *testing.T) {
	validator := NewVideoValidator()

	formats := validator.GetSupportedFormats()
	assert.NotEmpty(t, formats, "支持的格式列表不应为空")

	// 验证包含预期的格式
	expectedFormats := []string{"mp4", "webm", "avi", "mov"}
	for _, expected := range expectedFormats {
		assert.Contains(t, formats, expected, "应该支持%s格式", expected)
	}
}

// TestVideoValidator_ValidateFileSize 测试文件大小验证
func TestVideoValidator_ValidateFileSize(t *testing.T) {
	validator := NewVideoValidator()

	testCases := []struct {
		name        string
		size        int64
		expectValid bool
		expectError string
	}{
		{
			name:        "正常大小文件",
			size:        100 * 1024 * 1024, // 100MB
			expectValid: true,
		},
		{
			name:        "最大允许大小",
			size:        2 * 1024 * 1024 * 1024, // 2GB
			expectValid: true,
		},
		{
			name:        "文件过大",
			size:        3 * 1024 * 1024 * 1024, // 3GB
			expectValid: false,
			expectError: "文件大小超过限制",
		},
		{
			name:        "文件为空",
			size:        0,
			expectValid: false,
			expectError: "文件不能为空",
		},
		{
			name:        "负数大小",
			size:        -1,
			expectValid: false,
			expectError: "文件大小无效",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateFileSize(tc.size)

			if tc.expectValid {
				assert.NoError(t, err, "文件大小验证应该成功")
			} else {
				assert.Error(t, err, "文件大小验证应该失败")
				assert.Contains(t, err.Error(), tc.expectError, "错误信息应该包含预期内容")
			}
		})
	}
}

// TestVideoValidator_DetectFormatByMagicNumber 测试通过魔数检测格式
func TestVideoValidator_DetectFormatByMagicNumber(t *testing.T) {
	validator := NewVideoValidator()

	testCases := []struct {
		name           string
		data           []byte
		expectedFormat string
		expectError    bool
	}{
		{
			name:           "MP4格式",
			data:           []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x31},
			expectedFormat: "mp4",
			expectError:    false,
		},
		{
			name:           "WebM格式",
			data:           []byte{0x1A, 0x45, 0xDF, 0xA3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedFormat: "webm",
			expectError:    false,
		},
		{
			name:           "AVI格式",
			data:           []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x41, 0x56, 0x49, 0x20},
			expectedFormat: "avi",
			expectError:    false,
		},
		{
			name:           "MOV格式",
			data:           []byte{0x00, 0x00, 0x00, 0x14, 0x66, 0x74, 0x79, 0x70, 0x71, 0x74, 0x20, 0x20},
			expectedFormat: "mov",
			expectError:    false,
		},
		{
			name:        "未知格式",
			data:        []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expectError: true,
		},
		{
			name:        "数据太短",
			data:        []byte{0x00, 0x00},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format, err := validator.DetectFormatByMagicNumber(tc.data)

			if tc.expectError {
				assert.Error(t, err, "检测应该失败")
			} else {
				assert.NoError(t, err, "检测应该成功")
				assert.Equal(t, tc.expectedFormat, format, "检测到的格式应该匹配")
			}
		})
	}
}

// TestVideoValidator_ValidateContentType 测试内容类型验证
func TestVideoValidator_ValidateContentType(t *testing.T) {
	validator := NewVideoValidator()

	testCases := []struct {
		name        string
		contentType string
		expectValid bool
	}{
		{
			name:        "MP4内容类型",
			contentType: "video/mp4",
			expectValid: true,
		},
		{
			name:        "WebM内容类型",
			contentType: "video/webm",
			expectValid: true,
		},
		{
			name:        "AVI内容类型",
			contentType: "video/avi",
			expectValid: true,
		},
		{
			name:        "AVI替代类型",
			contentType: "video/x-msvideo",
			expectValid: true,
		},
		{
			name:        "MOV内容类型",
			contentType: "video/quicktime",
			expectValid: true,
		},
		{
			name:        "不支持的内容类型",
			contentType: "video/x-ms-wmv",
			expectValid: false,
		},
		{
			name:        "非视频类型",
			contentType: "audio/mp3",
			expectValid: false,
		},
		{
			name:        "空内容类型",
			contentType: "",
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateContentType(tc.contentType)

			if tc.expectValid {
				assert.NoError(t, err, "内容类型验证应该成功")
			} else {
				assert.Error(t, err, "内容类型验证应该失败")
			}
		})
	}
}

// TestVideoValidator_ComprehensiveValidation 测试综合验证
func TestVideoValidator_ComprehensiveValidation(t *testing.T) {
	validator := NewVideoValidator()

	// 创建一个模拟的MP4文件内容
	mp4Data := make([]byte, 1024) // 1KB
	copy(mp4Data, []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x31})

	request := &ComprehensiveValidationRequest{
		Filename:    "test-video.mp4",
		ContentType: "video/mp4",
		Data:        mp4Data,
		Size:        int64(len(mp4Data)),
	}

	result, err := validator.ComprehensiveValidation(request)
	assert.NoError(t, err, "综合验证应该成功")
	require.NotNil(t, result, "验证结果不应为空")

	assert.True(t, result.IsValid, "文件应该是有效的")
	assert.Equal(t, "mp4", result.DetectedFormat, "检测到的格式应该是mp4")
	assert.True(t, result.FormatValid, "格式应该有效")
	assert.True(t, result.SizeValid, "大小应该有效")
	assert.True(t, result.ContentTypeValid, "内容类型应该有效")
	assert.Empty(t, result.Errors, "不应该有错误")
}

// TestVideoValidator_GetMaxFileSize 测试获取最大文件大小限制
func TestVideoValidator_GetMaxFileSize(t *testing.T) {
	validator := NewVideoValidator()

	maxSize := validator.GetMaxFileSize()
	assert.Equal(t, int64(2*1024*1024*1024), maxSize, "最大文件大小应该是2GB")
}

// TestVideoValidator_IsFormatSupported 测试格式支持检查
func TestVideoValidator_IsFormatSupported(t *testing.T) {
	validator := NewVideoValidator()

	supportedFormats := []string{"mp4", "webm", "avi", "mov"}
	for _, format := range supportedFormats {
		assert.True(t, validator.IsFormatSupported(format), "%s格式应该被支持", format)
	}

	unsupportedFormats := []string{"wmv", "flv", "mkv", "rmvb"}
	for _, format := range unsupportedFormats {
		assert.False(t, validator.IsFormatSupported(format), "%s格式不应该被支持", format)
	}
}