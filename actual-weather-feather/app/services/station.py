"""
Station service for CRUD operations on weather stations.
"""

import structlog

from app.models.schemas import StationCreate, StationRead, StationUpdate
from app.repositories.station import NotFoundError, StationRepository

logger = structlog.get_logger(__name__)


class StationService:
    """Service for managing stations."""
    
    def __init__(self, repo: StationRepository) -> None:
        self._repo = repo

    async def list_stations(self) -> list[StationRead]:
        """Get all stations."""
        logger.info("Listing all stations")
        return await self._repo.get_all()

    async def list_active_stations(self) -> list[StationRead]:
        """Get all active stations."""
        logger.info("Listing active stations")
        return await self._repo.get_active()

    async def get_station(self, station_id: int) -> StationRead:
        """Get station by ID."""
        logger.info(f"Getting station {station_id}")
        try:
            return await self._repo.get_by_id(station_id)
        except NotFoundError:
            logger.warning(f"Station {station_id} not found")
            raise

    async def create_station(self, payload: StationCreate) -> StationRead:
        """Create a new station."""
        logger.info(f"Creating station: {payload.name}")
        return await self._repo.create(payload)

    async def update_station(self, station_id: int, payload: StationUpdate) -> StationRead:
        """Update an existing station."""
        logger.info(f"Updating station {station_id}")
        try:
            return await self._repo.update(station_id, payload)
        except NotFoundError:
            logger.warning(f"Station {station_id} not found for update")
            raise

    async def delete_station(self, station_id: int) -> None:
        """Delete a station."""
        logger.info(f"Deleting station {station_id}")
        try:
            await self._repo.delete(station_id)
        except NotFoundError:
            logger.warning(f"Station {station_id} not found for delete")
            raise

    async def count_stations(self) -> int:
        """Count total stations."""
        return await self._repo.count()


__all__ = ["NotFoundError", "StationService"]

