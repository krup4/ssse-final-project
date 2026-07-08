"""Unit tests for the HTTP API (stations + metrics)."""

from datetime import UTC, datetime

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.api.routes import metrics, stations
from app.models.schemas import StationCreate, StationRead, StationUpdate
from app.repositories.station import NotFoundError

STATION_JSON = {
    "name": "Moscow",
    "latitude": 55.75,
    "longitude": 37.61,
    "region": "RU",
    "active": True,
}


def _station_read(station_id=1):
    return StationRead(
        id=station_id,
        name="Moscow",
        latitude=55.75,
        longitude=37.61,
        region="RU",
        active=True,
        created_at=datetime(2026, 1, 1, tzinfo=UTC),
        updated_at=datetime(2026, 1, 1, tzinfo=UTC),
    )


class FakeStationService:
    async def list_stations(self):
        return [_station_read(1)]

    async def list_active_stations(self):
        return [_station_read(1)]

    async def get_station(self, station_id):
        if station_id == 999:
            raise NotFoundError("not found")
        return _station_read(station_id)

    async def create_station(self, payload: StationCreate):
        return _station_read(10)

    async def update_station(self, station_id, payload: StationUpdate):
        if station_id == 999:
            raise NotFoundError("not found")
        return _station_read(station_id)

    async def delete_station(self, station_id):
        if station_id == 999:
            raise NotFoundError("not found")


@pytest.fixture
def client():
    app = FastAPI()
    app.include_router(stations.router)
    app.include_router(metrics.router)
    app.dependency_overrides[stations.get_station_service] = lambda: FakeStationService()
    with TestClient(app) as c:
        yield c


def test_list_stations(client):
    resp = client.get("/stations")
    assert resp.status_code == 200
    assert isinstance(resp.json(), list)
    assert resp.json()[0]["name"] == "Moscow"


def test_get_station(client):
    resp = client.get("/stations/1")
    assert resp.status_code == 200
    assert resp.json()["id"] == 1


def test_get_station_not_found(client):
    resp = client.get("/stations/999")
    assert resp.status_code == 404


def test_create_station(client):
    resp = client.post("/stations", json=STATION_JSON)
    assert resp.status_code == 201
    assert resp.json()["id"] == 10


def test_update_station(client):
    resp = client.put("/stations/1", json={"name": "Updated"})
    assert resp.status_code == 200
    assert resp.json()["name"] == "Moscow"


def test_update_station_not_found(client):
    resp = client.put("/stations/999", json={"name": "X"})
    assert resp.status_code == 404


def test_delete_station(client):
    resp = client.delete("/stations/1")
    assert resp.status_code == 204


def test_delete_station_not_found(client):
    resp = client.delete("/stations/999")
    assert resp.status_code == 404


def test_metrics_endpoint(client):
    resp = client.get("/metrics")
    assert resp.status_code == 200
    assert "text/plain" in resp.headers["content-type"]
