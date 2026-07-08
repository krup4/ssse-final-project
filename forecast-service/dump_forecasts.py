from datetime import datetime
import asyncpg
import os
from forecast_fetcher import get_forecast
from db.utils import get_stations, update_forecasts
import logging

DB_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/weather")


async def update_and_dump_forecast_data():
    logging.info("start updating forecasts")

    pool = await asyncpg.create_pool(DB_URL, min_size=1, max_size=5)
    stations = await get_stations(pool)

    data_to_update = []

    for station in stations:
        lat, lon = station['lat'], station['lon']

        forecast = await get_forecast(lat, lon)
        if not forecast:
            continue

        for vals in forecast:
            dt = datetime.fromisoformat(vals['time']).replace(tzinfo=None)
            temp = vals["temperature"]
            humidity = vals["humidity"]
            press = vals["pressure"]
            wind_speed = vals["windSpeed"]

            for idx, val in enumerate([temp, wind_speed, humidity, press], start=1):
                if val:
                    data_to_update.append([dt, idx, val, station['id']])

    await update_forecasts(data_to_update, pool)
