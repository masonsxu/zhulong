package main

import (
	"log"

	"github.com/manteia/zhulong/router"
)

func main() {
	// 使用路由配置创建服务器
	h := router.SetupRouter()

	// 启动服务器
	log.Println("Zhulong backend server starting on :8080")
	h.Spin()
}