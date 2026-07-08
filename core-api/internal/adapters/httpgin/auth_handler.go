package httpgin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation(c, "invalid login payload")
		return
	}
	token, user, err := h.svc.Login(c.Request.Context(), strings.ToLower(req.Login), req.Password)
	if err != nil {
		unauthorized(c)
		return
	}
	c.JSON(http.StatusOK, loginResponse{Token: token, User: toUserResponse(user)})
}

func (h *Handler) me(c *gin.Context) {
	user, err := h.svc.CurrentUser(c.Request.Context(), currentClaims(c).UserID)
	respond(c, http.StatusOK, toUserResponse(user), err)
}
