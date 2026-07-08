"""
Redis client initialization and management.
Used for caching and deduplication.
"""

from datetime import timedelta
from typing import Optional

import redis.asyncio as aioredis
import structlog

from app.config.settings import Settings

logger = structlog.get_logger(__name__)

                              
_redis_client: Optional["RedisClient"] = None


class RedisClient:
    """Wrapper for Redis async client."""
    
    def __init__(self, client: aioredis.Redis) -> None:
        self._client = client

    async def ping(self) -> None:
        """Ping Redis to check connectivity."""
        await self._client.ping()
        logger.debug("Redis ping successful")

    async def get(self, key: str) -> str | None:
        """Get value from Redis."""
        value = await self._client.get(key)
        if value is None:
            return None
        if isinstance(value, bytes):
            return value.decode("utf-8")
        return str(value)

    async def set(self, key: str, value: str, ttl: timedelta | int = 3600) -> None:
        """Set value in Redis with TTL."""
        if isinstance(ttl, int):
            ttl = timedelta(seconds=ttl)
        await self._client.set(key, value, ex=int(ttl.total_seconds()))

    async def delete(self, key: str) -> None:
        """Delete value from Redis."""
        await self._client.delete(key)

    async def close(self) -> None:
        """Close Redis connection."""
        await self._client.aclose()
        logger.info("Redis client closed")


async def create_redis_client(url: str) -> RedisClient:
    """Create Redis client instance."""
    logger.info(f"Creating Redis client: {url}")
    client = aioredis.from_url(url, decode_responses=False)
    return RedisClient(client)


async def init_redis(settings: Settings) -> RedisClient:
    """Initialize global Redis client."""
    global _redis_client
    
    logger.info(f"Initializing Redis: {settings.redis_url}")
    _redis_client = await create_redis_client(settings.redis_url)
    
                     
    await _redis_client.ping()
    logger.info("Redis initialized successfully")
    
    return _redis_client


async def close_redis() -> None:
    """Close global Redis client."""
    global _redis_client
    if _redis_client:
        await _redis_client.close()
        _redis_client = None


async def get_redis() -> RedisClient:
    """Get global Redis client."""
    global _redis_client
    if not _redis_client:
        raise RuntimeError("Redis client not initialized")
    return _redis_client

