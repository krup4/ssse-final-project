import asyncpg
from db.init_db import _init_db
import os
import asyncio
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from dump_forecasts import update_and_dump_forecast_data

DB_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/weather")


async def init_db() -> None:
    pool = await asyncpg.create_pool(DB_URL, min_size=1, max_size=5)
    await _init_db(pool)


async def main():
    await init_db()

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
