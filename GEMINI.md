# GEMINI.md

This file provides guidance to Gemini when working with code in this repository.

## 项目概述

烛龙（Zhulong）是一个局域网视频播放网站项目，目标是创建一个轻量级的本地视频流媒体服务。

## 技术栈

### 后端
- **框架**: CloudWeGo Hertz (Go HTTP框架)
- **特点**: 高性能、强扩展性的微服务框架
- **对象存储**: MinIO (S3兼容的对象存储)
- **数据库**: 待定（用于视频元数据）

### 前端  
- **框架**: React 18+
- **构建工具**: 待定（Vite或Webpack）
- **样式**: 待定（CSS Modules、Tailwind CSS或Styled Components）

### 开发方法
- **架构**: TDD (测试驱动开发)
- **先测试后实现的开发流程**

## 第一版本功能规划

### 核心功能
- **视频上传**: 支持常见视频格式（MP4、AVI、MOV等）
- **视频播放**: 基于HTML5的视频播放器
- **视频存储**: 使用MinIO对象存储管理视频文件
- **基础元数据**: 视频标题、时长、大小等基本信息

### API设计 (RESTful)
```
POST   /api/v1/videos          # 视频上传
GET    /api/v1/videos          # 获取视频列表
GET    /api/v1/videos/:id      # 获取单个视频信息
GET    /api/v1/videos/:id/play # 获取视频播放URL
DELETE /api/v1/videos/:id      # 删除视频
```

### MinIO集成
- **存储桶配置**: 默认桶名 `zhulong-videos`
- **文件路径**: `videos/{year}/{month}/{videoId}.{ext}`
- **预签名URL**: 用于安全的视频播放和下载
- **多部分上传**: 支持大文件分片上传

## 项目结构

```
zhulong/
├── backend/           # Hertz后端服务
│   ├── cmd/          # 入口文件
│   ├── pkg/          # 业务逻辑包
│   │   ├── config/   # 配置管理
│   │   ├── storage/  # MinIO存储层
│   │   └── models/   # 数据模型
│   ├── router/       # 路由配置
│   ├── handler/      # API处理器
│   ├── service/      # 业务服务层
│   ├── repository/   # 数据访问层
│   └── test/         # 测试文件
├── frontend/         # React前端应用
│   ├── src/
│   │   ├── components/  # React组件
│   │   │   ├── VideoUpload/    # 上传组件
│   │   │   ├── VideoPlayer/    # 播放器组件
│   │   │   └── VideoList/      # 列表组件
│   │   ├── hooks/       # 自定义Hooks
│   │   ├── services/    # API服务
│   │   ├── utils/       # 工具函数
│   │   └── __tests__/   # 测试文件
│   ├── public/       # 静态资源
│   └── package.json
├── docs/             # 项目文档
├── config/           # 配置文件目录
│   ├── README.md     # 配置说明
│   ├── app.yml       # 应用配置模板
│   ├── development.yml # 开发环境配置
│   └── production.yml  # 生产环境配置
├── .env              # 环境变量文件（项目根目录）
├── .env.example      # 环境变量模板
├── scripts/          # 构建脚本
└── compose.yml       # 服务器端开发环境配置
```

## 开发规范

### TDD开发流程
1. **红阶段**: 编写失败的测试用例
2. **绿阶段**: 编写最少代码使测试通过
3. **重构阶段**: 重构代码，保持测试通过
4. **循环**: 重复上述过程

### TODO管理流程
1. **任务开始前**: 
   - 在TODO.md中将任务状态从 `[ ]` 改为 `[*]`
   - 评估任务优先级和复杂度
2. **开发过程中**: 
   - 遵循TDD流程：先写测试，再实现功能
   - 确保代码通过所有测试
   - 进行代码审查
3. **任务完成后**: 
   - 将TODO状态从 `[*]` 改为 `[x]`
   - 更新项目总进度统计
   - 更新相关文档
   - 提交代码并推送到仓库

