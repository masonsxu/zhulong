namespace go zhulong.api

// 基础响应结构
struct BaseResponse {
    1: i32 code = 0
    2: string message = ""
    3: optional string trace_id = ""
}

// 视频信息结构
struct Video {
    1: string id = ""                      // 视频唯一标识
    2: string title = ""                   // 视频标题
    3: string filename = ""                // 原始文件名
    4: string content_type = ""            // MIME类型
    5: i64 size = 0                        // 文件大小（字节）
    6: i64 duration = 0                    // 视频时长（秒）
    7: i32 width = 0                       // 视频宽度
    8: i32 height = 0                      // 视频高度
    9: string storage_path = ""            // 存储路径
    10: optional string thumbnail_path = "" // 缩略图路径
    11: i64 uploaded_at = 0                // 上传时间戳（毫秒）
    12: i64 updated_at = 0                 // 更新时间戳（毫秒）
}

// 视频上传请求
struct VideoUploadRequest {
    1: string title                        // 视频标题（必填）
    2: optional string description = ""    // 视频描述
}

// 视频上传响应
struct VideoUploadResponse {
    1: BaseResponse base
    2: optional Video video
    3: optional string upload_url = ""     // 预签名上传URL
}

// 视频列表请求
struct VideoListRequest {
    1: optional i32 page = 1               // 页码，默认第1页
    2: optional i32 page_size = 20         // 每页大小，默认20条
    3: optional string search = ""         // 搜索关键词
    4: optional string sort_by = "uploaded_at" // 排序字段
    5: optional string sort_order = "desc" // 排序方向：asc/desc
}

// 视频列表响应
struct VideoListResponse {
    1: BaseResponse base
    2: list<Video> videos = []
    3: i32 total = 0                       // 总数量
    4: i32 page = 1                        // 当前页码
    5: i32 page_size = 20                  // 每页大小
    6: i32 total_pages = 0                 // 总页数
}

// 视频详情请求
struct VideoDetailRequest {
    1: string video_id                     // 视频ID
}

// 视频详情响应
struct VideoDetailResponse {
    1: BaseResponse base
    2: optional Video video
}

// 视频播放URL请求
struct VideoPlayURLRequest {
    1: string video_id                     // 视频ID
    2: optional i32 expire_seconds = 3600  // URL过期时间（秒），默认1小时
}

// 视频播放URL响应
struct VideoPlayURLResponse {
    1: BaseResponse base
    2: optional string play_url = ""       // 播放URL
    3: optional i64 expires_at = 0         // URL过期时间戳（毫秒）
}

// 视频删除请求
struct VideoDeleteRequest {
    1: string video_id                     // 视频ID
}

// 视频删除响应
struct VideoDeleteResponse {
    1: BaseResponse base
}

// 健康检查响应
struct HealthCheckResponse {
    1: BaseResponse base
    2: string status = "ok"
    3: string service = "zhulong-backend"
    4: string version = "v1.0.0"
    5: i64 timestamp = 0                   // 当前时间戳（毫秒）
}

// 服务器信息响应
struct ServerInfoResponse {
    1: BaseResponse base
    2: string name = "Zhulong Video Server"
    3: string description = "局域网视频播放服务后端"
    4: string version = "v1.0.0"
    5: string framework = "CloudWeGo Hertz"
    6: map<string, string> capabilities = {} // 服务能力
}

// 视频服务接口定义
service VideoService {
    // 视频上传接口
    VideoUploadResponse UploadVideo(1: VideoUploadRequest req) (api.post="/api/v1/videos")
    
    // 获取视频列表
    VideoListResponse GetVideoList(1: VideoListRequest req) (api.get="/api/v1/videos")
    
    // 获取视频详情
    VideoDetailResponse GetVideoDetail(1: VideoDetailRequest req) (api.get="/api/v1/videos/:video_id")
    
    // 获取视频播放URL
    VideoPlayURLResponse GetVideoPlayURL(1: VideoPlayURLRequest req) (api.get="/api/v1/videos/:video_id/play")
    
    // 删除视频
    VideoDeleteResponse DeleteVideo(1: VideoDeleteRequest req) (api.delete="/api/v1/videos/:video_id")
}

// 系统服务接口定义
service SystemService {
    // 健康检查
    HealthCheckResponse HealthCheck() (api.get="/health")
    
    // 服务器信息
    ServerInfoResponse GetServerInfo() (api.get="/api/v1/info")
}