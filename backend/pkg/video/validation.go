package video

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

// VideoValidator 视频格式验证器
type VideoValidator struct {
	supportedFormats   map[string]bool
	contentTypeMapping map[string]string
	magicNumbers       map[string][]byte
	maxFileSize        int64
}

// ValidationRequest 格式验证请求
type ValidationRequest struct {
	Filename    string `json:"filename"`    // 文件名
	ContentType string `json:"content_type"` // 内容类型
	Data        []byte `json:"data"`        // 文件数据（至少前512字节）
}

// ValidationResult 格式验证结果
type ValidationResult struct {
	IsValid        bool   `json:"is_valid"`        // 是否有效
	DetectedFormat string `json:"detected_format"` // 检测到的格式
	ErrorMessage   string `json:"error_message"`   // 错误信息
}

// ComprehensiveValidationRequest 综合验证请求
type ComprehensiveValidationRequest struct {
	Filename    string `json:"filename"`     // 文件名
	ContentType string `json:"content_type"` // 内容类型
	Data        []byte `json:"data"`         // 文件数据
	Size        int64  `json:"size"`         // 文件大小
}

// ComprehensiveValidationResult 综合验证结果
type ComprehensiveValidationResult struct {
	IsValid          bool     `json:"is_valid"`           // 总体是否有效
	DetectedFormat   string   `json:"detected_format"`    // 检测到的格式
	FormatValid      bool     `json:"format_valid"`       // 格式是否有效
	SizeValid        bool     `json:"size_valid"`         // 大小是否有效
	ContentTypeValid bool     `json:"content_type_valid"` // 内容类型是否有效
	Errors           []string `json:"errors"`             // 错误列表
}

// NewVideoValidator 创建视频验证器
func NewVideoValidator() *VideoValidator {
	validator := &VideoValidator{
		supportedFormats:   make(map[string]bool),
		contentTypeMapping: make(map[string]string),
		magicNumbers:       make(map[string][]byte),
		maxFileSize:        2 * 1024 * 1024 * 1024, // 2GB
	}

	// 初始化支持的格式
	validator.initSupportedFormats()
	validator.initContentTypeMapping()
	validator.initMagicNumbers()

	return validator
}

// initSupportedFormats 初始化支持的格式
func (v *VideoValidator) initSupportedFormats() {
	formats := []string{"mp4", "webm", "avi", "mov"}
	for _, format := range formats {
		v.supportedFormats[format] = true
	}
}

// initContentTypeMapping 初始化内容类型映射
func (v *VideoValidator) initContentTypeMapping() {
	v.contentTypeMapping["video/mp4"] = "mp4"
	v.contentTypeMapping["video/webm"] = "webm"
	v.contentTypeMapping["video/avi"] = "avi"
	v.contentTypeMapping["video/x-msvideo"] = "avi"
	v.contentTypeMapping["video/quicktime"] = "mov"
}

// initMagicNumbers 初始化文件魔数
func (v *VideoValidator) initMagicNumbers() {
	// MP4 魔数：ftyp
	v.magicNumbers["mp4"] = []byte{0x66, 0x74, 0x79, 0x70}
	
	// WebM 魔数：EBML header
	v.magicNumbers["webm"] = []byte{0x1A, 0x45, 0xDF, 0xA3}
	
	// AVI 魔数：RIFF...AVI
	v.magicNumbers["avi"] = []byte{0x52, 0x49, 0x46, 0x46} // RIFF
	
	// MOV 魔数：ftyp
	v.magicNumbers["mov"] = []byte{0x66, 0x74, 0x79, 0x70}
}

