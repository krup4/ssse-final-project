"""
Station repository for managing weather stations.
"""

import asyncpg
import structlog

from app.models.schemas import StationCreate, StationRead, StationUpdate

logger = structlog.get_logger(__name__)


class NotFoundError(Exception):
    """Station not found exception."""
    pass


class StationRepository:
    """Repository for Station CRUD operations."""
    
    def __init__(self, pool: asyncpg.Pool, name_filter: str | None = None) -> None:
        self._pool = pool
        self._name_filter = name_filter

    async def get_active(self) -> list[StationRead]:
        """Get all active stations."""
        query = """
                SELECT id,
                       name,
                       COALESCE(latitude, lat) AS latitude,
                       COALESCE(longitude, lon) AS longitude,
                       region,
                       COALESCE(is_active, active, true) AS active,
                       created_at,
                       updated_at
                FROM stations
            """
        if self._name_filter:
            query += " WHERE name = $1"
            query += " ORDER BY id"
            async with self._pool.acquire() as conn:
                rows = await conn.fetch(query, self._name_filter)
        else:
            query += " WHERE COALESCE(is_active, active, true) = true"
            query += " ORDER BY id"
            async with self._pool.acquire() as conn:
                rows = await conn.fetch(query)
        return [StationRead(**dict(row)) for row in rows]

    async def get_all(self) -> list[StationRead]:
        """Get all stations."""
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(
                """
                SELECT id,
                       name,
                       COALESCE(latitude, lat) AS latitude,
                       COALESCE(longitude, lon) AS longitude,
                       region,
                       COALESCE(is_active, active, true) AS active,
                       created_at,
                       updated_at
                FROM stations
                ORDER BY id
                """
            )
        return [StationRead(**dict(row)) for row in rows]

    async def get_by_id(self, station_id: int) -> StationRead:
        """Get station by ID."""
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                SELECT id,
                       name,
                       COALESCE(latitude, lat) AS latitude,
                       COALESCE(longitude, lon) AS longitude,
                       region,
                       COALESCE(is_active, active, true) AS active,
                       created_at,
                       updated_at
                FROM stations
                WHERE id = $1
                """,
                station_id,
            )
        if row is None:
            raise NotFoundError(f"Station {station_id} not found")
        return StationRead(**dict(row))

    async def create(self, station: StationCreate) -> StationRead:
        """Create a new station."""
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                INSERT INTO stations (name, lat, lon, latitude, longitude, region, is_active, active)
                VALUES ($1, $2, $3, $2, $3, $4, $5, $5)
                RETURNING id,
                          name,
                          COALESCE(latitude, lat) AS latitude,
                          COALESCE(longitude, lon) AS longitude,
                          region,
                          COALESCE(is_active, active, true) AS active,
                          created_at,
                          updated_at
                """,
                station.name,
                station.latitude,
                station.longitude,
                station.region,
                station.active,
            )
        logger.info(f"Created station: {row['id']}")
        return StationRead(**dict(row))

    async def update(self, station_id: int, station: StationUpdate) -> StationRead:
        """Update station by ID."""
                                    
        existing = await self.get_by_id(station_id)
        
                                                        
        name = station.name if station.name is not None else existing.name
        latitude = station.latitude if station.latitude is not None else existing.latitude
        longitude = station.longitude if station.longitude is not None else existing.longitude
        region = station.region if station.region is not None else existing.region
        active = station.active if station.active is not None else existing.active
        
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                UPDATE stations
                SET name = $1,
                    lat = $2,
                    lon = $3,
                    latitude = $2,
                    longitude = $3,
                    region = $4,
                    is_active = $5,
                    active = $5,
                    updated_at = CURRENT_TIMESTAMP
                WHERE id = $6
                RETURNING id,
                          name,
                          COALESCE(latitude, lat) AS latitude,
                          COALESCE(longitude, lon) AS longitude,
                          region,
                          COALESCE(is_active, active, true) AS active,
                          created_at,
                          updated_at
                """,
                name,
                latitude,
                longitude,
                region,
                active,
                station_id,
            )
        
        if row is None:
            raise NotFoundError(f"Station {station_id} not found")
        
        logger.info(f"Updated station: {station_id}")
        return StationRead(**dict(row))

    async def delete(self, station_id: int) -> None:
        """Delete station by ID."""
        async with self._pool.acquire() as conn:
            result = await conn.execute("DELETE FROM stations WHERE id = $1", station_id)
        
        if result == "DELETE 0":
            raise NotFoundError(f"Station {station_id} not found")
        
        logger.info(f"Deleted station: {station_id}")

    async def count(self) -> int:
        """Count total stations."""
        async with self._pool.acquire() as conn:
            return await conn.fetchval("SELECT COUNT(*) FROM stations")
