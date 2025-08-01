package metadata

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetadataService_SaveMetadata 测试保存文件元数据
func TestMetadataService_SaveMetadata(t *testing.T) {
	metadataService := NewMetadataService()

	// 创建测试元数据
	metadata := &FileMetadata{
		FileID:      "test-file-001",
		BucketName:  "test-bucket",
		ObjectName:  "videos/2025/08/test-video.mp4",
		FileName:    "test-video.mp4",
		FileSize:    1024 * 1024 * 10, // 10MB
		ContentType: "video/mp4",
		Title:       "测试视频",
		Description: "这是一个测试视频文件",
		Tags:        []string{"测试", "视频", "demo"},
		Duration:    300, // 5分钟
		Resolution:  "1920x1080",
		Bitrate:     2500,
		CreatedBy:   "test-user",
	}

	ctx := context.Background()

	// 测试保存元数据
	err := metadataService.SaveMetadata(ctx, metadata)
	assert.NoError(t, err, "保存元数据应该成功")

	// 验证元数据已保存
	savedMetadata, err := metadataService.GetMetadata(ctx, metadata.FileID)
	assert.NoError(t, err, "获取元数据应该成功")
	require.NotNil(t, savedMetadata, "保存的元数据不应为空")

	// 验证保存的数据
	assert.Equal(t, metadata.FileID, savedMetadata.FileID, "文件ID应该匹配")
	assert.Equal(t, metadata.Title, savedMetadata.Title, "标题应该匹配")
	assert.Equal(t, metadata.Description, savedMetadata.Description, "描述应该匹配")
	assert.ElementsMatch(t, metadata.Tags, savedMetadata.Tags, "标签应该匹配")
	assert.Equal(t, metadata.Duration, savedMetadata.Duration, "时长应该匹配")
	assert.NotZero(t, savedMetadata.CreatedAt, "创建时间不应为零值")
}

// TestMetadataService_UpdateMetadata 测试更新文件元数据
func TestMetadataService_UpdateMetadata(t *testing.T) {
	metadataService := NewMetadataService()

	// 先保存一个元数据
	originalMetadata := &FileMetadata{
		FileID:      "test-file-002",
		BucketName:  "test-bucket",
		ObjectName:  "videos/2025/08/update-test.mp4",
		FileName:    "update-test.mp4",
		Title:       "原始标题",
		Description: "原始描述",
		Tags:        []string{"原始", "标签"},
		CreatedBy:   "test-user",
	}

	ctx := context.Background()
	err := metadataService.SaveMetadata(ctx, originalMetadata)
	require.NoError(t, err)

	// 更新元数据
	updateRequest := &UpdateMetadataRequest{
		FileID:      "test-file-002",
		Title:       stringPtr("更新后的标题"),
		Description: stringPtr("更新后的描述"),
		Tags:        &[]string{"更新", "标签", "新增"},
	}

	err = metadataService.UpdateMetadata(ctx, updateRequest)
	assert.NoError(t, err, "更新元数据应该成功")

	// 验证更新结果
	updatedMetadata, err := metadataService.GetMetadata(ctx, "test-file-002")
	assert.NoError(t, err, "获取更新后的元数据应该成功")
	require.NotNil(t, updatedMetadata, "更新后的元数据不应为空")

	assert.Equal(t, "更新后的标题", updatedMetadata.Title, "标题应该已更新")
	assert.Equal(t, "更新后的描述", updatedMetadata.Description, "描述应该已更新")
	assert.ElementsMatch(t, []string{"更新", "标签", "新增"}, updatedMetadata.Tags, "标签应该已更新")
	assert.True(t, updatedMetadata.UpdatedAt.After(updatedMetadata.CreatedAt), "更新时间应该晚于创建时间")
}

// TestMetadataService_DeleteMetadata 测试删除文件元数据
func TestMetadataService_DeleteMetadata(t *testing.T) {
	metadataService := NewMetadataService()

	// 先保存一个元数据
	metadata := &FileMetadata{
		FileID:     "test-file-003",
		BucketName: "test-bucket",
		ObjectName: "videos/2025/08/delete-test.mp4",
		FileName:   "delete-test.mp4",
		Title:      "待删除的文件",
		CreatedBy:  "test-user",
	}

	ctx := context.Background()
	err := metadataService.SaveMetadata(ctx, metadata)
	require.NoError(t, err)

	// 验证元数据存在
	_, err = metadataService.GetMetadata(ctx, metadata.FileID)
	assert.NoError(t, err, "删除前应该能获取到元数据")

	// 删除元数据
	err = metadataService.DeleteMetadata(ctx, metadata.FileID)
	assert.NoError(t, err, "删除元数据应该成功")

	// 验证元数据已删除
	deletedMetadata, err := metadataService.GetMetadata(ctx, metadata.FileID)
	assert.Error(t, err, "删除后获取元数据应该失败")
	assert.Nil(t, deletedMetadata, "删除后元数据应为空")
	assert.Contains(t, err.Error(), "元数据不存在", "错误信息应该表明元数据不存在")
}

