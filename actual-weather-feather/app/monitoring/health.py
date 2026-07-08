
import asyncpg

from app.core.database import ping_pool
from app.core.kafka import check_kafka
from app.core.redis_client import RedisClient


class HealthChecker:
    def __init__(
        self,
        pool: asyncpg.Pool,
        redis_client: RedisClient,
        kafka_brokers: list[str],
    ) -> None:
        self._pool = pool
        self._redis = redis_client
        self._kafka_brokers = kafka_brokers

    async def check(self) -> None:
        await ping_pool(self._pool)
        await check_kafka(self._kafka_brokers)
        await self._redis.ping()
