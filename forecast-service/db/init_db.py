import asyncpg
import json


async def _fill_stations(pool: asyncpg.Pool) -> None:
    with open("./db/stations.json") as file:
        stations = json.load(file)

    async with pool.acquire() as conn:
        rows = await conn.fetch(
            """
            SELECT
                id,
                name,
                lon,
                lat
            FROM stations
            ORDER BY id
            """
        )

        if not rows:
            await conn.executemany(
                """
                INSERT INTO stations (name, lat, lon)
                VALUES ($1, $2, $3)
                """,
                stations
            )


async def _fill_forecast_fields(pool: asyncpg.Pool) -> None:
    async with pool.acquire() as conn:
        await conn.executemany(
            """
            INSERT INTO forecast_fields (name)
            VALUES ($1)
            ON CONFLICT (name) DO NOTHING
            """,
            [
                ("temperature",),
                ("wind_speed",),
                ("humidity",),
                ("pressure",),
            ],
        )


async def _ensure_tables(pool: asyncpg.Pool) -> None:
    async with pool.acquire() as conn:
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS forecast_fields (
                id      SERIAL PRIMARY KEY,
                name    TEXT NOT NULL UNIQUE
            )
        """)

        await conn.execute("""
            CREATE TABLE IF NOT EXISTS stations (
                id      SERIAL PRIMARY KEY,
                name    TEXT NOT NULL,
                lon     DOUBLE PRECISION NOT NULL,
                lat     DOUBLE PRECISION NOT NULL
            )
        """)

        await conn.execute("""
            CREATE TABLE IF NOT EXISTS forecasts (
                id          SERIAL PRIMARY KEY,
                dt          TIMESTAMP NOT NULL,
                field_id    INTEGER NOT NULL REFERENCES forecast_fields(id),
                value       DOUBLE PRECISION NOT NULL,
                station_id  INTEGER NOT NULL REFERENCES stations(id)
            )
        """)

        await conn.execute("""
            CREATE INDEX IF NOT EXISTS idx_forecasts_dt
            ON forecasts(dt)
        """)

        await conn.execute("""
            CREATE INDEX IF NOT EXISTS idx_forecasts_station
            ON forecasts(station_id)
        """)

        await conn.execute("""
            CREATE INDEX IF NOT EXISTS idx_forecasts_field
            ON forecasts(field_id)
        """)

        await conn.execute("""
            CREATE UNIQUE INDEX IF NOT EXISTS idx_forecasts_unique
            ON forecasts(dt, station_id, field_id)
        """)


async def _init_db(pool: asyncpg.Pool) -> None:
    print(1)
    await _ensure_tables(pool)
    print(2)
    await _fill_stations(pool)
    print(3)
    await _fill_forecast_fields(pool)
    print(4)
