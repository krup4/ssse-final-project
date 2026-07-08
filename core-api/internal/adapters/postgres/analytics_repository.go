package postgres

import (
	"context"
	"math"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"

	"weather-accuracy/core-api/internal/domain"
)

type forecastActualRow struct {
	ForecastID    int
	MetricID      int
	ID            string
	StationID     string
	StationName   string
	RegionName    string
	Parameter     string
	Metric        string
	ForecastValue float64
	ActualValue   float64
	ObservedAt    time.Time
}

type archiveCalculation struct {
	ForecastID  int
	StationID   int
	MetricID    int
	ObservedAt  time.Time
	MetricValue float64
}

type parameterActualRow struct {
	ForecastID    int
	MetricID      int
	ID            string
	Parameter     string
	StationID     string
	StationName   string
	RegionName    string
	ForecastValue float64
	ActualValue   float64
	ObservedAt    time.Time
}

type errorAccumulator struct {
	sumAbs    float64
	sumSquare float64
	samples   int
}

func (a *errorAccumulator) add(absoluteError float64) {
	a.sumAbs += absoluteError
	a.sumSquare += absoluteError * absoluteError
	a.samples++
}

func (a errorAccumulator) mae() float64 {
	if a.samples == 0 {
		return 0
	}
	return a.sumAbs / float64(a.samples)
}

func (a errorAccumulator) rmse() float64 {
	if a.samples == 0 {
		return 0
	}
	return math.Sqrt(a.sumSquare / float64(a.samples))
}

func (r *AnalyticsRepository) Overview(ctx context.Context, filter domain.AnalyticsFilter) (domain.OverviewMetrics, error) {
	var overview domain.OverviewMetrics
	var active, inactive int64
	_ = r.db.WithContext(ctx).Model(&StationModel{}).Where("is_active = true").Count(&active).Error
	_ = r.db.WithContext(ctx).Model(&StationModel{}).Where("is_active = false").Count(&inactive).Error
	overview.ActiveStations = int(active)
	overview.DegradedStations = 0
	overview.OfflineStations = int(inactive)
	overview.RequestRate = 0
	overview.P95LatencyMs = 0
	overview.KafkaLag = 0

	currentRows, err := r.forecastErrorRows(ctx, filter, true)
	if err != nil {
		return overview, err
	}
	sortForecastErrors(currentRows, "absoluteError_desc")
	if len(currentRows) > 0 {
		overview.WorstErrorToday = currentRows[0].AbsoluteError
	}

	trendRows, err := r.forecastErrorRows(ctx, filter, false)
	if err != nil {
		return overview, err
	}

	byDate := map[string]*errorAccumulator{}
	for _, row := range trendRows {
		date := row.ObservedAt.Format("2006-01-02")
		acc := byDate[date]
		if acc == nil {
			acc = &errorAccumulator{}
			byDate[date] = acc
		}
		acc.add(row.AbsoluteError)
	}
	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	overview.ErrorTrend = make([]domain.ErrorTrendPoint, 0, len(dates))
	for _, date := range dates {
		acc := byDate[date]
		overview.ErrorTrend = append(overview.ErrorTrend, domain.ErrorTrendPoint{
			Date: date,
			MAE:  acc.mae(),
			RMSE: acc.rmse(),
		})
	}
	return overview, nil
}

func (r *AnalyticsRepository) WorstErrors(ctx context.Context, filter domain.AnalyticsFilter, limit int, sortOrder string) ([]domain.ForecastErrorRow, error) {
	rows, err := r.forecastErrorRows(ctx, filter, true)
	if err != nil {
		return nil, err
	}
	sortForecastErrors(rows, sortOrder)
	if limit > len(rows) {
		limit = len(rows)
	}
	return rows[:limit], nil
}

func (r *AnalyticsRepository) ParameterErrors(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorRow, error) {
	rows, err := r.parameterErrorRows(ctx, filter, true)
	if err != nil {
		return nil, err
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].AbsoluteError > rows[j].AbsoluteError
	})
	totalError := 0.0
	for _, row := range rows {
		totalError += row.AbsoluteError
	}
	for i := range rows {
		if totalError > 0 {
			rows[i].ContributionPct = round2(rows[i].AbsoluteError / totalError * 100)
		}
		rows[i].Samples = 1
	}
	if len(rows) > 500 {
		rows = rows[:500]
	}
	return rows, nil
}

