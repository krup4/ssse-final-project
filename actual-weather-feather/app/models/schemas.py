"""
Pydantic models and database schemas for Station and Weather Measurement.
"""

from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field


class StationBase(BaseModel):
    """Base Station model."""
    name: str = Field(..., min_length=1, max_length=255, description="Station name")
    latitude: float = Field(..., ge=-90, le=90, description="Latitude coordinate")
    longitude: float = Field(..., ge=-180, le=180, description="Longitude coordinate")
    region: str = Field(default="", max_length=255, description="Region name")
    active: bool = Field(default=True, description="Station is active for data collection")


class StationCreate(StationBase):
    """Station creation request."""
    pass


class StationUpdate(BaseModel):
    """Station update request - all fields optional."""
    name: str | None = Field(None, min_length=1, max_length=255)
    latitude: float | None = Field(None, ge=-90, le=90)
    longitude: float | None = Field(None, ge=-180, le=180)
    region: str | None = Field(None, max_length=255)
    active: bool | None = None


class StationRead(StationBase):
    """Station response model."""
    model_config = ConfigDict(from_attributes=True)

    id: int
    created_at: datetime | None = None
    updated_at: datetime | None = None


class WeatherMeasurement(BaseModel):
    """Weather measurement model."""
    model_config = ConfigDict(from_attributes=True)

    id: int | None = None
    station_id: int | None = None
    timestamp: datetime
    temperature: float | None = Field(None, ge=-100, le=100, description="Temperature in °C")
    humidity: float | None = Field(None, ge=0, le=100, description="Relative humidity in %")
    pressure: float | None = Field(None, ge=300, le=1100, description="Pressure in hPa")
    wind_speed: float | None = Field(None, ge=0, le=200, description="Wind speed in m/s")
    created_at: datetime | None = None


class WeatherKafkaMessage(BaseModel):
    """Kafka message for weather data."""
    model_config = ConfigDict(populate_by_name=True)

    id: str | None = None
    station_id: int = Field(..., alias="stationId")
    observed_at: datetime = Field(..., alias="observedAt")
    temperature: float | None = None
    humidity: float | None = None
    pressure: float | None = None
    wind_speed: float | None = Field(None, alias="windSpeed")
    source: str = "yandex-weather"
    trace_id: str | None = Field(None, alias="traceId")


class HealthResponse(BaseModel):
    """Health check response."""
    status: str = Field(..., description="Overall status: ok or error")
    database: str | None = None
    redis: str | None = None
    kafka: str | None = None
    timestamp: datetime | None = None


class ErrorResponse(BaseModel):
    """Error response."""
    detail: str
    status_code: int


class YandexWeatherFact(BaseModel):
    """Yandex weather fact block."""
    temp: float | None = None
    feels_like: float | None = None
    wind_speed: float | None = None
    wind_dir: str | None = None
    pressure_mm: float | None = None
    pressure_pa: float | None = None
    humidity: float | None = None
    prec_mm: float | None = None
    obs_time: int | None = None


class YandexWeatherResponse(BaseModel):
    """Yandex Weather API response."""
    fact: YandexWeatherFact | None = None
