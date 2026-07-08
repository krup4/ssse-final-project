import asyncio
from pathlib import Path

import asyncpg
import structlog

logger = structlog.get_logger(__name__)


async def apply_migrations(database_url: str, migrations_dir: Path | None = None) -> None:
    if migrations_dir is None:
        migrations_dir = Path(__file__).resolve().parents[2] / "migrations"

    sql_files = sorted(migrations_dir.glob("*.sql"))
    if not sql_files:
        logger.warning("no_migration_files", dir=str(migrations_dir))
        return

    conn = await asyncpg.connect(dsn=database_url)
    try:
        for path in sql_files:
            sql = path.read_text(encoding="utf-8")
            logger.info("applying_migration", file=path.name)
            await conn.execute(sql)
    finally:
        await conn.close()


def main() -> None:
    import os

    database_url = os.environ.get(
        "DATABASE_URL",
        "postgres://postgres:password@localhost:5432/weather?sslmode=disable",
    )
    asyncio.run(apply_migrations(database_url))


if __name__ == "__main__":
    main()
