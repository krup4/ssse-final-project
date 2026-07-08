"""
Stations endpoints for CRUD operations.
"""

import logging

from fastapi import APIRouter, Depends, HTTPException, status

from app.core.database import get_db_pool
from app.models.schemas import StationCreate, StationRead, StationUpdate
from app.repositories.station import NotFoundError, StationRepository
from app.services.station import StationService

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/stations", tags=["stations"])


async def get_station_service() -> StationService:
    """Dependency: get station service."""
    pool = await get_db_pool()
    repo = StationRepository(pool)
    return StationService(repo)


@router.get("", response_model=list[StationRead])
async def list_stations(service: StationService = Depends(get_station_service)):
    """Get all stations."""
    try:
        return await service.list_stations()
    except Exception as e:
        logger.error(f"Error listing stations: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to list stations",
        )


@router.get("/{station_id}", response_model=StationRead)
async def get_station(
    station_id: int,
    service: StationService = Depends(get_station_service),
):
    """Get station by ID."""
    try:
        return await service.get_station(station_id)
    except NotFoundError:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Station {station_id} not found",
        )
    except Exception as e:
        logger.error(f"Error getting station {station_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get station",
        )


@router.post("", response_model=StationRead, status_code=status.HTTP_201_CREATED)
async def create_station(
    payload: StationCreate,
    service: StationService = Depends(get_station_service),
):
    """Create a new station."""
    try:
        return await service.create_station(payload)
    except Exception as e:
        logger.error(f"Error creating station: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to create station",
        )


@router.put("/{station_id}", response_model=StationRead)
async def update_station(
    station_id: int,
    payload: StationUpdate,
    service: StationService = Depends(get_station_service),
):
    """Update a station."""
    try:
        return await service.update_station(station_id, payload)
    except NotFoundError:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Station {station_id} not found",
        )
    except Exception as e:
        logger.error(f"Error updating station {station_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to update station",
        )


@router.delete("/{station_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_station(
    station_id: int,
    service: StationService = Depends(get_station_service),
):
    """Delete a station."""
    try:
        await service.delete_station(station_id)
    except NotFoundError:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Station {station_id} not found",
        )
    except Exception as e:
        logger.error(f"Error deleting station {station_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to delete station",
        )
