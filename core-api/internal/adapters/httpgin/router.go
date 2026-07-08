package httpgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(deps Dependencies) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	h := NewHandler(deps)
	router := gin.New()
	router.Use(gin.Recovery(), h.requestLogger(), h.timeout(), h.rateLimit())

	h.registerSystemRoutes(router)
	h.registerAPIRoutes(router.Group("/api/v1"))

	return router
}

func (h *Handler) registerSystemRoutes(router *gin.Engine) {
	router.GET("/healthz", h.healthz)
	router.GET("/readyz", h.readyz)
	router.GET("/metrics", h.metrics)
}

func (h *Handler) registerAPIRoutes(api *gin.RouterGroup) {
	api.POST("/auth/login", h.login)

	protected := api.Group("")
	protected.Use(h.auth())
	protected.GET("/auth/me", h.me)

	protected.GET("/regions", h.regions)
	protected.GET("/forecast-fields", h.forecastFields)
	protected.GET("/metrics/catalog", h.metricDefinitions)
	protected.GET("/stations", h.stations)
	protected.POST("/stations", h.createStation)
	protected.PATCH("/stations/:id", h.updateStation)

	protected.GET("/metrics/overview", h.overview)
	protected.GET("/analytics/worst-errors", h.worstErrors)
	protected.GET("/analytics/parameter-errors", h.parameterErrors)
	protected.GET("/analytics/parameter-error-trend", h.parameterTrend)
	protected.GET("/analytics/station-series", h.stationSeries)
	protected.GET("/analytics/history", h.history)

	protected.GET("/alerts", h.alerts)
	protected.PATCH("/alerts/:id", h.updateAlert)

	protected.GET("/users", h.users)
	protected.PATCH("/users/:id/role", h.updateUserRole)

	protected.POST("/backfills", h.createBackfill)
	protected.GET("/backfills/:id", h.getBackfill)
}
