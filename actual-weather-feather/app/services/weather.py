from datetime import timedelta

import structlog

from app.clients.yandex_weather import YandexWeatherClient
from app.core.kafka import KafkaPublisher
from app.core.redis_client import RedisClient
from app.models.schemas import StationRead, WeatherKafkaMessage
from app.monitoring.metrics import ERRORS_TOTAL

logger = structlog.get_logger(__name__)


class WeatherService:
    def __init__(
        self,
        weather_client: YandexWeatherClient,
        redis_client: RedisClient,
        publisher: KafkaPublisher,
    ) -> None:
        self._weather_client = weather_client
        self._redis = redis_client
        self._publisher = publisher

    async def collect_for_station(self, station: StationRead) -> None:
        try:
            measurement = await self._weather_client.get_current_measurement(
                station.latitude,
                station.longitude,
            )
        except Exception as exc:
            ERRORS_TOTAL.inc()
            logger.error(
                "weather_fetch_failed",
                station_id=station.id,
                error=str(exc),
            )
            raise

        if measurement.timestamp is None:
            ERRORS_TOTAL.inc()
            raise ValueError("empty timestamp in measurement")

        cache_key = f"station:{station.id}:last_timestamp"
        previous = await self._redis.get(cache_key)
        if previous is not None:
            from datetime import datetime

            prev_ts = datetime.fromisoformat(previous.replace("Z", "+00:00"))
            if measurement.timestamp <= prev_ts:
                logger.info(
                    "measurement_skipped_duplicate",
                    station_id=station.id,
                    timestamp=measurement.timestamp.isoformat(),
                )
                return

        await self._redis.set(
            cache_key,
            measurement.timestamp.isoformat().replace("+00:00", "Z"),
            timedelta(hours=2),
        )

        kafka_message = WeatherKafkaMessage(
            station_id=station.id,
            timestamp=measurement.timestamp,
            temperature=measurement.temperature,
            humidity=measurement.humidity,
            pressure=measurement.pressure,
            wind_speed=measurement.wind_speed,
        )

        try:
            await self._publisher.publish(kafka_message)
        except Exception as exc:
            ERRORS_TOTAL.inc()
            logger.error("kafka_publish_failed", station_id=station.id, error=str(exc))
            raise

        logger.info(
            "measurement_published",
            station_id=station.id,
            timestamp=measurement.timestamp.isoformat(),
        )
