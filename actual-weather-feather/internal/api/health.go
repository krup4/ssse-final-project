package api

import (
    "encoding/json"
    "net/http"

    "actual-weather-feather/internal/monitoring"
)

type HealthDependencies struct {
    Checker *monitoring.Checker
}

func NewHealthHandler(checker *monitoring.Checker) *HealthDependencies {
    return &HealthDependencies{Checker: checker}
}

func (h *HealthDependencies) Handle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    status := struct {
        Status string `json:"status"`
        Error  string `json:"error,omitempty"`
    }{Status: "ok"}

    if err := h.Checker.Check(ctx); err != nil {
        status.Status = "error"
        status.Error = err.Error()
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(status)
}
