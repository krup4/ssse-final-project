package httpgin

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/service"
)

const claimsKey = "claims"

type rateLimitBucket struct {
	window time.Time
	count  int
}

func (h *Handler) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			unauthorized(c)
			return
		}
		claims, err := h.svc.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			unauthorized(c)
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func currentClaims(c *gin.Context) service.Claims {
	value, ok := c.Get(claimsKey)
	if !ok {
		return service.Claims{}
	}
	claims, _ := value.(service.Claims)
	return claims
}

func (h *Handler) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		if h.metricRegistry != nil {
			h.metricRegistry.IncHTTPInFlight()
			defer h.metricRegistry.DecHTTPInFlight()
		}
		c.Next()
		duration := time.Since(start)
		if h.metricRegistry != nil {
			h.metricRegistry.ObserveHTTPRequest(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
		}
		h.log.Info("http request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
		)
	}
}

func (h *Handler) timeout() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.cfg.HTTPRequestTTL <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), h.cfg.HTTPRequestTTL)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			errorJSON(c, http.StatusGatewayTimeout, "INTERNAL_ERROR", "request timeout")
		}
	}
}

func (h *Handler) rateLimit() gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]rateLimitBucket{}
	return func(c *gin.Context) {
		if h.cfg.RateLimitRPS <= 0 {
			c.Next()
			return
		}
		key := c.ClientIP()
		allowed, err := h.allowRequest(c.Request.Context(), key, &mu, buckets)
		if err != nil {
			h.log.Warn("redis rate limiter failed, using local fallback", slog.Any("error", err))
			allowed = h.allowLocal(key, &mu, buckets)
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "rate limit exceeded"}})
			return
		}
		c.Next()
	}
}

func (h *Handler) allowRequest(ctx context.Context, key string, mu *sync.Mutex, buckets map[string]rateLimitBucket) (bool, error) {
	if h.limiter != nil {
		return h.limiter.Allow(ctx, key, h.cfg.RateLimitRPS, h.cfg.Redis.RateLimitTTL)
	}
	return h.allowLocal(key, mu, buckets), nil
}

func (h *Handler) allowLocal(key string, mu *sync.Mutex, buckets map[string]rateLimitBucket) bool {
	now := time.Now().Truncate(time.Second)
	mu.Lock()
	defer mu.Unlock()
	state := buckets[key]
	if state.window != now {
		state = rateLimitBucket{window: now}
	}
	state.count++
	buckets[key] = state
	return state.count <= h.cfg.RateLimitRPS
}
