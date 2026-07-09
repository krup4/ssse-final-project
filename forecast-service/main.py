import asyncpg
from db.init_db import fill_initial_stations
import os
import asyncio
import logging
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from dump_forecasts import update_and_dump_forecast_data

DB_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/weather")
DB_CONNECT_RETRIES = int(os.getenv("DB_CONNECT_RETRIES", "30"))
DB_CONNECT_RETRY_DELAY_SECONDS = int(os.getenv("DB_CONNECT_RETRY_DELAY_SECONDS", "5"))


async def prepare_stations() -> None:
    last_error = None

    for attempt in range(1, DB_CONNECT_RETRIES + 1):
        try:
            pool = await asyncpg.create_pool(DB_URL, min_size=1, max_size=5)
            try:
                await fill_initial_stations(pool)
            finally:
                await pool.close()
            return
        except Exception as error:
            last_error = error
            logging.warning("database init attempt %s failed: %s", attempt, error)
            await asyncio.sleep(DB_CONNECT_RETRY_DELAY_SECONDS)

    raise RuntimeError("station preparation failed") from last_error


async def main():
    logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"))
    await prepare_stations()
    await update_and_dump_forecast_data()

    scheduler = AsyncIOScheduler(timezone="Europe/Moscow")

    scheduler.add_job(
        update_and_dump_forecast_data,
        trigger="cron",
        hour=0,
        minute=27,
        second=50,
    )

    scheduler.start()
    await asyncio.Event().wait()


asyncio.run(main())
