package httpgin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"weather-accuracy/core-api/internal/domain"
)

func respond(c *gin.Context, status int, payload any, err error) {
	if err == nil {
		c.JSON(status, payload)
		return
	}
	switch {
	case errors.Is(err, domain.ErrValidation):
		validation(c, "validation error")
	case errors.Is(err, domain.ErrInvalidCredentials):
		unauthorized(c)
	case errors.Is(err, domain.ErrForbidden):
		errorJSON(c, http.StatusForbidden, "FORBIDDEN", "permission denied")
	case errors.Is(err, domain.ErrNotFound):
		errorJSON(c, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case errors.Is(err, domain.ErrConflict):
		errorJSON(c, http.StatusConflict, "CONFLICT", "request conflicts with current state")
	default:
		errorJSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}
}

func validation(c *gin.Context, message string) {
	errorJSON(c, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func unauthorized(c *gin.Context) {
	errorJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authentication token")
}

func errorJSON(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
