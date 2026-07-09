from typing import Optional, Any
import asyncpg


async def get_stations(pool: asyncpg.Pool) -> Optional[dict]:
    async with pool.acquire() as conn:
        stations = await conn.fetch(
            """
                SELECT *
                FROM stations
                """
        )

        return (dict(st) for st in stations)


async def update_forecasts(data: list[list[Any]], pool: asyncpg.Pool) -> Optional[dict]:
    async with pool.acquire() as conn:
        await conn.executemany(
            """
            INSERT INTO forecasts (dt, field_id, value, station_id)
            VALUES ($1, $2, $3, $4)
            """,
            data
        )


async def get_forecasts(pool: asyncpg.Pool) -> list:
    async with pool.acquire() as conn:
        return await conn.fetch("SELECT * FROM forecasts")
