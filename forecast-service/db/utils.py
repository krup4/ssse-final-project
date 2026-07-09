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


async def get_forecast_field_ids(pool: asyncpg.Pool) -> dict[str, int]:
    async with pool.acquire() as conn:
        fields = await conn.fetch(
            """
            SELECT id, name
            FROM forecast_fields
            WHERE name = ANY($1::text[])
            """,
            ["temperature", "wind_speed", "humidity", "pressure"],
        )

        return {field["name"]: field["id"] for field in fields}


async def update_forecasts(data: list[list[Any]], pool: asyncpg.Pool) -> Optional[dict]:
    if not data:
        return None

    async with pool.acquire() as conn:
        await conn.executemany(
            """
            INSERT INTO forecasts (date, field_id, value, interval, station_id, is_archived)
            VALUES ($1, $2, $3, '24h', $4, false)
            ON CONFLICT (date, station_id, field_id)
            DO UPDATE SET
                value = EXCLUDED.value,
                interval = EXCLUDED.interval,
                is_archived = EXCLUDED.is_archived
            """,
            data
        )


async def get_forecasts(pool: asyncpg.Pool) -> list:
    async with pool.acquire() as conn:
        return await conn.fetch("SELECT * FROM forecasts")
