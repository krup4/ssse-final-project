#!/usr/bin/env sh
set -eu

BATCHES="${BATCHES:-3}"
TOPIC="${KAFKA_ACTUAL_WEATHER_TOPIC:-actual-weather.raw.v1}"
BACKFILL_TOPIC="${KAFKA_BACKFILL_JOBS_TOPIC:-backfill.jobs.v1}"

docker compose exec redpanda rpk topic create "$TOPIC" --brokers redpanda:9092 >/dev/null 2>&1 || true
docker compose exec redpanda rpk topic create "$BACKFILL_TOPIC" --brokers redpanda:9092 >/dev/null 2>&1 || true

i=0
while [ "$i" -lt "$BATCHES" ]; do
  observed_at="$(date -u +"%Y-%m-%dT%H:%M:00Z")"
  event_suffix="$(date -u +%Y%m%d%H%M%S)-$i"

  docker compose exec -T redpanda rpk topic produce "$TOPIC" --brokers redpanda:9092 <<EOF
{"id":"dev-kafka-101-$event_suffix","stationId":"101","observedAt":"$observed_at","temperature":24.3,"windSpeed":31.8,"humidity":68.0,"pressure":1007.4,"source":"dev-seed","traceId":"dev-trace-101-$event_suffix"}
{"id":"dev-kafka-102-$event_suffix","stationId":"102","observedAt":"$observed_at","temperature":18.7,"windSpeed":9.4,"humidity":73.0,"pressure":987.2,"source":"dev-seed","traceId":"dev-trace-102-$event_suffix"}
{"id":"dev-kafka-103-$event_suffix","stationId":"103","observedAt":"$observed_at","temperature":-5.6,"windSpeed":14.2,"humidity":81.0,"pressure":1001.8,"source":"dev-seed","traceId":"dev-trace-103-$event_suffix"}
EOF

  i=$((i + 1))
  sleep 1
done

echo "Published $BATCHES batch(es) to $TOPIC"
