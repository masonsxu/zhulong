# IDL 文件说明

本目录包含烛龙项目的 Thrift IDL（Interface Definition Language）定义文件。

## 文件结构

- `zhulong_api.thrift` - 主要API接口定义文件

## Thrift IDL 设计原则

### 1. 命名规范
- **命名空间**: 使用 `namespace go zhulong.api` 定义Go语言命名空间
- **结构体**: 使用 PascalCase 命名（如 `VideoUploadRequest`）
- **字段**: 使用 snake_case 命名（如 `video_id`）
- **服务**: 使用 PascalCase 命名（如 `VideoService`）

### 2. 数据结构设计
- **响应统一性**: 所有响应都包含 `BaseResponse` 基础结构
- **可选字段**: 使用 `optional` 标记非必需字段
- **默认值**: 为字段提供合理的默认值
- **类型安全**: 使用明确的数据类型（i32, i64, string等）

### 3. API路由映射
- 使用 Thrift 注解定义 HTTP 路由：`(api.get="/path")`, `(api.post="/path")` 等
- 路径参数使用冒号语法：`:video_id`
- 遵循 RESTful 设计原则

### 4. 版本管理
- API路径包含版本号：`/api/v1/`
- 结构体字段使用数字标识符，便于向后兼容
- 新增字段应使用 `optional` 标记

## 服务分组

### VideoService
负责视频相关的核心功能：
- 视频上传
- 视频列表查询
- 视频详情获取
- 视频播放URL生成
- 视频删除

### SystemService  
负责系统级功能：
- 健康检查
- 服务器信息查询

## 数据模型

### Video
核心视频数据结构，包含：
- 基础信息：ID、标题、文件名
- 技术信息：格式、大小、时长、分辨率
- 存储信息：存储路径、缩略图路径
- 时间信息：上传时间、更新时间

### 请求/响应对
每个API操作都有对应的请求和响应结构：
- `*Request` - 请求参数
- `*Response` - 响应数据，包含 BaseResponse 和具体数据

## 使用方法

1. **安装 hz 工具**:
   ```bash
   go install github.com/cloudwego/hertz/cmd/hz@latest
   ```

2. **生成代码**:
   ```bash
   hz new --idl=idl/zhulong_api.thrift --mod=github.com/yourusername/zhulong backend
   ```

3. **更新代码**（IDL变更后）:
   ```bash
   hz update --idl=idl/zhulong_api.thrift
   ```

## 注意事项

1. **字段标识符**: Thrift 字段必须有唯一的数字标识符，不要随意修改
2. **向后兼容**: 新增字段使用 `optional`，避免删除已有字段
3. **类型变更**: 避免修改已有字段的类型，考虑新增字段替代
4. **路由设计**: 保持 RESTful 风格，使用合适的 HTTP 方法