### Git提交规范
- 每完成一个TODO项目，进行一次提交
- 提交信息格式：`[TODO-ID] 简短描述`
- 例如：`[BACKEND-001] 初始化Go项目结构`

### 后端开发规范 (Hertz)
- 使用Hertz框架时，应使用hz工具进行代码生成，以统一编码规范。
- 使用标准的Go项目布局
- 遵循Hertz框架的最佳实践
- API设计遵循RESTful规范
- 使用结构体标签进行数据验证
- 统一的错误处理机制
- 日志记录使用结构化格式
- MinIO Go SDK集成最佳实践
- 使用预签名URL确保安全访问

### 前端开发规范 (React)
- 使用函数组件和Hooks
- 组件名使用PascalCase命名
- 文件名使用kebab-case命名
- 使用TypeScript提高代码质量
- 组件职责单一原则
- 状态管理优先使用内置hooks
- 使用Suspense处理异步加载

### Git分支管理规范
- 使用feature分支进行开发
- 提交信息使用约定式提交格式
- 代码review必须通过才能合并
- 保持提交历史清晰

### Git工作流程

#### 1. 分支策略
- **main**: 主分支，始终保持稳定可部署状态
- **develop**: 开发分支，集成所有feature分支
- **feature/xxx**: 功能分支，基于develop创建
- **hotfix/xxx**: 热修复分支，基于main创建

#### 2. 分支命名规范
```
feature/BACKEND-001-init-go-project
feature/FRONTEND-001-setup-react
hotfix/fix-video-upload-bug
release/v1.0.0
```

#### 3. 约定式提交规范 (Conventional Commits)
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**提交类型 (type):**
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档变更
- `style`: 代码格式调整（不影响功能）
- `refactor`: 重构（既不是新功能也不是修复）
- `perf`: 性能优化
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动
- `ci`: CI/CD配置文件和脚本的变更
- `revert`: 回滚先前的提交

**作用域 (scope):**
- `backend`: 后端相关
- `frontend`: 前端相关
- `api`: API接口
- `storage`: 存储相关
- `config`: 配置相关

**示例提交信息:**
```
feat(backend): 添加视频上传API接口

实现基于MinIO的视频文件上传功能，支持分片上传和断点续传

Closes #BACKEND-003
```

#### 4. Git初始化和配置
```bash
# 初始化仓库
git init

# 配置用户信息
git config user.name "Your Name"
git config user.email "your.email@example.com"

# 创建初始分支结构
git checkout -b main
git checkout -b develop
```

#### 5. 开发工作流
```bash
# 1. 从develop创建功能分支
git checkout develop
git pull origin develop
git checkout -b feature/BACKEND-001-init-go-project

# 2. 开发并提交（遵循约定式提交）
git add .
git commit -m "feat(backend): 初始化Go项目结构"

# 3. 推送到远程
git push -u origin feature/BACKEND-001-init-go-project

# 4. 创建Pull Request到develop分支
# 5. 代码审查通过后合并
# 6. 删除功能分支
git branch -d feature/BACKEND-001-init-go-project
```

#### 6. 发布流程
```bash
# 1. 从develop创建release分支
git checkout -b release/v1.0.0 develop

# 2. 生成CHANGELOG
npm run changelog

# 3. 提交版本信息
git add CHANGELOG.md
git commit -m "chore(release): v1.0.0"

# 4. 合并到main并打标签
git checkout main
git merge --no-ff release/v1.0.0
git tag -a v1.0.0 -m "Release version 1.0.0"

# 5. 合并回develop
git checkout develop
git merge --no-ff release/v1.0.0

# 6. 推送所有更改
git push origin main develop --tags
```

### 测试策略
- **单元测试**: 覆盖核心业务逻辑
- **集成测试**: 测试API接口
- **端到端测试**: 测试关键用户流程
- **测试覆盖率**: 保持80%以上