func (r *AnalyticsRepository) ParameterTrend(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorTrendPoint, error) {
	rows, err := r.parameterErrorRows(ctx, filter, false)
	if err != nil {
		return nil, err
	}
	type key struct {
		timestamp time.Time
		parameter domain.WeatherParameter
	}
	type trendAccumulator struct {
		errorAccumulator
		maxAbs float64
	}
	byBucket := map[key]*trendAccumulator{}
	for _, row := range rows {
		k := key{timestamp: bucketStart(row.ObservedAt, filter.Bucket), parameter: row.Parameter}
		acc := byBucket[k]
		if acc == nil {
			acc = &trendAccumulator{}
			byBucket[k] = acc
		}
		if row.AbsoluteError > acc.maxAbs {
			acc.maxAbs = row.AbsoluteError
		}
		acc.add(row.AbsoluteError)
	}
	keys := make([]key, 0, len(byBucket))
	for k := range byBucket {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].timestamp.Equal(keys[j].timestamp) {
			return keys[i].parameter < keys[j].parameter
		}
		return keys[i].timestamp.Before(keys[j].timestamp)
	})
	out := make([]domain.ParameterErrorTrendPoint, 0, len(keys))
	for _, k := range keys {
		acc := byBucket[k]
		out = append(out, domain.ParameterErrorTrendPoint{
			Timestamp:     k.timestamp,
			Parameter:     k.parameter,
			AbsoluteError: acc.maxAbs,
			MAE:           acc.mae(),
		})
	}
	return out, nil
}

func (r *AnalyticsRepository) StationSeries(ctx context.Context, filter domain.AnalyticsFilter, bucket domain.TimeBucket) ([]domain.StationSeriesPoint, error) {
	rows, err := r.forecastErrorRows(ctx, filter, false)
	if err != nil {
		return nil, err
	}
	type seriesAccumulator struct {
		forecastSum float64
		actualSum   float64
		errorSum    float64
		samples     int
	}
	byBucket := map[time.Time]*seriesAccumulator{}
	for _, row := range rows {
		ts := bucketStart(row.ObservedAt, bucket)
		acc := byBucket[ts]
		if acc == nil {
			acc = &seriesAccumulator{}
			byBucket[ts] = acc
		}
		acc.forecastSum += row.ForecastValue
		acc.actualSum += row.ActualValue
		acc.errorSum += row.AbsoluteError
		acc.samples++
	}
	timestamps := make([]time.Time, 0, len(byBucket))
	for ts := range byBucket {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})
	out := make([]domain.StationSeriesPoint, 0, len(timestamps))
	for _, ts := range timestamps {
		acc := byBucket[ts]
		samples := float64(acc.samples)
		out = append(out, domain.StationSeriesPoint{
			Timestamp:     ts,
			Forecast:      acc.forecastSum / samples,
			Actual:        acc.actualSum / samples,
			AbsoluteError: acc.errorSum / samples,
		})
	}
	return out, nil
}

func (r *AnalyticsRepository) History(ctx context.Context, filter domain.AnalyticsFilter) ([]domain.HistoricalMetric, error) {
	rows, err := r.forecastErrorRows(ctx, filter, false)
	if err != nil {
		return nil, err
	}
	type key struct {
		date        string
		regionName  string
		stationName string
	}
	byGroup := map[key]*errorAccumulator{}
	for _, row := range rows {
		k := key{
			date:        row.ObservedAt.Format("2006-01-02"),
			regionName:  row.RegionName,
			stationName: row.StationName,
		}
		acc := byGroup[k]
		if acc == nil {
			acc = &errorAccumulator{}
			byGroup[k] = acc
		}
		acc.add(row.AbsoluteError)
	}
	keys := make([]key, 0, len(byGroup))
	for k := range byGroup {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].date == keys[j].date {
			return byGroup[keys[i]].mae() > byGroup[keys[j]].mae()
		}
		return keys[i].date > keys[j].date
	})
	out := make([]domain.HistoricalMetric, 0, len(keys))
	for _, k := range keys {
		acc := byGroup[k]
		out = append(out, domain.HistoricalMetric{
			Date:            k.date,
			RegionName:      k.regionName,
			StationName:     k.stationName,
			MAE:             acc.mae(),
			RMSE:            acc.rmse(),
			Samples:         acc.samples,
			BackfillVersion: "calc-v1",
		})
	}
	return out, nil
}

