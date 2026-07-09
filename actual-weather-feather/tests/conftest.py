"""
Shared pytest fixtures and configuration.
"""

import pytest

from app.core.logging_config import configure_logging


@pytest.fixture(autouse=True)
def _configure_logging():
    configure_logging("INFO", "actual-weather-test")
    yield
