package postgres

import (
	"errors"

	"gorm.io/gorm"

	"weather-accuracy/core-api/internal/domain"
)

func setActualMetric(model *ActualWeatherReadingModel, metric string, value float64) {
	switch metric {
	case "temperature":
		model.TemperatureMax = &value
	case "wind_speed":
		model.WindSpeed = &value
	case "humidity":
		model.Humidity = &value
	case "pressure":
		model.Pressure = &value
	case "precipitation":
		model.PrecipitationTotal = &value
	}
}

func setParameter(model *ActualWeatherReadingModel, parameter domain.WeatherParameter, value float64) {
	switch parameter {
	case domain.ParameterTemperatureMin:
		model.TemperatureMin = &value
	case domain.ParameterTemperatureMax:
		model.TemperatureMax = &value
	case domain.ParameterPrecipitation:
		model.PrecipitationTotal = &value
	case domain.ParameterWindSpeed:
		model.WindSpeed = &value
	case domain.ParameterWindGust:
		model.WindGust = &value
	case domain.ParameterHumidity:
		model.Humidity = &value
	case domain.ParameterPressure:
		model.Pressure = &value
	}
}

func toUser(model UserModel) domain.User {
	return domain.User{
		ID:           model.ID,
		Name:         model.Name,
		Email:        model.Email,
		Role:         domain.UserRole(model.Role),
		LastSeen:     model.LastSeen,
		PasswordHash: model.PasswordHash,
	}
}

func toAlert(model AlertModel) domain.Alert {
	return domain.Alert{
		ID:        model.ID,
		Title:     model.Title,
		Source:    model.Source,
		Severity:  domain.AlertSeverity(model.Severity),
		Status:    domain.AlertStatus(model.Status),
		StartedAt: model.StartedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toBackfill(model BackfillJobModel) domain.BackfillJob {
	return domain.BackfillJob{
		ID:                 model.ID,
		Status:             domain.BackfillStatus(model.Status),
		CalculationVersion: model.CalculationVersion,
		ProgressPct:        model.ProgressPct,
		ProcessedRows:      model.ProcessedRows,
		FailedRows:         model.FailedRows,
		CreatedAt:          model.CreatedAt,
		StartedAt:          model.StartedAt,
		FinishedAt:         model.FinishedAt,
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return err
}
