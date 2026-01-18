package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestRateLimitConfig_Defaults 测试配置默认值
func TestRateLimitConfig_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		config   RateLimitConfig
		limit    int
		window   time.Duration
		errorMsg string
	}{
		{
			name:     "GlobalRateLimitConfig",
			config:   GlobalRateLimitConfig(),
			limit:    100,
			window:   time.Second,
			errorMsg: "Rate limit exceeded",
		},
		{
			name:     "EndpointRateLimitConfig",
			config:   EndpointRateLimitConfig(10, time.Minute),
			limit:    10,
			window:   time.Minute,
			errorMsg: "Too many requests",
		},
		{
			name:     "UserRateLimitConfig",
			config:   UserRateLimitConfig(50, 5*time.Minute),
			limit:    50,
			window:   5 * time.Minute,
			errorMsg: "User rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Limit != tt.limit {
				t.Errorf("Limit = %d, want %d", tt.config.Limit, tt.limit)
			}
			if tt.config.Window != tt.window {
				t.Errorf("Window = %v, want %v", tt.config.Window, tt.window)
			}
			if tt.config.ErrorMsg != tt.errorMsg {
				t.Errorf("ErrorMsg = %s, want %s", tt.config.ErrorMsg, tt.errorMsg)
			}
		})
	}
}

// TestIPKeyFunc 测试 IP key 函数
func TestIPKeyFunc(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.1:12345"

	key, ok := IPKeyFunc(c)
	if !ok {
		t.Error("IPKeyFunc should return ok=true")
	}
	if key == "" {
		t.Error("IPKeyFunc should return non-empty key")
	}
}

// TestIPEndpointKeyFunc 测试 IP+端点 key 函数
func TestIPEndpointKeyFunc(t *testing.T) {
	router := gin.New()
	router.GET("/api/test", func(c *gin.Context) {
		key, ok := IPEndpointKeyFunc(c)
		if !ok {
			t.Error("IPEndpointKeyFunc should return ok=true")
		}
		if key == "" {
			t.Error("IPEndpointKeyFunc should return non-empty key")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)
}

// TestUserKeyFunc 测试用户 key 函数
func TestUserKeyFunc(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		wantOK   bool
		wantKey  string
	}{
		{
			name:    "authenticated user",
			userID:  123,
			wantOK:  true,
			wantKey: "user:123",
		},
		{
			name:   "unauthenticated user",
			userID: 0,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.userID > 0 {
				c.Set("user_id", tt.userID)
			}

			key, ok := UserKeyFunc(c)
			if ok != tt.wantOK {
				t.Errorf("UserKeyFunc ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && key != tt.wantKey {
				t.Errorf("UserKeyFunc key = %s, want %s", key, tt.wantKey)
			}
		})
	}
}

// TestCreateRateLimiter_AllowsWithinLimit 测试限流器在限制内允许请求
func TestCreateRateLimiter_AllowsWithinLimit(t *testing.T) {
	config := RateLimitConfig{
		Limit:   5,
		Window:  time.Second,
		KeyFunc: IPKeyFunc,
	}
	limiter := CreateRateLimiter(config)

	router := gin.New()
	router.Use(limiter)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 发送 5 个请求，都应该成功
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}
}

// TestCreateRateLimiter_BlocksOverLimit 测试限流器在超限时阻止请求
func TestCreateRateLimiter_BlocksOverLimit(t *testing.T) {
	config := RateLimitConfig{
		Limit:    3,
		Window:   time.Second,
		KeyFunc:  IPKeyFunc,
		ErrorMsg: "Test rate limit",
	}
	limiter := CreateRateLimiter(config)

	router := gin.New()
	router.Use(limiter)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 发送 3 个请求，都应该成功
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// 第 4 个请求应该被阻止
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 4: got status %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

// TestCreateRateLimiter_SkipOnError 测试 SkipOnError 选项
func TestCreateRateLimiter_SkipOnError(t *testing.T) {
	// 创建一个总是返回失败的 KeyFunc
	failingKeyFunc := func(c *gin.Context) (string, bool) {
		return "", false
	}

	tests := []struct {
		name        string
		skipOnError bool
		wantStatus  int
	}{
		{
			name:        "skip on error",
			skipOnError: true,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "use default key on error",
			skipOnError: false,
			wantStatus:  http.StatusOK, // 第一个请求应该成功
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RateLimitConfig{
				Limit:       1,
				Window:      time.Second,
				KeyFunc:     failingKeyFunc,
				SkipOnError: tt.skipOnError,
			}
			limiter := CreateRateLimiter(config)

			router := gin.New()
			router.Use(limiter)
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestRateLimiter_DifferentKeys 测试不同 key 独立限流
func TestRateLimiter_DifferentKeys(t *testing.T) {
	config := RateLimitConfig{
		Limit:  2,
		Window: time.Second,
		KeyFunc: func(c *gin.Context) (string, bool) {
			return c.GetHeader("X-User-ID"), true
		},
	}
	limiter := CreateRateLimiter(config)

	router := gin.New()
	router.Use(limiter)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 用户 A 发送 2 个请求
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user-a")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User A request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// 用户 B 发送 2 个请求（应该成功，因为是不同的 key）
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user-b")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User B request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// 用户 A 第 3 个请求应该被阻止
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user-a")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("User A request 3: got status %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

// TestV2Limiter_Equivalence 测试 V2 限流器与原版行为等价
func TestV2Limiter_Equivalence(t *testing.T) {
	// 测试 GlobalRateLimiterV2
	t.Run("GlobalRateLimiterV2", func(t *testing.T) {
		limiter := GlobalRateLimiterV2()
		router := gin.New()
		router.Use(limiter)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// 应该允许请求
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
	})

	// 测试 EndpointRateLimiterV2
	t.Run("EndpointRateLimiterV2", func(t *testing.T) {
		limiter := EndpointRateLimiterV2(2, time.Second)
		router := gin.New()
		router.Use(limiter)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// 前 2 个请求应该成功
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
			}
		}

		// 第 3 个请求应该被阻止
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Request 3: got status %d, want %d", w.Code, http.StatusTooManyRequests)
		}
	})

	// 测试 UserRateLimiterV2
	t.Run("UserRateLimiterV2", func(t *testing.T) {
		limiter := UserRateLimiterV2(2, time.Second)
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", 123)
			c.Next()
		})
		router.Use(limiter)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// 前 2 个请求应该成功
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
			}
		}

		// 第 3 个请求应该被阻止
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Request 3: got status %d, want %d", w.Code, http.StatusTooManyRequests)
		}
	})
}

// TestUserRateLimiterV2_SkipsUnauthenticated 测试用户限流器跳过未认证用户
func TestUserRateLimiterV2_SkipsUnauthenticated(t *testing.T) {
	limiter := UserRateLimiterV2(1, time.Second)
	router := gin.New()
	// 不设置 user_id
	router.Use(limiter)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 多个请求都应该成功（因为未认证用户跳过限流）
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}
}
