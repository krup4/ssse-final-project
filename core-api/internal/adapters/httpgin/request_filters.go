package httpgin

import (
	"time"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

func (h *Handler) analyticsFilter(c *gin.Context, stationRequired bool, metricRequired bool) (domain.AnalyticsFilter, bool) {
	dateFrom, dateTo, ok := parseDateRange(c, c.Query("dateFrom"), c.Query("dateTo"))
	if !ok {
		return domain.AnalyticsFilter{}, false
	}
	stationID := c.Query("stationId")
	metric := domain.Metric(c.Query("metric"))
	if stationRequired && stationID == "" {
		validation(c, "stationId is required")
		return domain.AnalyticsFilter{}, false
	}
	if metricRequired && metric == "" {
		validation(c, "metric is required")
		return domain.AnalyticsFilter{}, false
	}
	if metric != "" && !validMetric(metric) {
		validation(c, "invalid metric")
		return domain.AnalyticsFilter{}, false
	}
	return domain.AnalyticsFilter{
		DateFrom:  dateFrom,
		DateTo:    dateTo.Add(24*time.Hour - time.Nanosecond),
		RegionID:  c.Query("regionId"),
		StationID: stationID,
		Metric:    metric,
	}, true
}

func (h *Handler) parameterFilter(c *gin.Context) (domain.ParameterFilter, bool) {
	base, ok := h.analyticsFilter(c, false, false)
	if !ok {
		return domain.ParameterFilter{}, false
	}
	parameter := defaultString(c.Query("parameter"), "all")
	if parameter != "all" && !validParameter(domain.WeatherParameter(parameter)) {
		validation(c, "invalid weather parameter")
		return domain.ParameterFilter{}, false
	}
	return domain.ParameterFilter{AnalyticsFilter: base, Parameter: parameter}, true
}

func parseDateRange(c *gin.Context, rawFrom, rawTo string) (time.Time, time.Time, bool) {
	dateFrom, errFrom := time.Parse(time.DateOnly, rawFrom)
	dateTo, errTo := time.Parse(time.DateOnly, rawTo)
	if errFrom != nil || errTo != nil {
		validation(c, "dateFrom and dateTo must use YYYY-MM-DD")
		return time.Time{}, time.Time{}, false
	}
	if dateFrom.After(dateTo) {
		validation(c, "dateFrom must be before or equal to dateTo")
		return time.Time{}, time.Time{}, false
	}
	return dateFrom.UTC(), dateTo.UTC(), true
}
