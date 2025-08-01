package main

import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 创建 Hertz 服务器实例
	h := server.Default()

	// 基础健康检查端点
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]string{
			"status": "ok",
			"service": "zhulong-backend",
		})
	})

	// 启动服务器
	log.Println("Zhulong backend server starting on :8080")
	h.Spin()
}