"""
Prometheus metrics endpoint.
"""

from fastapi import APIRouter, Response
from prometheus_client import CONTENT_TYPE_LATEST, generate_latest

router = APIRouter(prefix="/metrics", tags=["metrics"])


@router.get("")
async def get_metrics() -> Response:
    """Expose Prometheus metrics in text exposition format."""
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)
