package httpgin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

type updateAlertRequest struct {
	Status domain.AlertStatus `json:"status" binding:"required"`
}

func (h *Handler) alerts(c *gin.Context) {
	status := domain.AlertStatus(c.Query("status"))
	severity := domain.AlertSeverity(c.Query("severity"))
	if status != "" && !validAlertStatus(status) {
		validation(c, "invalid alert status")
		return
	}
	if severity != "" && !validAlertSeverity(severity) {
		validation(c, "invalid alert severity")
		return
	}
	rows, err := h.svc.ListAlerts(c.Request.Context(), status, severity, c.Query("source"))
	respond(c, http.StatusOK, toAlertResponses(rows), err)
}

func (h *Handler) updateAlert(c *gin.Context) {
	var req updateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil || !validAlertStatus(req.Status) {
		validation(c, "invalid alert status")
		return
	}
	row, err := h.svc.UpdateAlert(c.Request.Context(), c.Param("id"), req.Status, currentClaims(c).Role)
	respond(c, http.StatusOK, toAlertResponse(row), err)
}
