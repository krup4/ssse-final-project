import asyncio
from contextlib import asynccontextmanager

import structlog
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware

from app.api.routes import health, metrics, stations
from app.clients.yandex_weather import YandexWeatherClient
from app.config.settings import settings
from app.core.database import close_db, create_tables, init_db
from app.core.kafka import close_kafka, init_kafka
from app.core.logging_config import configure_logging
from app.core.redis_client import close_redis, init_redis
from app.monitoring.metrics import (
    HTTP_REQUEST_DURATION_SECONDS,
    HTTP_REQUESTS_TOTAL,
)
from app.repositories.station import StationRepository
from app.scheduler.scheduler import WeatherScheduler
from app.services.weather import WeatherService

logger = structlog.get_logger(__name__)

                                   
_scheduler_task: asyncio.Task | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan: startup and shutdown."""
    global _scheduler_task

    configure_logging(settings.log_level, settings.app_name)
    logger.info("starting_service", version=settings.app_version)

    try:
        db_pool = await init_db(settings)
        await create_tables(settings)

        redis_client = await init_redis(settings)
        kafka_publisher = await init_kafka(settings)

        weather_client = YandexWeatherClient(
            settings.yandex_weather_url,
            settings.yandex_weather_graphql_url,
            settings.yandex_weather_api_key,
            settings.request_timeout,
            settings.yandex_weather_rate_limit_rps,
        )
        station_repo = StationRepository(db_pool, name_filter=settings.station_name_filter)

        weather_service = WeatherService(
            weather_client=weather_client,
            redis_client=redis_client,
            publisher=kafka_publisher,
        )

        scheduler = WeatherScheduler(
            weather_service,
            station_repo,
            redis_client=redis_client,
            collection_interval=settings.update_interval_seconds,
            scheduler_lock_enabled=settings.scheduler_lock_enabled,
            scheduler_lock_key=settings.scheduler_lock_key,
            scheduler_lock_ttl_seconds=settings.scheduler_lock_ttl_seconds,
        )
        _scheduler_task = asyncio.create_task(scheduler.start())
        logger.info("scheduler_started")
    except Exception as exc:                                              
        logger.error("startup_failed", error=str(exc), exc_info=True)
        raise

    yield

    logger.info("shutting_down")
    try:
        if _scheduler_task is not None:
            _scheduler_task.cancel()
            try:
                await _scheduler_task
            except asyncio.CancelledError:
                logger.info("scheduler_stopped")

        await close_kafka()
        await close_redis()
        await close_db()
    except Exception as exc:                    
        logger.error("shutdown_error", error=str(exc), exc_info=True)
    logger.info("shutdown_complete")


app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    description="Weather data collection microservice (Yandex Weather -> PostgreSQL -> Kafka)",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.middleware("http")
async def prometheus_middleware(request: Request, call_next):
    """Record request count and duration for every inbound HTTP request."""
    method = request.method
    endpoint = request.url.path
    start = asyncio.get_event_loop().time()
    try:
        response = await call_next(request)
    except Exception:
        HTTP_REQUESTS_TOTAL.labels(method, endpoint, 500).inc()
        raise
    duration = asyncio.get_event_loop().time() - start
    HTTP_REQUESTS_TOTAL.labels(method, endpoint, response.status_code).inc()
    HTTP_REQUEST_DURATION_SECONDS.labels(method, endpoint).observe(duration)
    return response


app.include_router(health.router)
app.include_router(metrics.router)
app.include_router(stations.router)


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "name": settings.app_name,
        "version": settings.app_version,
        "status": "running",
    }


if __name__ == "__main__":
    import uvicorn

    configure_logging(settings.log_level, settings.app_name)
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.reload,
        log_level=settings.log_level.lower(),
    )
