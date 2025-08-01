# Backend

烛龙项目后端服务，基于CloudWeGo Hertz框架开发。

## 目录结构

```
backend/
├── cmd/                    # 应用入口
│   ├── main/              # 主应用
│   └── migrate/           # 数据迁移工具
├── pkg/                   # 业务逻辑包
│   ├── config/           # 配置管理
│   ├── models/           # 数据模型
│   └── storage/          # 存储层
├── router/               # 路由配置
├── handler/              # HTTP处理器
├── service/              # 业务服务层
├── repository/           # 数据访问层
└── test/                 # 测试文件
```

## 快速开始

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/main/main.go

# 运行测试
go test ./...
```