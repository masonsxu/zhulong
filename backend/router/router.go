package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

// SetupRouter 配置路由和中间件
func SetupRouter() *server.Hertz {
	// 创建Hertz实例
	h := server.Default()

	// 配置CORS中间件
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 开发环境允许所有源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// 注册路由
	registerRoutes(h)

	return h
}

// registerRoutes 注册所有路由
func registerRoutes(h *server.Hertz) {
	// 健康检查端点
	h.GET("/health", healthCheckHandler)

	// API版本1路由组
	v1 := h.Group("/api/v1")
	{
		// 基础信息接口
		v1.GET("/info", serverInfoHandler)
	}
}

// healthCheckHandler 健康检查处理器
func healthCheckHandler(ctx context.Context, c *app.RequestContext) {
	c.JSON(200, map[string]interface{}{
		"status":  "ok",
		"service": "zhulong-backend",
		"version": "v1.0.0",
	})
}

// serverInfoHandler 服务器信息处理器
func serverInfoHandler(ctx context.Context, c *app.RequestContext) {
	c.JSON(200, map[string]interface{}{
		"name":        "Zhulong Video Server",
		"description": "局域网视频播放服务后端",
		"version":     "v1.0.0",
		"framework":   "CloudWeGo Hertz",
	})
}