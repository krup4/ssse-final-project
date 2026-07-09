"""
Health check endpoint.
"""

import logging
from datetime import datetime

from fastapi import APIRouter, HTTPException, status

from app.config.settings import settings
from app.core.database import get_db_pool, ping_pool
from app.core.kafka import check_kafka
from app.core.redis_client import get_redis
from app.models.schemas import HealthResponse

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/health", tags=["health"])


@router.get("", response_model=HealthResponse)
async def health_check():
    """Check health of all services."""
    errors = []
    
                    
    try:
        pool = await get_db_pool()
        await ping_pool(pool)
        db_status = "ok"
    except Exception as e:
        logger.error(f"Database health check failed: {e}")
        db_status = "error"
        errors.append(str(e))
    
                 
    try:
        redis = await get_redis()
        await redis.ping()
        redis_status = "ok"
    except Exception as e:
        logger.error(f"Redis health check failed: {e}")
        redis_status = "error"
        errors.append(str(e))
    
                 
    try:
        await check_kafka(settings.kafka_broker_list)
        kafka_status = "ok"
    except Exception as e:
        logger.error(f"Kafka health check failed: {e}")
        kafka_status = "error"
        errors.append(str(e))
    
                    
    status_value = "ok" if not errors else "error"
    
    if status_value == "error":
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=f"Service unhealthy: {', '.join(errors)}"
        )
    
    return HealthResponse(
        status=status_value,
        database=db_status,
        redis=redis_status,
        kafka=kafka_status,
        timestamp=datetime.utcnow(),
    )
