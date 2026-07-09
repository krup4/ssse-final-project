"""Scheduler that collects weather exactly on hourly multiples."""

import asyncio
from datetime import datetime, timedelta, timezone
from uuid import uuid4

import structlog

from app.core.redis_client import RedisClient
from app.monitoring.metrics import ERRORS_TOTAL, STATIONS_PROCESSED_TOTAL
from app.repositories.station import StationRepository
from app.services.weather import WeatherService

logger = structlog.get_logger(__name__)

                                                    
MAX_CONCURRENCY = 5


class WeatherScheduler:
    """Scheduler for periodic weather data collection."""

    def __init__(self,
        weather_service: WeatherService,
        station_repo: StationRepository,
        redis_client: RedisClient | None = None,
        collection_interval: int = 3600,
        max_concurrency: int = MAX_CONCURRENCY,
        scheduler_lock_enabled: bool = True,
        scheduler_lock_key: str = "actual-weather-feather:scheduler:leader",
        scheduler_lock_ttl_seconds: int = 3900,
    ) -> None:
        self.weather_service = weather_service
        self.station_repo = station_repo
        self.redis_client = redis_client
        self._semaphore = asyncio.Semaphore(max_concurrency)
        self._running = False
        self._stop_event = asyncio.Event()
        self._collection_interval = max(3600, collection_interval)
        self._interval_hours = max(1, self._collection_interval // 3600)
        self._lock_enabled = scheduler_lock_enabled
        self._lock_key = scheduler_lock_key
        self._lock_ttl_seconds = max(self._collection_interval + 300, scheduler_lock_ttl_seconds)
        self._lock_value = str(uuid4())

    def _seconds_until_next_multiple(self, now: datetime | None = None) -> int:
        now = (now or datetime.now(timezone.utc)).astimezone(timezone.utc)
        base = now.replace(minute=0, second=0, microsecond=0)
        remainder = now.hour % self._interval_hours
        next_hours = self._interval_hours - remainder if remainder != 0 else self._interval_hours
        target = base + timedelta(hours=next_hours)
        return max(1, int((target - now).total_seconds()))

    async def start(self) -> None:
        self._running = True
        self._stop_event.clear()
        logger.info("weather_scheduler_started")

        while self._running:
            try:
                sleep_seconds = self._seconds_until_next_multiple()
                next_run = datetime.now(timezone.utc) + timedelta(seconds=sleep_seconds)
                logger.info(
                    "scheduler_waiting",
                    interval_seconds=self._collection_interval,
                    next_run=next_run.isoformat(),
                )
                try:
                    await asyncio.wait_for(self._stop_event.wait(), timeout=sleep_seconds)
                    break
                except asyncio.TimeoutError:
                    pass

                if not self._running:
                    break

                if await self._should_run_cycle():
                    await self._collect_all()

            except asyncio.CancelledError:
                logger.info("scheduler_cancelled")
                break
            except Exception as exc:
                logger.error("scheduler_error", error=str(exc), exc_info=True)
                ERRORS_TOTAL.inc()
                await asyncio.sleep(60)

    async def stop(self) -> None:
        self._running = False
        self._stop_event.set()
        logger.info("scheduler_stop_requested")

    async def _should_run_cycle(self) -> bool:
        if not self._lock_enabled:
            return True
        if self.redis_client is None:
            logger.warning("scheduler_lock_disabled_no_redis")
            return True

        acquired = await self.redis_client.acquire_lock(
            self._lock_key,
            self._lock_value,
            self._lock_ttl_seconds,
        )
        if not acquired:
            logger.info("scheduler_cycle_skipped_not_leader", lock_key=self._lock_key)
            return False

        logger.info("scheduler_leader_lock_acquired", lock_key=self._lock_key)
        return True

    async def _collect_all(self) -> None:
        try:
            stations = await self.station_repo.get_active()
        except Exception as exc:
            logger.error("load_stations_failed", error=str(exc), exc_info=True)
            ERRORS_TOTAL.inc()
            return

        if not stations:
            logger.warning("no_active_stations")
            return

        logger.info("collecting_weather", station_count=len(stations))

        async def collect_one(station):
            async with self._semaphore:
                STATIONS_PROCESSED_TOTAL.inc()
                try:
                    await self.weather_service.collect_for_station(station)
                except Exception as exc:
                    logger.error(
                        "station_collection_failed",
                        station_id=station.id,
                        error=str(exc),
                    )

        await asyncio.gather(
            *(collect_one(station) for station in stations),
            return_exceptions=True,
        )
        logger.info("weather_collection_completed", station_count=len(stations))