// TestMetadataService_SearchMetadata 测试搜索文件元数据
func TestMetadataService_SearchMetadata(t *testing.T) {
	metadataService := NewMetadataService()
	ctx := context.Background()

	// 准备测试数据
	testMetadata := []*FileMetadata{
		{
			FileID:     "search-001",
			Title:      "Python教程视频",
			Tags:       []string{"Python", "编程", "教程"},
			Duration:   1800,
			CreatedBy:  "teacher1",
		},
		{
			FileID:     "search-002",
			Title:      "Go语言入门",
			Tags:       []string{"Go", "编程", "入门"},
			Duration:   2400,
			CreatedBy:  "teacher2",
		},
		{
			FileID:     "search-003",
			Title:      "JavaScript高级特性",
			Tags:       []string{"JavaScript", "编程", "高级"},
			Duration:   3600,
			CreatedBy:  "teacher1",
		},
	}

	// 保存测试数据
	for _, metadata := range testMetadata {
		err := metadataService.SaveMetadata(ctx, metadata)
		require.NoError(t, err)
	}

	// 测试按标题搜索
	searchRequest := &SearchMetadataRequest{
		Query: "Python",
		Limit: 10,
	}

	results, err := metadataService.SearchMetadata(ctx, searchRequest)
	assert.NoError(t, err, "搜索应该成功")
	require.NotNil(t, results, "搜索结果不应为空")
	assert.Len(t, results.Items, 1, "应该找到1个匹配的结果")
	assert.Equal(t, "search-001", results.Items[0].FileID, "应该找到Python教程视频")

	// 测试按标签搜索
	searchRequest = &SearchMetadataRequest{
		Tags:  []string{"编程"},
		Limit: 10,
	}

	results, err = metadataService.SearchMetadata(ctx, searchRequest)
	assert.NoError(t, err, "按标签搜索应该成功")
	assert.Len(t, results.Items, 3, "应该找到3个包含'编程'标签的结果")

	// 测试按创建者搜索
	searchRequest = &SearchMetadataRequest{
		CreatedBy: "teacher1",
		Limit:     10,
	}

	results, err = metadataService.SearchMetadata(ctx, searchRequest)
	assert.NoError(t, err, "按创建者搜索应该成功")
	assert.Len(t, results.Items, 2, "应该找到teacher1创建的2个视频")
}

// TestMetadataService_ListMetadata 测试列出文件元数据
func TestMetadataService_ListMetadata(t *testing.T) {
	metadataService := NewMetadataService()
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 15; i++ {
		metadata := &FileMetadata{
			FileID:    fmt.Sprintf("list-test-%03d", i),
			Title:     fmt.Sprintf("测试视频 %d", i),
			Duration:  300 + int64(i*60), // 递增的时长
			CreatedBy: "test-user",
		}
		err := metadataService.SaveMetadata(ctx, metadata)
		require.NoError(t, err)
	}

	// 测试分页列表
	listRequest := &ListMetadataRequest{
		Offset: 0,
		Limit:  10,
		SortBy: "duration",
		Order:  "asc",
	}

	results, err := metadataService.ListMetadata(ctx, listRequest)
	assert.NoError(t, err, "列表查询应该成功")
	require.NotNil(t, results, "列表结果不应为空")
	assert.Len(t, results.Items, 10, "应该返回10个结果")
	assert.Equal(t, 15, results.Total, "总数应该是15")

	// 验证排序
	for i := 1; i < len(results.Items); i++ {
		assert.True(t, results.Items[i-1].Duration <= results.Items[i].Duration, "应该按时长升序排列")
	}

	// 测试第二页
	listRequest.Offset = 10
	results, err = metadataService.ListMetadata(ctx, listRequest)
	assert.NoError(t, err, "第二页查询应该成功")
	assert.Len(t, results.Items, 5, "第二页应该返回5个结果")
}

