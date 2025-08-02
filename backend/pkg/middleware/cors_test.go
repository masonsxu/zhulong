package middleware

import (
	"testing"
)

func TestCORS_DefaultConfig(t *testing.T) {
	config := DefaultCORSConfig()
	if config == nil {
		t.Error("DefaultCORSConfig应该返回非nil配置")
	}
	if len(config.AllowOrigins) == 0 {
		t.Error("AllowOrigins不应为空")
	}
	if len(config.AllowMethods) == 0 {
		t.Error("AllowMethods不应为空")
	}
	if len(config.AllowHeaders) == 0 {
		t.Error("AllowHeaders不应为空")
	}
	if !config.AllowCredentials {
		t.Error("AllowCredentials应该为true")
	}
	if config.MaxAge <= 0 {
		t.Error("MaxAge应该大于0")
	}
}

func TestCORS_LocalNetworkConfig(t *testing.T) {
	config := LocalNetworkCORSConfig()
	if config == nil {
		t.Error("LocalNetworkCORSConfig应该返回非nil配置")
	}
	
	found := false
	for _, origin := range config.AllowOrigins {
		if origin == "*" {
			found = true
			break
		}
	}
	if !found {
		t.Error("LocalNetworkConfig应该包含通配符源")
	}
}

func TestUtils(t *testing.T) {
	t.Run("isAllowedOrigin", func(t *testing.T) {
		origins := []string{"http://localhost:3000", "https://example.com"}
		if !isAllowedOrigin("http://localhost:3000", origins) {
			t.Error("应该允许localhost:3000")
		}
		if isAllowedOrigin("http://malicious.com", origins) {
			t.Error("不应该允许malicious.com")
		}
		if isAllowedOrigin("", origins) {
			t.Error("不应该允许空源")
		}
	})

	t.Run("contains", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		if !contains(slice, "b") {
			t.Error("应该包含b")
		}
		if contains(slice, "d") {
			t.Error("不应该包含d")
		}
	})

	t.Run("joinStrings", func(t *testing.T) {
		if joinStrings([]string{}, ", ") != "" {
			t.Error("空数组应该返回空字符串")
		}
		if joinStrings([]string{"a"}, ", ") != "a" {
			t.Error("单元素数组应该返回该元素")
		}
		if joinStrings([]string{"a", "b", "c"}, ", ") != "a, b, c" {
			t.Error("多元素数组应该正确连接")
		}
	})

	t.Run("intToString", func(t *testing.T) {
		if intToString(0) != "0" {
			t.Error("0应该转换为'0'")
		}
		if intToString(123) != "123" {
			t.Error("123应该转换为'123'")
		}
		if intToString(86400) != "86400" {
			t.Error("86400应该转换为'86400'")
		}
	})
}