package httpgin

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) readyz(c *gin.Context) {
	if h.ready != nil {
		timeout := h.readyTTL
		if timeout <= 0 {
			timeout = 2 * time.Second
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		if err := h.ready(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) metrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4")
	if h.metricRegistry == nil {
		c.String(http.StatusOK, "# HELP core_api_up Core API process availability.\n# TYPE core_api_up gauge\ncore_api_up 1\n")
		return
	}
	c.Status(http.StatusOK)
	h.metricRegistry.WritePrometheus(c.Writer)
}
