"""Unit tests for the Yandex Weather HTTP client."""

import httpx
import pytest

from app.clients.yandex_weather import YandexWeatherClient
from app.models.schemas import WeatherMeasurement


PAYLOAD = {
    "fact": {
        "temp": 18.6,
        "humidity": 71,
        "pressure_mm": 760,
        "pressure_pa": 1013,
        "wind_speed": 5.2,
        "wind_dir": "s",
        "prec_mm": 0.0,
        "obs_time": 1710000000,
    }
}


@pytest.mark.asyncio
async def test_get_current_measurement_parses_payload(monkeypatch):
    transport = httpx.MockTransport(lambda request: httpx.Response(200, json=PAYLOAD))
    client = YandexWeatherClient("https://api.weather.yandex.ru/v2/informers", "key", 10.0)
    client._client = httpx.AsyncClient(transport=transport)

    measurement = await client.get_current_measurement(55.75, 37.61)

    assert isinstance(measurement, WeatherMeasurement)
    assert measurement.temperature == 18.6
    assert measurement.humidity == 71
    assert measurement.pressure == pytest.approx(1013 / 100)
    assert measurement.wind_speed == 5.2
