package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSizeLimitManager_GetMaxFileSize 测试获取最大文件大小
func TestSizeLimitManager_GetMaxFileSize(t *testing.T) {
	manager := NewSizeLimitManager()

	maxSize := manager.GetMaxFileSize()
	assert.Equal(t, int64(2*1024*1024*1024), maxSize, "默认最大文件大小应该是2GB")
}

// TestSizeLimitManager_SetMaxFileSize 测试设置最大文件大小
func TestSizeLimitManager_SetMaxFileSize(t *testing.T) {
	manager := NewSizeLimitManager()

	// 设置新的大小限制
	newSize := int64(1024 * 1024 * 1024) // 1GB
	manager.SetMaxFileSize(newSize)

	assert.Equal(t, newSize, manager.GetMaxFileSize(), "应该更新最大文件大小")
}

// TestSizeLimitManager_ValidateSize 测试文件大小验证
func TestSizeLimitManager_ValidateSize(t *testing.T) {
	manager := NewSizeLimitManager()

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
			name:        "边界值-1字节",
			size:        2*1024*1024*1024 - 1, // 2GB-1
			expectValid: true,
		},
		{
			name:        "边界值-正好2GB",
			size:        2 * 1024 * 1024 * 1024, // 2GB
			expectValid: true,
		},
		{
			name:        "超过限制-1字节",
			size:        2*1024*1024*1024 + 1, // 2GB+1
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
		{
			name:        "非常大的文件",
			size:        10 * 1024 * 1024 * 1024, // 10GB
			expectValid: false,
			expectError: "文件大小超过限制",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.ValidateSize(tc.size)

			if tc.expectValid {
				assert.NoError(t, err, "文件大小验证应该成功")
			} else {
				assert.Error(t, err, "文件大小验证应该失败")
				assert.Contains(t, err.Error(), tc.expectError, "错误信息应该包含预期内容")
			}
		})
	}
}

// TestSizeLimitManager_GetSizeInDifferentUnits 测试获取不同单位的大小
func TestSizeLimitManager_GetSizeInDifferentUnits(t *testing.T) {
	manager := NewSizeLimitManager()

	// 测试获取MB单位的大小
	sizeInMB := manager.GetMaxFileSizeInMB()
	assert.Equal(t, int64(2048), sizeInMB, "2GB应该等于2048MB")

	// 测试获取KB单位的大小
	sizeInKB := manager.GetMaxFileSizeInKB()
	assert.Equal(t, int64(2097152), sizeInKB, "2GB应该等于2097152KB")
}

// TestSizeLimitManager_FormatSize 测试格式化文件大小显示
func TestSizeLimitManager_FormatSize(t *testing.T) {
	manager := NewSizeLimitManager()

	testCases := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "字节单位",
			size:     512,
			expected: "512 B",
		},
		{
			name:     "KB单位",
			size:     1024,
			expected: "1.00 KB",
		},
		{
			name:     "MB单位",
			size:     1024 * 1024,
			expected: "1.00 MB",
		},
		{
			name:     "GB单位",
			size:     1024 * 1024 * 1024,
			expected: "1.00 GB",
		},
		{
			name:     "2GB",
			size:     2 * 1024 * 1024 * 1024,
			expected: "2.00 GB",
		},
		{
			name:     "混合单位",
			size:     1536 * 1024 * 1024, // 1.5GB
			expected: "1.50 GB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := manager.FormatSize(tc.size)
			assert.Equal(t, tc.expected, formatted, "格式化结果应该匹配")
		})
	}
}

