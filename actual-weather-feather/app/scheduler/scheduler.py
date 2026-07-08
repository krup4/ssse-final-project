"""Scheduler that collects weather exactly on hourly multiples."""

import asyncio
from datetime import datetime, timedelta, timezone

import structlog

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
        collection_interval: int = 3600,
        max_concurrency: int = MAX_CONCURRENCY,
    ) -> None:
        self.weather_service = weather_service
        self.station_repo = station_repo
        self._semaphore = asyncio.Semaphore(max_concurrency)
        self._running = False
        self._stop_event = asyncio.Event()
        self._collection_interval = max(3600, collection_interval)
        self._interval_hours = max(1, self._collection_interval // 3600)

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
