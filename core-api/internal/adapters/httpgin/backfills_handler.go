package httpgin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

type createBackfillRequest struct {
	DateFrom           string        `json:"dateFrom" binding:"required"`
	DateTo             string        `json:"dateTo" binding:"required"`
	RegionID           string        `json:"regionId"`
	StationID          string        `json:"stationId"`
	Metric             domain.Metric `json:"metric"`
	CalculationVersion string        `json:"calculationVersion" binding:"required"`
}

func (h *Handler) createBackfill(c *gin.Context) {
	var req createBackfillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation(c, "invalid backfill payload")
		return
	}
	dateFrom, dateTo, ok := parseDateRange(c, req.DateFrom, req.DateTo)
	if !ok {
		return
	}
	job, err := h.svc.CreateBackfill(c.Request.Context(), domain.BackfillRequest{
		DateFrom:           dateFrom,
		DateTo:             dateTo,
		RegionID:           req.RegionID,
		StationID:          req.StationID,
		Metric:             req.Metric,
		CalculationVersion: req.CalculationVersion,
	}, currentClaims(c).Role)
	respond(c, http.StatusAccepted, toBackfillJobResponse(job), err)
}

func (h *Handler) getBackfill(c *gin.Context) {
	row, err := h.svc.GetBackfill(c.Request.Context(), c.Param("id"))
	respond(c, http.StatusOK, toBackfillJobResponse(row), err)
}
