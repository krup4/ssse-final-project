package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"weather-accuracy/core-api/internal/domain"
)

func (r *BackfillRepository) Create(ctx context.Context, request domain.BackfillRequest) (domain.BackfillJob, error) {
	now := time.Now().UTC()
	model := BackfillJobModel{
		ID:                 "bf-" + now.Format("20060102-150405") + "-" + uuid.NewString()[:8],
		Status:             string(domain.BackfillQueued),
		DateFrom:           request.DateFrom,
		DateTo:             request.DateTo,
		StationID:          request.StationID,
		Metric:             string(request.Metric),
		CalculationVersion: request.CalculationVersion,
		CreatedAt:          now,
	}
	err := r.db.WithContext(ctx).Create(&model).Error
	return toBackfill(model), err
}

func (r *BackfillRepository) Get(ctx context.Context, id string) (domain.BackfillJob, error) {
	var model BackfillJobModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	return toBackfill(model), mapError(err)
}

func (r *BackfillRepository) HasActiveConflict(ctx context.Context, request domain.BackfillRequest) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&BackfillJobModel{}).
		Where("status in ?", []string{"queued", "running"}).
		Where("date_from <= ? and date_to >= ?", request.DateTo, request.DateFrom).
		Where("(station_id = ? or station_id = '' or ? = '')", request.StationID, request.StationID).
		Count(&count).Error
	return count > 0, err
}
