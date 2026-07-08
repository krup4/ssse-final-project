"""
Central logging configuration.

Configures structlog to emit structured JSON logs that can be consumed by
centralized log aggregation. All components obtain their logger via
``structlog.get_logger(__name__)`` so logs are consistent end-to-end.
"""

import logging
import sys

import structlog


def configure_logging(log_level: str = "INFO", service_name: str = "actual-weather") -> None:
    """Configure structlog with JSON output and sensible defaults.

    Safe to call multiple times; subsequent calls re-apply the configuration.
    """
    level = getattr(logging, log_level.upper(), logging.INFO)

    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=level,
    )

    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.stdlib.add_log_level,
            structlog.stdlib.add_logger_name,
            structlog.processors.TimeStamper(fmt="iso", utc=True),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(level),
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )

                                                   
    structlog.contextvars.bind_contextvars(service=service_name)
