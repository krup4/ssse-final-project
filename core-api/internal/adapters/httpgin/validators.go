package httpgin

import "weather-accuracy/core-api/internal/domain"

func validRole(role domain.UserRole) bool {
	switch role {
	case domain.RoleAdmin, domain.RoleAnalyst, domain.RoleOperator, domain.RoleViewer:
		return true
	default:
		return false
	}
}

func validMetric(metric domain.Metric) bool {
	switch metric {
	case domain.MetricTemperature, domain.MetricWindSpeed, domain.MetricHumidity, domain.MetricPressure, domain.MetricPrecipitation:
		return true
	default:
		return false
	}
}

func validParameter(parameter domain.WeatherParameter) bool {
	switch parameter {
	case domain.ParameterTemperatureMin, domain.ParameterTemperatureMax, domain.ParameterPrecipitation, domain.ParameterWindSpeed, domain.ParameterWindGust, domain.ParameterHumidity, domain.ParameterPressure:
		return true
	default:
		return false
	}
}

func validBucket(bucket domain.TimeBucket) bool {
	switch bucket {
	case domain.Bucket1H, domain.Bucket3H, domain.Bucket6H, domain.Bucket1D:
		return true
	default:
		return false
	}
}

func validStationStatus(status domain.StationStatus) bool {
	switch status {
	case domain.StationOnline, domain.StationDegraded, domain.StationOffline:
		return true
	default:
		return false
	}
}

func validAlertStatus(status domain.AlertStatus) bool {
	switch status {
	case domain.AlertOpen, domain.AlertAcknowledged, domain.AlertResolved:
		return true
	default:
		return false
	}
}

func validAlertSeverity(severity domain.AlertSeverity) bool {
	switch severity {
	case domain.AlertCritical, domain.AlertWarning, domain.AlertInfo:
		return true
	default:
		return false
	}
}
