package httpgin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

type updateUserRoleRequest struct {
	Role domain.UserRole `json:"role" binding:"required"`
}

func (h *Handler) users(c *gin.Context) {
	rows, err := h.svc.ListUsers(c.Request.Context(), currentClaims(c).Role)
	respond(c, http.StatusOK, toUserResponses(rows), err)
}

func (h *Handler) updateUserRole(c *gin.Context) {
	var req updateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil || !validRole(req.Role) {
		validation(c, "invalid user role")
		return
	}
	row, err := h.svc.UpdateUserRole(c.Request.Context(), c.Param("id"), req.Role, currentClaims(c).Role)
	respond(c, http.StatusOK, toUserResponse(row), err)
}
