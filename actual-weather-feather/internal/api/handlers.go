package api

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"

    "actual-weather-feather/internal/models"
    "actual-weather-feather/internal/repository"
)

type AppDeps struct {
    StationRepo repository.StationRepository
}

func RegisterRoutes(r *chi.Mux, repo repository.StationRepository, healthHandler *HealthDependencies) {
    deps := &AppDeps{StationRepo: repo}

    r.Get("/health", healthHandler.Handle)
    r.Route("/stations", func(r chi.Router) {
        r.Get("/", deps.listStations)
        r.Post("/", deps.createStation)
        r.Get("/{id}", deps.getStation)
        r.Put("/{id}", deps.updateStation)
        r.Delete("/{id}", deps.deleteStation)
    })
}

func (a *AppDeps) health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (a *AppDeps) listStations(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    stations, err := a.StationRepo.GetAll(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(stations)
}

func (a *AppDeps) getStation(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    st, err := a.StationRepo.GetByID(r.Context(), id)
    if err != nil {
        if err == repository.ErrNotFound {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(st)
}

func (a *AppDeps) createStation(w http.ResponseWriter, r *http.Request) {
    var s models.Station
    if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if s.Name == "" {
        http.Error(w, "name required", http.StatusBadRequest)
        return
    }
    if err := a.StationRepo.Create(r.Context(), &s); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Location", r.URL.Path+"/"+strconv.FormatInt(s.ID, 10))
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(s)
}

func (a *AppDeps) updateStation(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    var s models.Station
    if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    s.ID = id
    if err := a.StationRepo.Update(r.Context(), &s); err != nil {
        if err == repository.ErrNotFound {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (a *AppDeps) deleteStation(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.ParseInt(idStr, 10, 64)
    if err := a.StationRepo.Delete(r.Context(), id); err != nil {
        if err == repository.ErrNotFound {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
