import aiohttp
import os
import logging
from typing import Optional

from rate_limiter import AsyncRateLimiter


API_KEY = os.getenv("FORECAST_API_KEY", "API-KEY")
URL = os.getenv("URL", "https://api.weather.yandex.ru/graphql/query")
RATE_LIMIT_RPS = float(os.getenv("FORECAST_RATE_LIMIT_RPS", "1"))
RATE_LIMITER = AsyncRateLimiter(RATE_LIMIT_RPS)

HEADERS = {
    'X-Yandex-Weather-Key': API_KEY,
    "Content-Type": "application/json"
}


BASIC_QUERY = """{{
  weatherByPoint(request: {{ lat: {}, lon: {} }}) {{
    forecast {{
      days(limit: 5) {{
        hours {{
          time
          temperature
          humidity
          pressure
          windSpeed
        }}
      }}
    }}
  }}
}}"""


async def get_forecast(lat: float, lon: float) -> Optional[dict]:
    query = BASIC_QUERY.format(lat, lon)

    try:
        await RATE_LIMITER.wait()
        async with aiohttp.ClientSession() as session:
            async with session.post(
                URL,
                json={"query": query},
                headers=HEADERS,
            ) as response:
                response.raise_for_status()
                data = await response.json()

    except Exception as e:
        logging.error(f"Fail attempt of fetching forecast data: {e}")
        return None

    try:
        return data["data"]["weatherByPoint"]["forecast"]["days"][1]["hours"]
    except (KeyError, IndexError) as e:
        logging.error(f"Fail attempt of extracting forecast data: {e}")
        return None
