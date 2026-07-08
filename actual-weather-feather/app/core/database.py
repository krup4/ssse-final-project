"""
Database connection and initialization.
Uses asyncpg for direct access to PostgreSQL.
"""

from pathlib import Path

import asyncpg
import structlog

from app.config.settings import Settings
from app.core.migrations import apply_migrations

logger = structlog.get_logger(__name__)

                        
db_pool: asyncpg.Pool | None = None


def _to_asyncpg_dsn(database_url: str) -> str:
    """Convert a SQLAlchemy-style asyncpg DSN to a plain asyncpg DSN."""
    return database_url.replace("postgresql+asyncpg", "postgresql")


async def create_tables(settings: Settings, migrations_dir: Path | None = None) -> None:
    """Apply SQL migrations to ensure the schema exists."""
    dsn = _to_asyncpg_dsn(settings.database_url)
    logger.info("Applying database migrations")
    await apply_migrations(dsn, migrations_dir)
    logger.info("Database migrations applied")


async def create_pool(settings: Settings) -> asyncpg.Pool:
    """Create asyncpg connection pool."""
    logger.info(f"Creating asyncpg pool from {settings.database_url}")
    
                           
                                                                  
    dsn = settings.database_url.replace("postgresql+asyncpg", "postgresql")
    
    pool = await asyncpg.create_pool(
        dsn=dsn,
        min_size=1,
        max_size=10,
        command_timeout=10,
    )
    logger.info("Connection pool created")
    return pool


async def init_db(settings: Settings) -> asyncpg.Pool:
    """Initialize database pool."""
    global db_pool
    db_pool = await create_pool(settings)
    return db_pool


async def close_db() -> None:
    """Close database pool."""
    global db_pool
    if db_pool:
        await db_pool.close()
        logger.info("Database pool closed")
        db_pool = None


async def ping_pool(pool: asyncpg.Pool) -> None:
    """Ping database pool to check connectivity."""
    async with pool.acquire() as conn:
        result = await conn.fetchval("SELECT 1")
        if result == 1:
            logger.info("Database ping successful")
        else:
            raise RuntimeError("Database ping failed")


async def get_db_pool() -> asyncpg.Pool:
    """Get global database pool."""
    global db_pool
    if not db_pool:
        raise RuntimeError("Database pool not initialized")
    return db_pool

