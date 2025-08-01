# Zhulong Backend Service

烛龙项目后端服务，基于 CloudWeGo Hertz 框架和 hz 工具生成。

## 项目结构

```
backend/
├── biz/                    # 业务逻辑层（hz生成）
│   ├── handler/           # HTTP处理器
│   │   └── zhulong/api/   # API处理器实现
│   ├── model/             # 数据模型
│   │   └── zhulong/api/   # Thrift生成的模型
│   └── router/            # 路由配置
├── pkg/                   # 项目公共包（手动维护）
│   ├── config/           # 配置管理
│   ├── storage/          # MinIO存储层
│   ├── utils/            # 工具函数
│   └── middleware/       # 中间件
├── script/               # 脚本文件
│   └── bootstrap.sh      # 启动脚本
├── output/               # 构建输出
│   └── bin/             # 可执行文件
├── build.sh             # 构建脚本（hz生成）
├── main.go              # 入口文件（hz生成）
├── router.go            # 路由注册（hz生成）
├── router_gen.go        # 路由生成（hz生成）
└── go.mod               # Go模块文件
```

## 生成的API接口

### VideoService
- `POST /api/v1/videos` - 视频上传
- `GET /api/v1/videos` - 获取视频列表
- `GET /api/v1/videos/:video_id` - 获取视频详情
- `GET /api/v1/videos/:video_id/play` - 获取视频播放URL
- `DELETE /api/v1/videos/:video_id` - 删除视频

### SystemService
- `GET /health` - 健康检查
- `GET /api/v1/info` - 服务器信息

## 快速开始

### 1. 构建项目
```bash
bash build.sh
```

### 2. 运行服务
```bash
bash output/bootstrap.sh
```

### 3. 开发模式运行
```bash
go run .
```

## 开发说明

### 代码生成规则
- **biz/handler/**: 处理器实现，可以手动修改
- **biz/model/**: 数据模型，由thriftgo生成，**不要手动修改**
- **biz/router/**: 路由配置，部分由hz生成
- **main.go, router*.go**: 由hz生成，一般不需要修改

### 业务逻辑实现
1. 在 `biz/handler/` 中实现具体的业务逻辑
2. 在 `pkg/` 目录下添加自定义的业务包
3. 按需在 `pkg/` 下实现配置管理、存储服务等

### IDL更新流程
当IDL文件变更时：
```bash
hz update --idl=../idl/zhulong_api.thrift
```

### 注意事项
- hz生成的代码有注释标记，避免手动修改
- 业务逻辑应该在handler中实现或调用pkg中的服务
- 保持IDL文件和实现的一致性

## 依赖管理

项目使用Go Modules管理依赖：
```bash
go mod tidy    # 整理依赖
go mod download # 下载依赖
```

## 测试

```bash
go test ./...  # 运行所有测试
```