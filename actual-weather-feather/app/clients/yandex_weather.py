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
    def __init__(
        self,
        base_url: str,
        graphql_url: str,
        api_key: str,
        timeout: float,
        rate_limit_rps: float = 1.0,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._graphql_url = graphql_url.rstrip("/")
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
        try:
            data = await self.get_current_weather(lat, lon)
            return parse_measurement(data)
        except httpx.HTTPStatusError as exc:
            if exc.response.status_code != 403:
                raise
            logger.warning("yandex_weather_rest_forbidden_using_graphql_fallback")
            return await self.get_current_measurement_graphql(lat, lon)

    async def get_current_measurement_graphql(self, lat: float, lon: float) -> WeatherMeasurement:
        query = """query CurrentWeather($lat: Float!, $lon: Float!) {
  weatherByPoint(request: { lat: $lat, lon: $lon }) {
    forecast {
      days(limit: 1) {
        hours {
          time
          temperature
          humidity
          pressure
          windSpeed
        }
      }
    }
  }
}"""

        await self._rate_limiter.wait()
        WEATHER_REQUESTS_TOTAL.inc()
        start = time.perf_counter()
        try:
            response = await self._client.post(
                self._graphql_url,
                json={"query": query, "variables": {"lat": lat, "lon": lon}},
            )
            response.raise_for_status()
            payload = response.json()
        finally:
            REQUEST_DURATION_SECONDS.observe(time.perf_counter() - start)

        if payload.get("errors"):
            raise RuntimeError(f"Yandex Weather GraphQL errors: {payload['errors']}")

        hours = payload["data"]["weatherByPoint"]["forecast"]["days"][0]["hours"]
        now = datetime.now(timezone.utc)
        parsed = []
        for hour in hours:
            ts = datetime.fromisoformat(hour["time"]).astimezone(timezone.utc)
            parsed.append((ts, hour))
        past = [(ts, hour) for ts, hour in parsed if ts <= now]
        ts, hour = max(past or parsed, key=lambda item: item[0])
        return WeatherMeasurement(
            timestamp=ts,
            temperature=hour.get("temperature"),
            humidity=hour.get("humidity"),
            pressure=hour.get("pressure"),
            wind_speed=hour.get("windSpeed"),
            created_at=now,
        )


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
