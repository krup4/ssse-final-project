"""Unit tests for the weather scheduler."""

import pytest

from app.models.schemas import StationRead
from app.repositories.station import StationRepository
from app.scheduler.scheduler import WeatherScheduler


class FakeStationRepo(StationRepository):
    def __init__(self, stations):
        self._stations = stations

    async def get_active(self):
        return self._stations


class FakeWeatherService:
    def __init__(self):
        self.collected = []

    async def collect_for_station(self, station):
        self.collected.append(station)


def _station(station_id):
    return StationRead(
        id=station_id,
        name=f"S{station_id}",
        latitude=1.0,
        longitude=2.0,
        region="RU",
        active=True,
    )


@pytest.mark.asyncio
async def test_collect_all_processes_every_active_station():
    stations = [_station(1), _station(2), _station(3)]
    repo = FakeStationRepo(stations)
    service = FakeWeatherService()

    scheduler = WeatherScheduler(service, repo)
    await scheduler._collect_all()

    assert [s.id for s in service.collected] == [1, 2, 3]


@pytest.mark.asyncio
async def test_collect_all_handles_empty_stations():
    repo = FakeStationRepo([])
    service = FakeWeatherService()

    scheduler = WeatherScheduler(service, repo)
    await scheduler._collect_all()

    assert service.collected == []