func (r *AnalyticsRepository) forecastErrorRows(ctx context.Context, filter domain.AnalyticsFilter, currentOnly bool) ([]domain.ForecastErrorRow, error) {
	var raw []forecastActualRow
	err := r.db.WithContext(ctx).Raw(baseForecastActualSQL(currentOnly), sqlArgs(filter)...).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	usedFallback := false
	if currentOnly && len(raw) == 0 {
		err = r.db.WithContext(ctx).Raw(baseForecastActualSQL(false), sqlArgs(filter)...).Scan(&raw).Error
		if err != nil {
			return nil, err
		}
		usedFallback = true
	}
	rows := make([]domain.ForecastErrorRow, 0, len(raw))
	calculations := make([]archiveCalculation, 0, len(raw))
	for _, row := range raw {
		metricValue := row.ActualValue
		stationID, _ := strconv.Atoi(row.StationID)
		calculations = append(calculations, archiveCalculation{
			ForecastID:  row.ForecastID,
			StationID:   stationID,
			MetricID:    row.MetricID,
			ObservedAt:  row.ObservedAt,
			MetricValue: metricValue,
		})
		rows = append(rows, domain.ForecastErrorRow{
			ID:            row.ID,
			StationID:     row.StationID,
			StationName:   row.StationName,
			RegionName:    row.RegionName,
			Parameter:     row.Parameter,
			Metric:        domain.Metric(row.Metric),
			ForecastValue: row.ForecastValue,
			ActualValue:   row.ActualValue,
			AbsoluteError: metricValue,
			ErrorPct:      metricPercent(row.ForecastValue, metricValue),
			ObservedAt:    row.ObservedAt,
		})
	}
	if currentOnly && !usedFallback && len(calculations) > 0 {
		if err := r.archiveCalculations(ctx, calculations); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (r *AnalyticsRepository) parameterErrorRows(ctx context.Context, filter domain.ParameterFilter, currentOnly bool) ([]domain.ParameterErrorRow, error) {
	var raw []parameterActualRow
	err := r.db.WithContext(ctx).Raw(baseParameterActualSQL(filter.Parameter, currentOnly), parameterArgs(filter)...).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	usedFallback := false
	if currentOnly && len(raw) == 0 {
		err = r.db.WithContext(ctx).Raw(baseParameterActualSQL(filter.Parameter, false), parameterArgs(filter)...).Scan(&raw).Error
		if err != nil {
			return nil, err
		}
		usedFallback = true
	}
	rows := make([]domain.ParameterErrorRow, 0, len(raw))
	calculations := make([]archiveCalculation, 0, len(raw))
	for _, row := range raw {
		metricValue := row.ActualValue
		stationID, _ := strconv.Atoi(row.StationID)
		calculations = append(calculations, archiveCalculation{
			ForecastID:  row.ForecastID,
			StationID:   stationID,
			MetricID:    row.MetricID,
			ObservedAt:  row.ObservedAt,
			MetricValue: metricValue,
		})
		rows = append(rows, domain.ParameterErrorRow{
			ID:            row.ID,
			Parameter:     domain.WeatherParameter(row.Parameter),
			StationID:     row.StationID,
			StationName:   row.StationName,
			RegionName:    row.RegionName,
			ForecastValue: row.ForecastValue,
			ActualValue:   row.ActualValue,
			AbsoluteError: metricValue,
			ErrorPct:      metricPercent(row.ForecastValue, metricValue),
			ObservedAt:    row.ObservedAt,
		})
	}
	if currentOnly && !usedFallback && len(calculations) > 0 {
		if err := r.archiveCalculations(ctx, calculations); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (r *AnalyticsRepository) archiveCalculations(ctx context.Context, calculations []archiveCalculation) error {
	forecastIDs := make([]int, 0, len(calculations))
	seenForecasts := map[int]struct{}{}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, calc := range calculations {
			model := ArchiveModel{
				StationID: calc.StationID,
				MetricID:  calc.MetricID,
				Dt:        calc.ObservedAt,
				Value:     calc.MetricValue,
			}
			err := tx.Exec(`
				insert into archive (dt, station_id, metric_id, value)
				values (?, ?, ?, ?)
				on conflict (station_id, metric_id, dt) do update
				set value = excluded.value`,
				model.Dt,
				model.StationID,
				model.MetricID,
				model.Value,
			).Error
			if err != nil {
				return err
			}
			if _, ok := seenForecasts[calc.ForecastID]; !ok {
				seenForecasts[calc.ForecastID] = struct{}{}
				forecastIDs = append(forecastIDs, calc.ForecastID)
			}
		}
		if len(forecastIDs) == 0 {
			return nil
		}
		return tx.Model(&ForecastModel{}).Where("id in ?", forecastIDs).Update("is_archived", true).Error
	})
}

func calculateError(forecastValue, actualValue float64) (absoluteError float64, errorPct float64) {
	absoluteError = math.Abs(forecastValue - actualValue)
	if math.Abs(forecastValue) < 0.000001 {
		return absoluteError, 0
	}
	return absoluteError, math.Abs((actualValue - forecastValue) / forecastValue * 100)
}

func metricPercent(forecastValue, metricValue float64) float64 {
	if math.Abs(forecastValue) < 0.000001 {
		return 0
	}
	return math.Abs(metricValue / forecastValue * 100)
}

func sortForecastErrors(rows []domain.ForecastErrorRow, sortOrder string) {
	sort.Slice(rows, func(i, j int) bool {
		if sortOrder == "absoluteError_asc" {
			return rows[i].AbsoluteError < rows[j].AbsoluteError
		}
		return rows[i].AbsoluteError > rows[j].AbsoluteError
	})
}

func bucketStart(ts time.Time, bucket domain.TimeBucket) time.Time {
	switch bucket {
	case domain.Bucket3H:
		hour := ts.Hour() - ts.Hour()%3
		return time.Date(ts.Year(), ts.Month(), ts.Day(), hour, 0, 0, 0, ts.Location())
	case domain.Bucket6H:
		hour := ts.Hour() - ts.Hour()%6
		return time.Date(ts.Year(), ts.Month(), ts.Day(), hour, 0, 0, 0, ts.Location())
	case domain.Bucket1D:
		return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, ts.Location())
	default:
		return ts.Truncate(time.Hour)
	}
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
