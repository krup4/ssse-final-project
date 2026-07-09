"""
Kafka producer initialization and management.
Used for publishing weather events to Kafka.
"""

import asyncio
import json
import ssl
from typing import Optional

import structlog
from aiokafka import AIOKafkaProducer

from app.config.settings import Settings
from app.models.schemas import WeatherKafkaMessage

logger = structlog.get_logger(__name__)

                                 
_publisher: Optional["KafkaPublisher"] = None


class KafkaPublisher:
    """Kafka message publisher."""
    
    def __init__(self, producer: AIOKafkaProducer, topic: str) -> None:
        self._producer = producer
        self._topic = topic

    async def publish(self, message: WeatherKafkaMessage) -> None:
        """Publish weather message to Kafka."""
        payload = message.model_dump(mode="json")
        try:
            await self._producer.send_and_wait(
                self._topic,
                json.dumps(payload).encode("utf-8"),
            )
            logger.debug(f"Message published to {self._topic}: {payload}")
        except Exception as e:
            logger.error(f"Failed to publish to {self._topic}: {e}", exc_info=True)
            raise

    async def close(self) -> None:
        """Close Kafka publisher."""
        await self._producer.stop()
        logger.info("Kafka publisher stopped")


async def create_kafka_publisher(settings: Settings) -> KafkaPublisher:
    """Create Kafka publisher instance."""
    logger.info(f"Creating Kafka publisher with brokers: {settings.kafka_broker_list}")
    
    producer_kwargs = {
        "bootstrap_servers": settings.kafka_broker_list,
        "request_timeout_ms": settings.kafka_timeout_ms,
        "security_protocol": settings.kafka_security_protocol,
    }
    if "SSL" in settings.kafka_security_protocol:
        ssl_context = ssl.create_default_context()
        if settings.kafka_ssl_skip_verify:
            ssl_context.check_hostname = False
            ssl_context.verify_mode = ssl.CERT_NONE
        producer_kwargs["ssl_context"] = ssl_context
    if settings.kafka_username:
        producer_kwargs.update(
            {
                "sasl_mechanism": settings.kafka_sasl_mechanism,
                "sasl_plain_username": settings.kafka_username,
                "sasl_plain_password": settings.kafka_password,
            }
        )

    producer = AIOKafkaProducer(**producer_kwargs)
    await producer.start()
    logger.info("Kafka producer started")
    
    return KafkaPublisher(producer, settings.kafka_topic)


async def init_kafka(settings: Settings) -> KafkaPublisher:
    """Initialize global Kafka publisher."""
    global _publisher
    
    _publisher = await create_kafka_publisher(settings)
    return _publisher


async def close_kafka() -> None:
    """Close global Kafka publisher."""
    global _publisher
    if _publisher:
        await _publisher.close()
        _publisher = None


async def get_kafka() -> KafkaPublisher:
    """Get global Kafka publisher."""
    global _publisher
    if not _publisher:
        raise RuntimeError("Kafka publisher not initialized")
    return _publisher


async def check_kafka(brokers: list[str]) -> None:
    """Check Kafka connectivity."""
    if not brokers:
        raise RuntimeError("No Kafka brokers configured")
    
    host, _, port_str = brokers[0].partition(":")
    port = int(port_str or "9092")
    
    try:
        reader, writer = await asyncio.wait_for(
            asyncio.open_connection(host, port),
            timeout=5
        )
        writer.close()
        await writer.wait_closed()
        logger.info(f"Kafka connectivity check passed for {host}:{port}")
    except Exception as e:
        logger.error(f"Kafka connectivity check failed: {e}")
        raise
