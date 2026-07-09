from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    app_name: str = "Actual Weather Feather"
    app_version: str = "1.0.0"
    port: int = 8080
    host: str = "0.0.0.0"
    reload: bool = False
    log_level: str = "INFO"

    postgres_host: str = "postgres"
    postgres_port: int = 5432
    postgres_user: str = "postgres"
    postgres_password: str = "password"
    postgres_db: str = "weather"
    database_url: str | None = None

    redis_host: str = "redis"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0
    redis_url: str | None = None
    redis_ttl_seconds: int = 3600

    kafka_brokers: str = "kafka:9092"
    kafka_topic: str = "weather.actual"
    kafka_timeout_ms: int = 10000
    kafka_security_protocol: str = "PLAINTEXT"
    kafka_sasl_mechanism: str = "SCRAM-SHA-512"
    kafka_username: str = ""
    kafka_password: str = ""
    kafka_ssl_skip_verify: bool = False

    yandex_weather_url: str = "https://api.weather.yandex.ru/v2/informers"
    yandex_weather_graphql_url: str = "https://api.weather.yandex.ru/graphql/query"
    yandex_weather_api_key: str = ""
    yandex_weather_rate_limit_rps: float = 1.0
    request_timeout: float = 10.0
    http_retries: int = 3
    retry_backoff_factor: float = 2.0
    update_interval_seconds: int = 3600
    station_name_filter: str | None = None
    scheduler_lock_enabled: bool = True
    scheduler_lock_key: str = "actual-weather-feather:scheduler:leader"
    scheduler_lock_ttl_seconds: int = 3900

    temp_min: float = -100.0
    temp_max: float = 100.0
    humidity_min: float = 0.0
    humidity_max: float = 100.0
    pressure_min: float = 300.0       
    pressure_max: float = 1100.0       
    wind_speed_max: float = 200.0       

    def __init__(self, **data):
        super().__init__(**data)

                                                
        if not self.database_url:
            self.database_url = (
                f"postgresql+asyncpg://{self.postgres_user}:{self.postgres_password}"
                f"@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
            )

                                             
        if not self.redis_url:
            auth = f":{self.redis_password}@" if self.redis_password else ""
            self.redis_url = f"redis://{auth}{self.redis_host}:{self.redis_port}/{self.redis_db}"

    @property
    def kafka_broker_list(self) -> list[str]:
        """Get list of Kafka brokers."""
        return [b.strip() for b in self.kafka_brokers.split(",") if b.strip()]


def get_settings() -> Settings:
    """Get or create application settings instance."""
    return Settings()


                          
settings = get_settings()
