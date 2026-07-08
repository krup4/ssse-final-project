"""
AsyncPg mock helpers for repository/integration unit tests.

``FakeConn`` stores plain return values; its methods are coroutine functions,
so every call produces a fresh awaitable (mirroring asyncpg's per-call behaviour).
"""

from typing import Any


class FakeConn:
    def __init__(
        self,
        fetch: Any = None,
        fetchrow: Any = None,
        fetchval: Any = None,
        execute: Any = None,
    ):
        self._fetch = fetch if fetch is not None else []
        self._fetchrow = fetchrow
        self._fetchval = fetchval if fetchval is not None else 0
        self._execute = execute if execute is not None else ""

    async def fetch(self, *args, **kwargs):
        return self._fetch

    async def fetchrow(self, *args, **kwargs):
        return self._fetchrow

    async def fetchval(self, *args, **kwargs):
        return self._fetchval

    async def execute(self, *args, **kwargs):
        return self._execute


class FakeAcquire:
    def __init__(self, conn: FakeConn):
        self._conn = conn

    async def __aenter__(self):
        return self._conn

    async def __aexit__(self, *exc):
        return False


class FakePool:
    """asyncpg.Pool stand-in returning a single FakeConn from acquire()."""

    def __init__(self, conn: FakeConn):
        self._conn = conn

    def acquire(self):
        return FakeAcquire(self._conn)