// TestMetadataService_GetMetadataByObjectName 测试根据对象名获取元数据
func TestMetadataService_GetMetadataByObjectName(t *testing.T) {
	metadataService := NewMetadataService()

	metadata := &FileMetadata{
		FileID:     "object-test-001",
		BucketName: "test-bucket",
		ObjectName: "videos/2025/08/unique-object.mp4",
		Title:      "通过对象名查找的视频",
		CreatedBy:  "test-user",
	}

	ctx := context.Background()
	err := metadataService.SaveMetadata(ctx, metadata)
	require.NoError(t, err)

	// 测试根据对象名获取元数据
	foundMetadata, err := metadataService.GetMetadataByObjectName(ctx, metadata.BucketName, metadata.ObjectName)
	assert.NoError(t, err, "根据对象名获取元数据应该成功")
	require.NotNil(t, foundMetadata, "找到的元数据不应为空")
	assert.Equal(t, metadata.FileID, foundMetadata.FileID, "文件ID应该匹配")
	assert.Equal(t, metadata.Title, foundMetadata.Title, "标题应该匹配")
}

// TestMetadataService_AddTags 测试添加标签
func TestMetadataService_AddTags(t *testing.T) {
	metadataService := NewMetadataService()

	metadata := &FileMetadata{
		FileID:    "tags-test-001",
		Title:     "标签测试视频",
		Tags:      []string{"原始", "标签"},
		CreatedBy: "test-user",
	}

	ctx := context.Background()
	err := metadataService.SaveMetadata(ctx, metadata)
	require.NoError(t, err)

	// 添加新标签
	err = metadataService.AddTags(ctx, metadata.FileID, []string{"新增", "标签", "测试"})
	assert.NoError(t, err, "添加标签应该成功")

	// 验证标签已添加
	updatedMetadata, err := metadataService.GetMetadata(ctx, metadata.FileID)
	assert.NoError(t, err, "获取更新后的元数据应该成功")
	expectedTags := []string{"原始", "标签", "新增", "测试"} // 去重后的标签
	assert.ElementsMatch(t, expectedTags, updatedMetadata.Tags, "标签应该包含新添加的标签")
}

// TestMetadataService_RemoveTags 测试移除标签
func TestMetadataService_RemoveTags(t *testing.T) {
	metadataService := NewMetadataService()

	metadata := &FileMetadata{
		FileID:    "remove-tags-001",
		Title:     "移除标签测试视频",
		Tags:      []string{"标签1", "标签2", "标签3", "标签4"},
		CreatedBy: "test-user",
	}

	ctx := context.Background()
	err := metadataService.SaveMetadata(ctx, metadata)
	require.NoError(t, err)

	// 移除部分标签
	err = metadataService.RemoveTags(ctx, metadata.FileID, []string{"标签2", "标签4"})
	assert.NoError(t, err, "移除标签应该成功")

	// 验证标签已移除
	updatedMetadata, err := metadataService.GetMetadata(ctx, metadata.FileID)
	assert.NoError(t, err, "获取更新后的元数据应该成功")
	expectedTags := []string{"标签1", "标签3"}
	assert.ElementsMatch(t, expectedTags, updatedMetadata.Tags, "应该只剩下未移除的标签")
}

// TestMetadataService_ValidateMetadata 测试元数据验证
func TestMetadataService_ValidateMetadata(t *testing.T) {
	metadataService := NewMetadataService()

	// 测试有效的元数据
	validMetadata := &FileMetadata{
		FileID:     "valid-001",
		BucketName: "test-bucket",
		ObjectName: "videos/2025/08/valid.mp4",
		FileName:   "valid.mp4",
		Title:      "有效的视频",
		CreatedBy:  "test-user",
	}

	err := metadataService.ValidateMetadata(validMetadata)
	assert.NoError(t, err, "有效的元数据应该通过验证")

	// 测试无效的元数据
	invalidCases := []struct {
		name     string
		metadata *FileMetadata
		errMsg   string
	}{
		{
			name: "空文件ID",
			metadata: &FileMetadata{
				FileID:    "",
				Title:     "测试",
				CreatedBy: "user",
			},
			errMsg: "文件ID不能为空",
		},
		{
			name: "空标题",
			metadata: &FileMetadata{
				FileID:    "test-001",
				Title:     "",
				CreatedBy: "user",
			},
			errMsg: "标题不能为空",
		},
		{
			name: "空创建者",
			metadata: &FileMetadata{
				FileID:    "test-001",
				Title:     "测试",
				CreatedBy: "",
			},
			errMsg: "创建者不能为空",
		},
		{
			name: "标题过长",
			metadata: &FileMetadata{
				FileID:    "test-001",
				Title:     string(make([]byte, 256)), // 256个字符，超过限制
				CreatedBy: "user",
			},
			errMsg: "标题长度不能超过255个字符",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := metadataService.ValidateMetadata(tc.metadata)
			assert.Error(t, err, "无效的元数据应该验证失败")
			assert.Contains(t, err.Error(), tc.errMsg, "错误信息应该包含预期内容")
		})
	}
}

// stringPtr 辅助函数，返回字符串指针
func stringPtr(s string) *string {
	return &s
}