package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// CORSConfig CORS配置
type CORSConfig struct {
	// 允许的源域名
	AllowOrigins []string
	// 允许的HTTP方法
	AllowMethods []string
	// 允许的请求头
	AllowHeaders []string
	// 允许的响应头
	ExposeHeaders []string
	// 是否允许携带认证信息
	AllowCredentials bool
	// 预检请求缓存时间(秒)
	MaxAge int
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",  // React开发服务器
			"http://localhost:5173",  // Vite开发服务器
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		},
		AllowMethods: []string{
			consts.MethodGet,
			consts.MethodPost,
			consts.MethodPut,
			consts.MethodDelete,
			consts.MethodOptions,
			consts.MethodHead,
			consts.MethodPatch,
		},
		AllowHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Type",
			"Content-Language",
			"Origin",
			"Cache-Control",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Content-Range",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
	}
}

// LocalNetworkCORSConfig 局域网CORS配置
func LocalNetworkCORSConfig() *CORSConfig {
	config := DefaultCORSConfig()
	// 允许局域网所有IP访问
	config.AllowOrigins = []string{
		"*", // 开发阶段允许所有源，生产环境应该配置具体的域名
	}
	return config
}

// CORS 创建CORS中间件
func CORS(config ...*CORSConfig) app.HandlerFunc {
	var cfg *CORSConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultCORSConfig()
	}

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		method := string(c.Method())

		// 检查Origin是否被允许
		if isAllowedOrigin(origin, cfg.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if contains(cfg.AllowOrigins, "*") {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 设置允许的方法
		if len(cfg.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods, ", "))
		}

		// 设置允许的请求头
		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders, ", "))
		}

		// 设置暴露的响应头
		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders, ", "))
		}

		// 设置是否允许携带认证信息
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求缓存时间
		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", intToString(cfg.MaxAge))
		}

		// 处理预检请求
		if method == consts.MethodOptions {
			c.Status(consts.StatusNoContent)
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next(ctx)
	}
}

// isAllowedOrigin 检查源是否被允许
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// 这里可以添加更复杂的匹配逻辑，比如通配符匹配
	}
	return false
}

// contains 检查字符串数组是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// joinStrings 连接字符串数组
func joinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return slice[0]
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += separator + slice[i]
	}
	return result
}

// intToString 整数转字符串
func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	
	var result string
	for i > 0 {
		digit := i % 10
		result = string(rune('0'+digit)) + result
		i /= 10
	}
	return result
}