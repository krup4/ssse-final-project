package httpgin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) cachedJSON(c *gin.Context, ttl time.Duration, load func() (any, error)) {
	if h.cache == nil || ttl <= 0 {
		payload, err := load()
		respond(c, http.StatusOK, payload, err)
		return
	}

	key := cacheKey(c)
	if cached, err := h.cache.Get(c.Request.Context(), key); err == nil && len(cached) > 0 {
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
		return
	}

	payload, err := load()
	if err != nil {
		respond(c, http.StatusOK, nil, err)
		return
	}
	body, err := json.Marshal(payload)
	if err == nil {
		_ = h.cache.Set(c.Request.Context(), key, body, ttl)
	}
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

func cacheKey(c *gin.Context) string {
	sum := sha256.Sum256([]byte(c.Request.Method + ":" + c.FullPath() + "?" + c.Request.URL.RawQuery))
	return hex.EncodeToString(sum[:])
}
