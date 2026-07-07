package scheduler

import (
    "context"
    "log"
    "sync"
    "time"

    "actual-weather-feather/internal/models"
    "actual-weather-feather/internal/monitoring"
    "actual-weather-feather/internal/repository"
    wservice "actual-weather-feather/internal/service/weather"
)

type Scheduler struct {
    svc    *wservice.Service
    repo   repository.StationRepository
    metrics *monitoring.Metrics
}

func NewScheduler(svc *wservice.Service, repo repository.StationRepository, m *monitoring.Metrics) *Scheduler {
    return &Scheduler{svc: svc, repo: repo, metrics: m}
}

func (s *Scheduler) waitUntilNextFullHour(ctx context.Context) {
    now := time.Now()
    next := now.Truncate(time.Hour).Add(time.Hour)
    wait := time.Until(next)
    log.Printf("scheduler: waiting %s until next full hour (%s)", wait, next.UTC().Format(time.RFC3339))
    timer := time.NewTimer(wait)
    select {
    case <-ctx.Done():
        timer.Stop()
        return
    case <-timer.C:
        return
    }
}

func (s *Scheduler) Start(ctx context.Context) {
    for {
        // wait until full hour
        s.waitUntilNextFullHour(ctx)

        select {
        case <-ctx.Done():
            return
        default:
        }

        stations, err := s.repo.GetAll(ctx)
        if err != nil {
            log.Printf("scheduler: failed to load stations: %v", err)
            s.metrics.ErrorsTotal.Inc()
            continue
        }

        var wg sync.WaitGroup
        for _, st := range stations {
            wg.Add(1)
            station := st // capture
            go func(station models.Station) {
                defer wg.Done()
                s.metrics.StationsProcessedTotal.Inc()
                if err := s.svc.CollectForStation(ctx, station); err != nil {
                    log.Printf("scheduler: error collecting for station %d: %v", station.ID, err)
                    s.metrics.ErrorsTotal.Inc()
                }
            }(station)
        }
        wg.Wait()

        // sleep until next full hour loop will wait again
    }
}
