package httpgin

import (
	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

func (h *Handler) regions(c *gin.Context) {
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.ListRegions(c.Request.Context())
		return toRegionResponses(rows), err
	})
}

func (h *Handler) forecastFields(c *gin.Context) {
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.ListForecastFields(c.Request.Context())
		return toForecastFieldResponses(rows), err
	})
}

func (h *Handler) stations(c *gin.Context) {
	status := domain.StationStatus(c.Query("status"))
	if status != "" && !validStationStatus(status) {
		validation(c, "invalid station status")
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.ListStations(c.Request.Context(), c.Query("regionId"), status)
		return toStationResponses(rows), err
	})
}
