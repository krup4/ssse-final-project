package postgres

import "weather-accuracy/core-api/internal/domain"

func baseErrorSQL() string {
	return `
		select concat(f.id, '-', f.metric) as id,
		       f.station_id,
		       s.name as station_name,
		       r.name as region_name,
		       f.metric,
		       f.value as forecast_value,
		       case f.metric
		         when 'temperature' then a.temperature_max
		         when 'wind_speed' then a.wind_speed
		         when 'humidity' then a.humidity
		         when 'pressure' then a.pressure
		         when 'precipitation' then a.precipitation_total
		       end as actual_value,
		       abs(f.value - case f.metric
		         when 'temperature' then a.temperature_max
		         when 'wind_speed' then a.wind_speed
		         when 'humidity' then a.humidity
		         when 'pressure' then a.pressure
		         when 'precipitation' then a.precipitation_total
		       end) as absolute_error,
		       case when abs(f.value) < 0.000001 then 0 else abs((case f.metric
		         when 'temperature' then a.temperature_max
		         when 'wind_speed' then a.wind_speed
		         when 'humidity' then a.humidity
		         when 'pressure' then a.pressure
		         when 'precipitation' then a.precipitation_total
		       end - f.value) / f.value * 100) end as error_pct,
		       a.observed_at
		from forecast_reading_models f
		join actual_weather_reading_models a on a.station_id = f.station_id and a.observed_at = f.target_at
		join station_models s on s.id = f.station_id
		join region_models r on r.id = s.region_id
		where f.target_at >= ? and f.target_at <= ?
		  and (? = '' or ? = 'all' or s.region_id = ?)
		  and (? = '' or ? = 'all' or f.station_id = ?)
		  and (? = '' or f.metric = ?)
		  and case f.metric
		         when 'temperature' then a.temperature_max
		         when 'wind_speed' then a.wind_speed
		         when 'humidity' then a.humidity
		         when 'pressure' then a.pressure
		         when 'precipitation' then a.precipitation_total
		      end is not null`
}

func baseParameterErrorSQL(parameter string) string {
	filter := ""
	if parameter != "" && parameter != "all" {
		filter = " and p.parameter = ?"
	}
	return `
		select concat(f.id, '-', p.parameter) as id,
		       p.parameter,
		       f.station_id,
		       s.name as station_name,
		       r.name as region_name,
		       f.value as forecast_value,
		       p.actual_value,
		       abs(f.value - p.actual_value) as absolute_error,
		       case when abs(f.value) < 0.000001 then 0 else abs((p.actual_value - f.value) / f.value * 100) end as error_pct,
		       a.observed_at
		from forecast_reading_models f
		join actual_weather_reading_models a on a.station_id = f.station_id and a.observed_at = f.target_at
		join station_models s on s.id = f.station_id
		join region_models r on r.id = s.region_id
		join lateral (
			values
			  ('temperature_min', a.temperature_min),
			  ('temperature_max', a.temperature_max),
			  ('precipitation_total', a.precipitation_total),
			  ('wind_speed', a.wind_speed),
			  ('wind_gust', a.wind_gust),
			  ('humidity', a.humidity),
			  ('pressure', a.pressure)
		) as p(parameter, actual_value) on p.actual_value is not null
		where f.target_at >= ? and f.target_at <= ?
		  and (? = '' or ? = 'all' or s.region_id = ?)
		  and (? = '' or ? = 'all' or f.station_id = ?)
		  and f.metric = case
		     when p.parameter in ('temperature_min', 'temperature_max') then 'temperature'
		     when p.parameter = 'precipitation_total' then 'precipitation'
		     else p.parameter
		  end` + filter
}

func sqlArgs(filter domain.AnalyticsFilter) []any {
	return []any{
		filter.DateFrom,
		filter.DateTo,
		filter.RegionID, filter.RegionID, filter.RegionID,
		filter.StationID, filter.StationID, filter.StationID,
		string(filter.Metric), string(filter.Metric),
	}
}

func parameterArgs(filter domain.ParameterFilter) []any {
	args := []any{
		filter.DateFrom,
		filter.DateTo,
		filter.RegionID, filter.RegionID, filter.RegionID,
		filter.StationID, filter.StationID, filter.StationID,
	}
	if filter.Parameter != "" && filter.Parameter != "all" {
		args = append(args, filter.Parameter)
	}
	return args
}

func bucketExpression(bucket domain.TimeBucket, column string) string {
	switch bucket {
	case domain.Bucket3H:
		return "date_trunc('hour', " + column + ") - make_interval(hours => (extract(hour from " + column + ")::int % 3))"
	case domain.Bucket6H:
		return "date_trunc('hour', " + column + ") - make_interval(hours => (extract(hour from " + column + ")::int % 6))"
	case domain.Bucket1D:
		return "date_trunc('day', " + column + ")"
	default:
		return "date_trunc('hour', " + column + ")"
	}
}