// TestSizeLimitManager_ValidateSizeWithCustomLimit 测试自定义限制的大小验证
func TestSizeLimitManager_ValidateSizeWithCustomLimit(t *testing.T) {
	manager := NewSizeLimitManager()

	// 设置较小的限制进行测试
	customLimit := int64(500 * 1024 * 1024) // 500MB
	manager.SetMaxFileSize(customLimit)

	testCases := []struct {
		name        string
		size        int64
		expectValid bool
	}{
		{
			name:        "低于自定义限制",
			size:        400 * 1024 * 1024, // 400MB
			expectValid: true,
		},
		{
			name:        "等于自定义限制",
			size:        500 * 1024 * 1024, // 500MB
			expectValid: true,
		},
		{
			name:        "超过自定义限制",
			size:        600 * 1024 * 1024, // 600MB
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.ValidateSize(tc.size)

			if tc.expectValid {
				assert.NoError(t, err, "文件大小验证应该成功")
			} else {
				assert.Error(t, err, "文件大小验证应该失败")
			}
		})
	}
}

// TestSizeLimitManager_GetDefaultLimits 测试获取默认限制信息
func TestSizeLimitManager_GetDefaultLimits(t *testing.T) {
	manager := NewSizeLimitManager()

	limits := manager.GetLimits()
	require.NotNil(t, limits, "限制信息不应为空")

	assert.Equal(t, int64(2*1024*1024*1024), limits.MaxFileSize, "最大文件大小应该是2GB")
	assert.Equal(t, int64(1), limits.MinFileSize, "最小文件大小应该是1字节")
	assert.Equal(t, "2.00 GB", limits.MaxFileSizeFormatted, "格式化的最大文件大小")
	assert.Equal(t, "1 B", limits.MinFileSizeFormatted, "格式化的最小文件大小")
}

// TestSizeLimitManager_UpdateLimits 测试更新限制信息
func TestSizeLimitManager_UpdateLimits(t *testing.T) {
	manager := NewSizeLimitManager()

	// 更新限制
	newLimits := &SizeLimits{
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		MinFileSize: 1024,               // 1KB
	}

	err := manager.UpdateLimits(newLimits)
	assert.NoError(t, err, "更新限制应该成功")

	// 验证更新后的限制
	limits := manager.GetLimits()
	assert.Equal(t, newLimits.MaxFileSize, limits.MaxFileSize, "最大文件大小应该已更新")
	assert.Equal(t, newLimits.MinFileSize, limits.MinFileSize, "最小文件大小应该已更新")
}

// TestSizeLimitManager_ValidateWithVideoTypes 测试不同视频格式的大小限制
func TestSizeLimitManager_ValidateWithVideoTypes(t *testing.T) {
	manager := NewSizeLimitManager()

	// 设置不同格式的大小限制
	limits := map[string]int64{
		"mp4":  2 * 1024 * 1024 * 1024, // 2GB
		"webm": 1 * 1024 * 1024 * 1024, // 1GB
		"avi":  3 * 1024 * 1024 * 1024, // 3GB
		"mov":  2 * 1024 * 1024 * 1024, // 2GB
	}

	manager.SetFormatLimits(limits)

	testCases := []struct {
		name        string
		format      string
		size        int64
		expectValid bool
	}{
		{
			name:        "MP4格式在限制内",
			format:      "mp4",
			size:        1.5 * 1024 * 1024 * 1024, // 1.5GB
			expectValid: true,
		},
		{
			name:        "WebM格式超过限制",
			format:      "webm",
			size:        1.5 * 1024 * 1024 * 1024, // 1.5GB
			expectValid: false,
		},
		{
			name:        "AVI格式在限制内",
			format:      "avi",
			size:        2.5 * 1024 * 1024 * 1024, // 2.5GB
			expectValid: true,
		},
		{
			name:        "未知格式使用默认限制",
			format:      "unknown",
			size:        1.5 * 1024 * 1024 * 1024, // 1.5GB
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.ValidateSizeForFormat(tc.format, tc.size)

			if tc.expectValid {
				assert.NoError(t, err, "格式特定的大小验证应该成功")
			} else {
				assert.Error(t, err, "格式特定的大小验证应该失败")
			}
		})
	}
}