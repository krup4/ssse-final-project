package httpgin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

func (h *Handler) overview(c *gin.Context) {
	filter, ok := h.analyticsFilter(c, false, false)
	if !ok {
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.Overview(c.Request.Context(), filter)
		return toOverviewMetricsResponse(rows), err
	})
}

func (h *Handler) worstErrors(c *gin.Context) {
	filter, ok := h.analyticsFilter(c, false, false)
	if !ok {
		return
	}
	limit, err := strconv.Atoi(defaultString(c.Query("limit"), "50"))
	if err != nil {
		validation(c, "limit must be an integer")
		return
	}
	sort := defaultString(c.Query("sort"), "absoluteError_desc")
	if sort != "absoluteError_desc" && sort != "absoluteError_asc" {
		validation(c, "invalid sort")
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.WorstErrors(c.Request.Context(), filter, limit, sort)
		return toForecastErrorRowResponses(rows), err
	})
}

func (h *Handler) parameterErrors(c *gin.Context) {
	filter, ok := h.parameterFilter(c)
	if !ok {
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.ParameterErrors(c.Request.Context(), filter)
		return toParameterErrorRowResponses(rows), err
	})
}

func (h *Handler) parameterTrend(c *gin.Context) {
	filter, ok := h.parameterFilter(c)
	if !ok {
		return
	}
	filter.Bucket = domain.TimeBucket(defaultString(c.Query("bucket"), string(domain.Bucket1H)))
	if !validBucket(filter.Bucket) {
		validation(c, "invalid bucket")
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.ParameterTrend(c.Request.Context(), filter)
		return toParameterErrorTrendPointResponses(rows), err
	})
}

func (h *Handler) stationSeries(c *gin.Context) {
	filter, ok := h.analyticsFilter(c, true, true)
	if !ok {
		return
	}
	bucket := domain.TimeBucket(defaultString(c.Query("bucket"), string(domain.Bucket1H)))
	if !validBucket(bucket) {
		validation(c, "invalid bucket")
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.StationSeries(c.Request.Context(), filter, bucket)
		return toStationSeriesPointResponses(rows), err
	})
}

func (h *Handler) history(c *gin.Context) {
	filter, ok := h.analyticsFilter(c, false, false)
	if !ok {
		return
	}
	h.cachedJSON(c, h.cfg.Redis.CacheTTL, func() (any, error) {
		rows, err := h.svc.History(c.Request.Context(), filter)
		return toHistoricalMetricResponses(rows), err
	})
}
