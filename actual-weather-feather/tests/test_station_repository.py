"""Unit tests for the station repository."""

from datetime import UTC, datetime

import pytest
from helpers import FakeConn, FakePool

from app.models.schemas import StationCreate, StationRead, StationUpdate
from app.repositories.station import NotFoundError, StationRepository


def _station_row(station_id=1, name="Moscow", active=True):
    return {
        "id": station_id,
        "name": name,
        "latitude": 55.75,
        "longitude": 37.61,
        "region": "RU",
        "active": active,
        "created_at": datetime(2026, 1, 1, tzinfo=UTC),
        "updated_at": datetime(2026, 1, 1, tzinfo=UTC),
    }


@pytest.mark.asyncio
async def test_get_active_returns_only_active():
    conn = FakeConn(fetch=[_station_row(), _station_row(2, "SPB")])
    repo = StationRepository(FakePool(conn))

    result = await repo.get_active()

    assert len(result) == 2
    assert all(isinstance(s, StationRead) for s in result)
    assert result[0].name == "Moscow"


@pytest.mark.asyncio
async def test_get_by_id_found():
    conn = FakeConn(fetchrow=_station_row(7))
    repo = StationRepository(FakePool(conn))

    station = await repo.get_by_id(7)
    assert station.id == 7


@pytest.mark.asyncio
async def test_get_by_id_not_found():
    conn = FakeConn(fetchrow=None)
    repo = StationRepository(FakePool(conn))

    with pytest.raises(NotFoundError):
        await repo.get_by_id(99)


@pytest.mark.asyncio
async def test_create_returns_station():
    conn = FakeConn(fetchrow=_station_row(3))
    repo = StationRepository(FakePool(conn))

    payload = StationCreate(name="Kazan", latitude=55.79, longitude=49.12, region="RU")
    result = await repo.create(payload)

    assert isinstance(result, StationRead)
    assert result.id == 3


@pytest.mark.asyncio
async def test_update_merges_partial_fields():
    existing = _station_row(4, "Old")
    updated = _station_row(4, "New")
    conn = FakeConn(fetchrow=existing)
                                                                   
    conn._fetchrow = updated

    repo = StationRepository(FakePool(conn))
    result = await repo.update(4, StationUpdate(name="New"))

    assert result.name == "New"


@pytest.mark.asyncio
async def test_update_not_found():
    conn = FakeConn(fetchrow=None)
    repo = StationRepository(FakePool(conn))

    with pytest.raises(NotFoundError):
        await repo.update(4, StationUpdate(name="New"))


@pytest.mark.asyncio
async def test_delete_success():
    conn = FakeConn(execute="DELETE 1")
    repo = StationRepository(FakePool(conn))

    await repo.delete(4)                    


@pytest.mark.asyncio
async def test_delete_not_found():
    conn = FakeConn(execute="DELETE 0")
    repo = StationRepository(FakePool(conn))

    with pytest.raises(NotFoundError):
        await repo.delete(4)