### 代码质量
- 使用gofmt格式化Go代码
- 使用Prettier格式化前端代码
- 使用ESLint进行代码检查
- 定期进行代码审查

## 常用命令

### 后端开发
```bash
# 运行开发服务器
go run cmd/main.go

# 运行测试
go test ./...

# 代码格式化
gofmt -w .

# 构建
go build -o bin/zhulong cmd/main.go

# 安装MinIO Go SDK
go get github.com/minio/minio-go/v7
```

### 前端开发
```bash
# 安装依赖
npm install

# 开发服务器
npm run dev

# 运行测试
npm test

# 构建生产版本
npm run build

# 代码格式化
npm run format
```

### Git和版本管理
```bash
# 安装changelog生成工具
npm install -D conventional-changelog-cli

# 配置package.json scripts
{
  "scripts": {
    "changelog": "conventional-changelog -p angular -i CHANGELOG.md -s",
    "changelog:first": "conventional-changelog -p angular -i CHANGELOG.md -s -r 0",
    "version": "conventional-changelog -p angular -i CHANGELOG.md -s && git add CHANGELOG.md"
  }
}

# 生成changelog（首次）
npm run changelog:first

# 生成增量changelog
npm run changelog

# 提交时自动生成changelog
npm version patch  # 自动运行version脚本
```

### MinIO相关（服务器端执行）
```bash
# 启动MinIO服务（在服务器上使用Podman）
podman run -d -p 9000:9000 -p 9001:9001 \
  -e "MINIO_ROOT_USER=admin" \
  -e "MINIO_ROOT_PASSWORD=password123" \
  --name minio-server \
  quay.io/minio/minio server /data --console-address ":9001"

# 创建存储桶（需要先安装mc客户端）
mc alias set local http://server-ip:9000 admin password123
mc mb local/zhulong-videos
```

### 开发环境（服务器端执行）
```bash
# 启动完整开发环境（在服务器上使用Podman Compose）
podman-compose up -d

# 停止开发环境
podman-compose down

# 查看容器状态
podman ps

# 查看容器日志
podman logs minio-server
```

## 重要说明

- 这是一个局域网应用，专注于本地网络环境下的视频播放服务
- 项目名称"烛龙"来自中国神话，寓意光明和守护
- 严格遵循TDD开发流程，确保代码质量
- 优先考虑性能和用户体验
- 代码风格保持一致性

## 部署和容器化

### 容器化要求
- **容器运行时**: 使用Podman替代Docker
- **执行位置**: 所有容器化操作在服务器端执行，不在本地开发机执行
- **配置文件**: 使用`compose.yml`而非`docker-compose.yml`
- **服务管理**: 通过Podman Compose管理多容器应用

### 服务器环境要求
- 安装Podman和Podman Compose
- 配置适当的网络和存储
- 确保服务器有足够的存储空间用于视频文件
- 配置防火墙允许必要端口访问（9000、9001等）

## 视频处理和存储策略

### 支持的视频格式
- **主要格式**: MP4 (推荐)、WebM、AVI
- **编码支持**: H.264、H.265/HEVC、VP9
- **音频编码**: AAC、MP3、Opus
- **文件大小限制**: 单文件最大2GB

### 存储策略
- **存储路径规则**: `videos/{year}/{month}/{uuid}.{ext}`
- **文件命名**: 使用UUID避免命名冲突
- **缩略图存储**: `thumbnails/{year}/{month}/{uuid}.jpg`
- **元数据缓存**: 使用内存缓存提高访问速度

### 安全和访问控制
- **预签名URL**: 所有视频访问通过预签名URL
- **URL过期时间**: 播放链接有效期24小时
- **访问日志**: 记录所有文件访问操作
- **IP限制**: 仅限局域网IP访问

### 性能优化
- **分片上传**: 大文件支持分片上传和断点续传
- **并发处理**: 支持多文件同时上传
- **带宽限制**: 根据网络状况自动调节传输速度
- **视频预加载**: 智能预加载提升播放体验

```