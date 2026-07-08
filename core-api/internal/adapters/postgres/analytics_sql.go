package postgres

import "weather-accuracy/core-api/internal/domain"

func baseForecastActualSQL(currentOnly bool) string {
	currentFilter := ""
	if currentOnly {
		currentFilter = " and f.is_archived = false"
	}
	return `
		select f.id as forecast_id,
		       m.id as metric_id,
		       concat(f.id::text, '-', m.id::text) as id,
		       f.station_id::text as station_id,
		       s.name as station_name,
		       '' as region_name,
		       ff.name as parameter,
		       m.name as metric,
		       f.value as forecast_value,
		       a.value as actual_value,
		       a.dt as observed_at
		from forecasts f
		join stations s on s.id = f.station_id and s.is_active = true
		join forecast_fields ff on ff.id = f.field_id
		join metrics m on m.forecast_field_id = ff.id
		join archive a on a.station_id = f.station_id and a.metric_id = m.id and a.dt = f.date
		where f.date >= ? and f.date <= ?
		  and (? = '' or ? = 'all' or f.station_id::text = ?)
		  and (? = '' or ? = 'all' or ff.name = ?)
		  and (? = '' or ? = 'all' or m.name = ?)` + currentFilter
}

func baseParameterActualSQL(parameter string, currentOnly bool) string {
	filter := ""
	if parameter != "" && parameter != "all" {
		filter = " and ff.name = ?"
	}
	currentFilter := ""
	if currentOnly {
		currentFilter = " and f.is_archived = false"
	}
	return `
		select f.id as forecast_id,
		       m.id as metric_id,
		       concat(f.id::text, '-', m.id::text) as id,
		       ff.name as parameter,
		       f.station_id::text as station_id,
		       s.name as station_name,
		       '' as region_name,
		       f.value as forecast_value,
		       a.value as actual_value,
		       a.dt as observed_at
		from forecasts f
		join stations s on s.id = f.station_id and s.is_active = true
		join forecast_fields ff on ff.id = f.field_id
		join metrics m on m.forecast_field_id = ff.id
		join archive a on a.station_id = f.station_id and a.metric_id = m.id and a.dt = f.date
		where f.date >= ? and f.date <= ?
		  and (? = '' or ? = 'all' or f.station_id::text = ?)
		  and (? = '' or ? = 'all' or ff.name = ?)
		  and (? = '' or ? = 'all' or m.name = ?)` + currentFilter + filter
}

func sqlArgs(filter domain.AnalyticsFilter) []any {
	return []any{
		filter.DateFrom,
		filter.DateTo,
		filter.StationID, filter.StationID, filter.StationID,
		filter.Field, filter.Field, filter.Field,
		string(filter.Metric), string(filter.Metric), string(filter.Metric),
	}
}

func parameterArgs(filter domain.ParameterFilter) []any {
	args := []any{
		filter.DateFrom,
		filter.DateTo,
		filter.StationID, filter.StationID, filter.StationID,
		filter.Field, filter.Field, filter.Field,
		string(filter.Metric), string(filter.Metric), string(filter.Metric),
	}
	if filter.Parameter != "" && filter.Parameter != "all" {
		args = append(args, filter.Parameter)
	}
	return args
}
