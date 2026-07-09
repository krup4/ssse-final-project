import asyncpg
import json


async def _fill_stations(pool: asyncpg.Pool) -> None:
    with open("./db/stations.json") as file:
        stations = json.load(file)

    async with pool.acquire() as conn:
        await conn.executemany(
            """
            INSERT INTO stations (name, lat, lon)
            SELECT $1::text, $2::double precision, $3::double precision
            WHERE NOT EXISTS (
                SELECT 1
                FROM stations
                WHERE name::text = $1::text
                  AND lat = $2::double precision
                  AND lon = $3::double precision
            )
            """,
            stations
        )


async def fill_initial_stations(pool: asyncpg.Pool) -> None:
    await _fill_stations(pool)
