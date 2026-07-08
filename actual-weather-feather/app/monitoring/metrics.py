from prometheus_client import Counter, Histogram

WEATHER_REQUESTS_TOTAL = Counter(
    "weather_requests_total",
    "Total weather requests to external provider",
)
ERRORS_TOTAL = Counter(
    "errors_total",
    "Total errors",
)
STATIONS_PROCESSED_TOTAL = Counter(
    "stations_processed_total",
    "Total stations processed by scheduler",
)
RETRY_TOTAL = Counter(
    "retry_total",
    "Total retry attempts for external calls",
)
REQUEST_DURATION_SECONDS = Histogram(
    "request_duration_seconds",
    "Duration of external weather requests in seconds",
    buckets=(0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0),
)
HTTP_REQUESTS_TOTAL = Counter(
    "http_requests_total",
    "Total HTTP requests handled by the service",
    ["method", "endpoint", "status_code"],
)
HTTP_REQUEST_DURATION_SECONDS = Histogram(
    "http_request_duration_seconds",
    "Duration of HTTP requests in seconds",
    ["method", "endpoint"],
    buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
)
