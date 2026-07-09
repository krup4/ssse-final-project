"""Unit tests for the weather collection service."""

from datetime import UTC, datetime

import pytest

from app.models.schemas import StationRead, WeatherKafkaMessage, WeatherMeasurement
from app.services.weather import WeatherService


class FakeWeatherClient:
    def __init__(self, measurement=None, exc=None):
        self._measurement = measurement or WeatherMeasurement(
            station_id=1,
            timestamp=datetime(2026, 6, 1, 15, 0, tzinfo=UTC),
            temperature=18.6,
        )
        self._exc = exc
        self.calls = 0

    async def get_current_measurement(self, lat, lon):
        self.calls += 1
        if self._exc:
            raise self._exc
        return self._measurement


class FakeRedis:
    def __init__(self, previous=None):
        self._previous = previous
        self.stored = None

    async def get(self, key):
        return self._previous

    async def set(self, key, value, ttl):
        self.stored = (key, value, ttl)


class FakePublisher:
    def __init__(self):
        self.published = []

    async def publish(self, message: WeatherKafkaMessage):
        self.published.append(message)



@pytest.mark.asyncio
async def test_collect_for_station_saves_and_publishes():
    station = StationRead(
        id=1, name="Moscow", latitude=55.75, longitude=37.61, region="RU", active=True
    )
    client = FakeWeatherClient()
    redis = FakeRedis(previous=None)
    publisher = FakePublisher()
    service = WeatherService(client, redis, publisher)
    await service.collect_for_station(station)

    assert len(publisher.published) == 1
    assert redis.stored is not None
    message = publisher.published[0]
    assert message.station_id == station.id
    assert message.observed_at == client._measurement.timestamp


def test_weather_kafka_message_uses_core_api_json_contract():
    message = WeatherKafkaMessage(
        station_id=7,
        observed_at=datetime(2026, 6, 1, 15, 0, tzinfo=UTC),
        temperature=18.6,
        wind_speed=4.2,
    )

    payload = message.model_dump(mode="json", by_alias=True, exclude_none=True)

    assert payload == {
        "stationId": 7,
        "observedAt": "2026-06-01T15:00:00Z",
        "temperature": 18.6,
        "windSpeed": 4.2,
        "source": "yandex-weather",
    }


@pytest.mark.asyncio
async def test_collect_skips_duplicate_timestamp():
    station = StationRead(
        id=1, name="Moscow", latitude=55.75, longitude=37.61, region="RU", active=True
    )
    measurement = WeatherMeasurement(
        station_id=1,
        timestamp=datetime(2026, 6, 1, 14, 0, tzinfo=UTC),
        temperature=10.0,
    )
    client = FakeWeatherClient(measurement=measurement)
                                                                       
    redis = FakeRedis(previous="2026-06-01T15:00:00Z")
    publisher = FakePublisher()
    service = WeatherService(client, redis, publisher)
    await service.collect_for_station(station)

    assert len(publisher.published) == 0


@pytest.mark.asyncio
async def test_collect_propagates_weather_error():
    station = StationRead(
        id=1, name="Moscow", latitude=55.75, longitude=37.61, region="RU", active=True
    )
    client = FakeWeatherClient(exc=RuntimeError("boom"))
    redis = FakeRedis(previous=None)
    publisher = FakePublisher()
    service = WeatherService(client, redis, publisher)
    with pytest.raises(RuntimeError):
        await service.collect_for_station(station)

    assert len(publisher.published) == 0
