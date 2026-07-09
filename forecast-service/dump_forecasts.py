from datetime import datetime
import asyncpg
import os
from forecast_fetcher import get_forecast
from db.utils import get_forecast_field_ids, get_stations, update_forecasts
import logging

DB_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/weather")


async def update_and_dump_forecast_data():
    logging.info("start updating forecasts")

    pool = await asyncpg.create_pool(DB_URL, min_size=1, max_size=5)
    try:
        stations = await get_stations(pool)
        field_ids = await get_forecast_field_ids(pool)
        required_fields = ["temperature", "wind_speed", "humidity", "pressure"]
        missing_fields = [field for field in required_fields if field not in field_ids]
        if missing_fields:
            logging.error("missing forecast fields: %s", ", ".join(missing_fields))
            return

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

                values_by_field = {
                    "temperature": temp,
                    "wind_speed": wind_speed,
                    "humidity": humidity,
                    "pressure": press,
                }

                for field_name, val in values_by_field.items():
                    if val is not None:
                        data_to_update.append([dt, field_ids[field_name], val, station['id']])

        await update_forecasts(data_to_update, pool)
    finally:
        await pool.close()