// ValidateFormat 验证视频格式
func (v *VideoValidator) ValidateFormat(request *ValidationRequest) (*ValidationResult, error) {
	// 验证输入参数
	if request.Filename == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}

	if len(request.Data) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}

	if len(request.Data) < 4 {
		return nil, fmt.Errorf("文件头信息不完整")
	}

	// 从文件名获取扩展名
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(request.Filename), "."))
	if !v.IsFormatSupported(ext) {
		return nil, fmt.Errorf("不支持的视频格式: %s", ext)
	}

	// 通过魔数检测实际格式
	detectedFormat, err := v.DetectFormatByMagicNumber(request.Data)
	if err != nil {
		return &ValidationResult{
			IsValid:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// 检查文件扩展名与检测到的格式是否匹配
	if ext != detectedFormat {
		return nil, fmt.Errorf("文件内容与扩展名不匹配：扩展名为 %s，但内容为 %s", ext, detectedFormat)
	}

	return &ValidationResult{
		IsValid:        true,
		DetectedFormat: detectedFormat,
	}, nil
}

// DetectFormatByMagicNumber 通过魔数检测文件格式
func (v *VideoValidator) DetectFormatByMagicNumber(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("数据长度不足以检测格式")
	}

	// 检测WebM格式（EBML header）
	if bytes.HasPrefix(data, v.magicNumbers["webm"]) {
		return "webm", nil
	}

	// 检测AVI格式（RIFF header）
	if bytes.HasPrefix(data, v.magicNumbers["avi"]) && len(data) >= 12 {
		// 进一步检查AVI标识
		if bytes.Equal(data[8:12], []byte{0x41, 0x56, 0x49, 0x20}) { // "AVI "
			return "avi", nil
		}
	}

	// 检测MP4和MOV格式（都使用FTYP box）
	if len(data) >= 12 {
		// 查找ftyp标识（可能在偏移4的位置）
		if bytes.Equal(data[4:8], v.magicNumbers["mp4"]) {
			// 检查文件类型标识符
			brand := data[8:12]
			
			// MP4品牌标识
			mp4Brands := [][]byte{
				{0x6D, 0x70, 0x34, 0x31}, // mp41
				{0x6D, 0x70, 0x34, 0x32}, // mp42
				{0x69, 0x73, 0x6F, 0x6D}, // isom
				{0x64, 0x61, 0x73, 0x68}, // dash
			}
			
			for _, mp4Brand := range mp4Brands {
				if bytes.Equal(brand, mp4Brand) {
					return "mp4", nil
				}
			}
			
			// MOV品牌标识
			movBrands := [][]byte{
				{0x71, 0x74, 0x20, 0x20}, // "qt  "
			}
			
			for _, movBrand := range movBrands {
				if bytes.Equal(brand, movBrand) {
					return "mov", nil
				}
			}
		}
	}

	return "", fmt.Errorf("无法识别的视频格式")
}

// ValidateFileSize 验证文件大小
func (v *VideoValidator) ValidateFileSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("文件大小无效")
	}

	if size == 0 {
		return fmt.Errorf("文件不能为空")
	}

	if size > v.maxFileSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %d MB", v.maxFileSize/(1024*1024))
	}

	return nil
}

// ValidateContentType 验证内容类型
func (v *VideoValidator) ValidateContentType(contentType string) error {
	if contentType == "" {
		return fmt.Errorf("内容类型不能为空")
	}

	_, exists := v.contentTypeMapping[contentType]
	if !exists {
		return fmt.Errorf("不支持的内容类型: %s", contentType)
	}

	return nil
}

// ComprehensiveValidation 综合验证
func (v *VideoValidator) ComprehensiveValidation(request *ComprehensiveValidationRequest) (*ComprehensiveValidationResult, error) {
	result := &ComprehensiveValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// 验证文件大小
	if err := v.ValidateFileSize(request.Size); err != nil {
		result.SizeValid = false
		result.Errors = append(result.Errors, err.Error())
		result.IsValid = false
	} else {
		result.SizeValid = true
	}

	// 验证内容类型
	if err := v.ValidateContentType(request.ContentType); err != nil {
		result.ContentTypeValid = false
		result.Errors = append(result.Errors, err.Error())
		result.IsValid = false
	} else {
		result.ContentTypeValid = true
	}

	// 验证格式
	formatRequest := &ValidationRequest{
		Filename:    request.Filename,
		ContentType: request.ContentType,
		Data:        request.Data,
	}

	formatResult, err := v.ValidateFormat(formatRequest)
	if err != nil {
		result.FormatValid = false
		result.Errors = append(result.Errors, err.Error())
		result.IsValid = false
	} else if !formatResult.IsValid {
		result.FormatValid = false
		result.Errors = append(result.Errors, formatResult.ErrorMessage)
		result.IsValid = false
	} else {
		result.FormatValid = true
		result.DetectedFormat = formatResult.DetectedFormat
	}

	return result, nil
}

// GetSupportedFormats 获取支持的格式列表
func (v *VideoValidator) GetSupportedFormats() []string {
	formats := make([]string, 0, len(v.supportedFormats))
	for format := range v.supportedFormats {
		formats = append(formats, format)
	}
	return formats
}

// GetMaxFileSize 获取最大文件大小限制
func (v *VideoValidator) GetMaxFileSize() int64 {
	return v.maxFileSize
}

// IsFormatSupported 检查格式是否支持
func (v *VideoValidator) IsFormatSupported(format string) bool {
	return v.supportedFormats[strings.ToLower(format)]
}