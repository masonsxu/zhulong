package test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/manteia/zhulong/router"
)

// TestHealthCheckEndpoint 测试健康检查端点
func TestHealthCheckEndpoint(t *testing.T) {
	h := router.SetupRouter()
	
	w := ut.PerformRequest(h.Engine, "GET", "/health", nil)
	
	assert.DeepEqual(t, http.StatusOK, w.Code)
	assert.True(t, len(w.Body.String()) > 0)
	assert.True(t, strings.Contains(w.Body.String(), "zhulong-backend"))
}

// TestServerInfoEndpoint 测试服务器信息端点
func TestServerInfoEndpoint(t *testing.T) {
	h := router.SetupRouter()
	
	w := ut.PerformRequest(h.Engine, "GET", "/api/v1/info", nil)
	
	assert.DeepEqual(t, http.StatusOK, w.Code)
	assert.True(t, len(w.Body.String()) > 0)
	assert.True(t, strings.Contains(w.Body.String(), "Zhulong Video Server"))
}

// TestCORSMiddleware 测试CORS中间件
func TestCORSMiddleware(t *testing.T) {
	h := router.SetupRouter()
	
	// 测试普通GET请求是否正常工作（CORS中间件不应该影响正常请求）
	w := ut.PerformRequest(h.Engine, "GET", "/api/v1/info", nil)
	
	// 确保正常请求工作正常
	assert.DeepEqual(t, http.StatusOK, w.Code)
}

// TestBasicHTTPHandling 测试基础HTTP处理
func TestBasicHTTPHandling(t *testing.T) {
	h := router.SetupRouter()
	
	w := ut.PerformRequest(h.Engine, "GET", "/health", nil)
	
	// 测试请求能正常处理即可
	assert.DeepEqual(t, http.StatusOK, w.Code)
}