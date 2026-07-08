package httpgin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

type stationRequest struct {
	Name     string  `json:"name" binding:"required"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	IsActive bool    `json:"isActive"`
}

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

func (h *Handler) metricDefinitions(c *gin.Context) {
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.ListMetrics(c.Request.Context())
		return toMetricDefinitionResponses(rows), err
	})
}

func (h *Handler) stations(c *gin.Context) {
	status := domain.StationStatus(c.Query("status"))
	if status != "" && !validStationStatus(status) {
		validation(c, "invalid station status")
		return
	}
	rows, err := h.svc.ListStations(c.Request.Context(), c.Query("regionId"), status)
	respond(c, http.StatusOK, toStationResponses(rows), err)
}

func (h *Handler) createStation(c *gin.Context) {
	input, ok := bindStationRequest(c)
	if !ok {
		return
	}
	row, err := h.svc.CreateStation(c.Request.Context(), input, currentClaims(c).Role)
	respond(c, http.StatusCreated, toStationResponses([]domain.Station{row})[0], err)
}

func (h *Handler) updateStation(c *gin.Context) {
	input, ok := bindStationRequest(c)
	if !ok {
		return
	}
	row, err := h.svc.UpdateStation(c.Request.Context(), c.Param("id"), input, currentClaims(c).Role)
	respond(c, http.StatusOK, toStationResponses([]domain.Station{row})[0], err)
}

func bindStationRequest(c *gin.Context) (domain.StationInput, bool) {
	var req stationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation(c, "invalid station payload")
		return domain.StationInput{}, false
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		validation(c, "station name is required")
		return domain.StationInput{}, false
	}
	if req.Lat < -90 || req.Lat > 90 {
		validation(c, "lat must be between -90 and 90")
		return domain.StationInput{}, false
	}
	if req.Lon < -180 || req.Lon > 180 {
		validation(c, "lon must be between -180 and 180")
		return domain.StationInput{}, false
	}
	return domain.StationInput{
		Name:     req.Name,
		Lat:      req.Lat,
		Lon:      req.Lon,
		IsActive: req.IsActive,
	}, true
}
