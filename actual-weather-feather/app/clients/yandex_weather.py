import time
from datetime import datetime, timezone

import httpx
import structlog
from tenacity import (
    RetryCallState,
    retry,
    retry_if_exception_type,
    stop_after_delay,
    wait_exponential,
)

from app.core.rate_limiter import AsyncRateLimiter
from app.models.schemas import WeatherMeasurement, YandexWeatherResponse
from app.monitoring.metrics import (
    REQUEST_DURATION_SECONDS,
    RETRY_TOTAL,
    WEATHER_REQUESTS_TOTAL,
)

logger = structlog.get_logger(__name__)


def _on_retry(retry_state: RetryCallState) -> None:
    RETRY_TOTAL.inc()
    logger.warning(
        "yandex_weather_retry",
        attempt=retry_state.attempt_number,
        error=str(retry_state.outcome.exception()) if retry_state.outcome else None,
    )


class YandexWeatherClient:
    def __init__(self, base_url: str, api_key: str, timeout: float, rate_limit_rps: float = 1.0) -> None:
        self._base_url = base_url.rstrip("/")
        if not api_key:
            raise ValueError("YANDEX_WEATHER_API_KEY is required for YandexWeatherClient")
        self._headers = {"X-Yandex-Weather-Key": api_key}
        self._timeout = timeout
        self._client = httpx.AsyncClient(timeout=timeout, headers=self._headers)
        self._rate_limiter = AsyncRateLimiter(rate_limit_rps)

    async def close(self) -> None:
        await self._client.aclose()

    async def get_current_weather(self, lat: float, lon: float) -> YandexWeatherResponse:
        url = f"{self._base_url}?lat={lat}&lon={lon}&lang=ru_RU"

        @retry(
            retry=retry_if_exception_type((httpx.HTTPError, httpx.HTTPStatusError)),
            wait=wait_exponential(multiplier=1, min=1, max=8),
            stop=stop_after_delay(15),
            before_sleep=_on_retry,
            reraise=True,
        )
        async def _fetch() -> YandexWeatherResponse:
            await self._rate_limiter.wait()
            WEATHER_REQUESTS_TOTAL.inc()
            start = time.perf_counter()
            try:
                response = await self._client.get(url)
                response.raise_for_status()
                return YandexWeatherResponse.model_validate(response.json())
            finally:
                REQUEST_DURATION_SECONDS.observe(time.perf_counter() - start)

        return await _fetch()

    async def get_current_measurement(self, lat: float, lon: float) -> WeatherMeasurement:
        data = await self.get_current_weather(lat, lon)
        return parse_measurement(data)


def parse_measurement(data: YandexWeatherResponse) -> WeatherMeasurement:
    measurement = WeatherMeasurement(timestamp=datetime.now(timezone.utc))

    fact = data.fact
    if fact is not None:
        if fact.obs_time is not None:
            measurement.timestamp = datetime.fromtimestamp(fact.obs_time, tz=timezone.utc)
        measurement.temperature = fact.temp
        measurement.humidity = fact.humidity
        measurement.wind_speed = fact.wind_speed

        if fact.pressure_pa is not None:
            measurement.pressure = fact.pressure_pa / 100.0
        elif fact.pressure_mm is not None:
            measurement.pressure = fact.pressure_mm * 1.33322

    measurement.created_at = datetime.now(timezone.utc)
    return measurement
