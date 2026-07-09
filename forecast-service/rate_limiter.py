import asyncio
import time


class AsyncRateLimiter:
    def __init__(self, requests_per_second: float) -> None:
        self._interval = 0.0 if requests_per_second <= 0 else 1.0 / requests_per_second
        self._lock = asyncio.Lock()
        self._next_at = 0.0

    async def wait(self) -> None:
        if self._interval <= 0:
            return

        async with self._lock:
            now = time.monotonic()
            delay = self._next_at - now
            if delay > 0:
                await asyncio.sleep(delay)
                now = time.monotonic()
            self._next_at = max(now, self._next_at) + self._interval
