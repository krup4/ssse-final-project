import type {
  Alert,
  ForecastErrorRow,
  HistoricalMetric,
  OverviewMetrics,
  ParameterErrorRow,
  ParameterErrorTrendPoint,
  Region,
  Station,
  StationSeriesPoint,
  User
} from "../entities/types";

export const users: User[] = [
  { id: "u-1", name: "Anna Admin", email: "admin@weather.local", role: "admin", lastSeen: "2026-07-07T09:10:00Z" },
  { id: "u-2", name: "Alex Analyst", email: "analyst@weather.local", role: "analyst", lastSeen: "2026-07-07T08:45:00Z" },
  { id: "u-3", name: "Olga Operator", email: "operator@weather.local", role: "operator", lastSeen: "2026-07-07T07:58:00Z" },
  { id: "u-4", name: "Vera Viewer", email: "viewer@weather.local", role: "viewer", lastSeen: "2026-07-06T18:20:00Z" }
];

export const regions: Region[] = [
  { id: "north", name: "Northern District" },
  { id: "central", name: "Central District" },
  { id: "south", name: "Southern District" },
  { id: "east", name: "Eastern District" }
];

export const stations: Station[] = [
  { id: "st-001", name: "Murmansk Port", regionId: "north", lat: 68.973, lon: 33.085, status: "degraded", activeSensors: 9, maxError: 18.4, lastTelemetryAt: "2026-07-07T09:12:00Z" },
  { id: "st-002", name: "Karelia Ridge", regionId: "north", lat: 61.785, lon: 34.346, status: "online", activeSensors: 12, maxError: 8.2, lastTelemetryAt: "2026-07-07T09:13:00Z" },
  { id: "st-003", name: "Moscow West", regionId: "central", lat: 55.753, lon: 37.421, status: "online", activeSensors: 14, maxError: 6.9, lastTelemetryAt: "2026-07-07T09:14:00Z" },
  { id: "st-004", name: "Ryazan Field", regionId: "central", lat: 54.626, lon: 39.735, status: "offline", activeSensors: 0, maxError: 21.1, lastTelemetryAt: "2026-07-07T07:35:00Z" },
  { id: "st-005", name: "Sochi Coast", regionId: "south", lat: 43.585, lon: 39.723, status: "online", activeSensors: 10, maxError: 11.3, lastTelemetryAt: "2026-07-07T09:13:00Z" },
  { id: "st-006", name: "Caspian Steppe", regionId: "south", lat: 46.349, lon: 48.041, status: "degraded", activeSensors: 7, maxError: 16.7, lastTelemetryAt: "2026-07-07T09:01:00Z" },
  { id: "st-007", name: "Baikal North", regionId: "east", lat: 53.558, lon: 108.165, status: "online", activeSensors: 11, maxError: 9.6, lastTelemetryAt: "2026-07-07T09:10:00Z" }
];

export const overview: OverviewMetrics = {
  activeStations: 5,
  degradedStations: 2,
  offlineStations: 1,
  kafkaLag: 2840,
  requestRate: 126,
  p95LatencyMs: 188,
  worstErrorToday: 21.1,
  errorTrend: [
    { date: "Jul 01", mae: 5.7, rmse: 7.4 },
    { date: "Jul 02", mae: 6.1, rmse: 7.8 },
    { date: "Jul 03", mae: 4.8, rmse: 6.3 },
    { date: "Jul 04", mae: 7.2, rmse: 9.2 },
    { date: "Jul 05", mae: 6.6, rmse: 8.7 },
    { date: "Jul 06", mae: 8.4, rmse: 11.5 },
    { date: "Jul 07", mae: 9.1, rmse: 12.8 }
  ]
};

export const worstErrors: ForecastErrorRow[] = [
  { id: "e-1", stationId: "st-004", stationName: "Ryazan Field", regionName: "Central District", metric: "wind_speed", forecastValue: 8.2, actualValue: 29.3, absoluteError: 21.1, errorPct: 257, observedAt: "2026-07-07T07:30:00Z" },
  { id: "e-2", stationId: "st-001", stationName: "Murmansk Port", regionName: "Northern District", metric: "temperature", forecastValue: 7.4, actualValue: -11.0, absoluteError: 18.4, errorPct: 249, observedAt: "2026-07-07T08:00:00Z" },
  { id: "e-3", stationId: "st-006", stationName: "Caspian Steppe", regionName: "Southern District", metric: "wind_speed", forecastValue: 5.1, actualValue: 21.8, absoluteError: 16.7, errorPct: 327, observedAt: "2026-07-07T08:20:00Z" },
  { id: "e-4", stationId: "st-005", stationName: "Sochi Coast", regionName: "Southern District", metric: "precipitation", forecastValue: 2.0, actualValue: 13.3, absoluteError: 11.3, errorPct: 565, observedAt: "2026-07-07T06:40:00Z" },
  { id: "e-5", stationId: "st-007", stationName: "Baikal North", regionName: "Eastern District", metric: "pressure", forecastValue: 1008.1, actualValue: 998.5, absoluteError: 9.6, errorPct: 1, observedAt: "2026-07-07T05:10:00Z" },
  { id: "e-6", stationId: "st-002", stationName: "Karelia Ridge", regionName: "Northern District", metric: "humidity", forecastValue: 62, actualValue: 70.2, absoluteError: 8.2, errorPct: 13, observedAt: "2026-07-07T04:50:00Z" }
];

export const parameterErrors: ParameterErrorRow[] = [
  { id: "pe-1", parameter: "temperature_min", stationId: "st-001", stationName: "Murmansk Port", regionName: "Northern District", forecastValue: 7.4, actualValue: -11.0, absoluteError: 18.4, errorPct: 249, contributionPct: 21, samples: 288, observedAt: "2026-07-07T08:00:00Z" },
  { id: "pe-2", parameter: "temperature_min", stationId: "st-007", stationName: "Baikal North", regionName: "Eastern District", forecastValue: 3.2, actualValue: -4.8, absoluteError: 8.0, errorPct: 250, contributionPct: 9, samples: 286, observedAt: "2026-07-07T05:00:00Z" },
  { id: "pe-3", parameter: "temperature_max", stationId: "st-006", stationName: "Caspian Steppe", regionName: "Southern District", forecastValue: 31.5, actualValue: 42.0, absoluteError: 10.5, errorPct: 33, contributionPct: 12, samples: 292, observedAt: "2026-07-07T12:00:00Z" },
  { id: "pe-4", parameter: "temperature_max", stationId: "st-003", stationName: "Moscow West", regionName: "Central District", forecastValue: 23.0, actualValue: 29.2, absoluteError: 6.2, errorPct: 27, contributionPct: 7, samples: 290, observedAt: "2026-07-07T14:00:00Z" },
  { id: "pe-5", parameter: "precipitation_total", stationId: "st-005", stationName: "Sochi Coast", regionName: "Southern District", forecastValue: 2.0, actualValue: 13.3, absoluteError: 11.3, errorPct: 565, contributionPct: 13, samples: 278, observedAt: "2026-07-07T06:40:00Z" },
  { id: "pe-6", parameter: "precipitation_total", stationId: "st-002", stationName: "Karelia Ridge", regionName: "Northern District", forecastValue: 0.0, actualValue: 6.7, absoluteError: 6.7, errorPct: 670, contributionPct: 8, samples: 280, observedAt: "2026-07-07T03:30:00Z" },
  { id: "pe-7", parameter: "wind_speed", stationId: "st-004", stationName: "Ryazan Field", regionName: "Central District", forecastValue: 8.2, actualValue: 29.3, absoluteError: 21.1, errorPct: 257, contributionPct: 24, samples: 224, observedAt: "2026-07-07T07:30:00Z" },
  { id: "pe-8", parameter: "wind_gust", stationId: "st-006", stationName: "Caspian Steppe", regionName: "Southern District", forecastValue: 11.0, actualValue: 34.8, absoluteError: 23.8, errorPct: 216, contributionPct: 27, samples: 220, observedAt: "2026-07-07T08:20:00Z" },
  { id: "pe-9", parameter: "humidity", stationId: "st-002", stationName: "Karelia Ridge", regionName: "Northern District", forecastValue: 62.0, actualValue: 70.2, absoluteError: 8.2, errorPct: 13, contributionPct: 9, samples: 288, observedAt: "2026-07-07T04:50:00Z" },
  { id: "pe-10", parameter: "pressure", stationId: "st-007", stationName: "Baikal North", regionName: "Eastern District", forecastValue: 1008.1, actualValue: 998.5, absoluteError: 9.6, errorPct: 1, contributionPct: 11, samples: 288, observedAt: "2026-07-07T05:10:00Z" }
];

export const parameterErrorTrend: ParameterErrorTrendPoint[] = [
  { timestamp: "00:00", parameter: "temperature_min", absoluteError: 4.2, mae: 3.1 },
  { timestamp: "03:00", parameter: "temperature_min", absoluteError: 6.8, mae: 4.4 },
  { timestamp: "06:00", parameter: "temperature_min", absoluteError: 18.4, mae: 8.9 },
  { timestamp: "09:00", parameter: "temperature_min", absoluteError: 9.6, mae: 6.2 },
  { timestamp: "12:00", parameter: "temperature_min", absoluteError: 5.3, mae: 4.1 },
  { timestamp: "15:00", parameter: "temperature_min", absoluteError: 3.7, mae: 3.5 },
  { timestamp: "18:00", parameter: "temperature_min", absoluteError: 6.2, mae: 4.0 },
  { timestamp: "21:00", parameter: "temperature_min", absoluteError: 7.1, mae: 4.7 },
  { timestamp: "00:00", parameter: "temperature_max", absoluteError: 3.8, mae: 2.9 },
  { timestamp: "03:00", parameter: "temperature_max", absoluteError: 4.9, mae: 3.2 },
  { timestamp: "06:00", parameter: "temperature_max", absoluteError: 6.7, mae: 4.1 },
  { timestamp: "09:00", parameter: "temperature_max", absoluteError: 8.4, mae: 5.0 },
  { timestamp: "12:00", parameter: "temperature_max", absoluteError: 10.5, mae: 6.7 },
  { timestamp: "15:00", parameter: "temperature_max", absoluteError: 9.3, mae: 5.8 },
  { timestamp: "18:00", parameter: "temperature_max", absoluteError: 5.8, mae: 4.0 },
  { timestamp: "21:00", parameter: "temperature_max", absoluteError: 4.1, mae: 3.3 },
  { timestamp: "00:00", parameter: "precipitation_total", absoluteError: 1.2, mae: 0.9 },
  { timestamp: "03:00", parameter: "precipitation_total", absoluteError: 6.7, mae: 3.5 },
  { timestamp: "06:00", parameter: "precipitation_total", absoluteError: 11.3, mae: 5.8 },
  { timestamp: "09:00", parameter: "precipitation_total", absoluteError: 8.1, mae: 4.6 },
  { timestamp: "12:00", parameter: "precipitation_total", absoluteError: 4.3, mae: 2.9 },
  { timestamp: "15:00", parameter: "precipitation_total", absoluteError: 3.1, mae: 2.1 },
  { timestamp: "18:00", parameter: "precipitation_total", absoluteError: 5.6, mae: 3.0 },
  { timestamp: "21:00", parameter: "precipitation_total", absoluteError: 2.8, mae: 1.8 },
  { timestamp: "00:00", parameter: "wind_speed", absoluteError: 2.1, mae: 1.7 },
  { timestamp: "03:00", parameter: "wind_speed", absoluteError: 4.3, mae: 2.8 },
  { timestamp: "06:00", parameter: "wind_speed", absoluteError: 21.1, mae: 9.9 },
  { timestamp: "09:00", parameter: "wind_speed", absoluteError: 12.4, mae: 7.2 },
  { timestamp: "12:00", parameter: "wind_speed", absoluteError: 6.0, mae: 4.1 },
  { timestamp: "15:00", parameter: "wind_speed", absoluteError: 4.8, mae: 3.5 },
  { timestamp: "18:00", parameter: "wind_speed", absoluteError: 3.2, mae: 2.4 },
  { timestamp: "21:00", parameter: "wind_speed", absoluteError: 2.6, mae: 2.0 },
  { timestamp: "00:00", parameter: "wind_gust", absoluteError: 3.6, mae: 2.7 },
  { timestamp: "03:00", parameter: "wind_gust", absoluteError: 8.8, mae: 4.9 },
  { timestamp: "06:00", parameter: "wind_gust", absoluteError: 16.5, mae: 8.1 },
  { timestamp: "09:00", parameter: "wind_gust", absoluteError: 23.8, mae: 11.6 },
  { timestamp: "12:00", parameter: "wind_gust", absoluteError: 14.0, mae: 7.4 },
  { timestamp: "15:00", parameter: "wind_gust", absoluteError: 9.2, mae: 5.6 },
  { timestamp: "18:00", parameter: "wind_gust", absoluteError: 5.1, mae: 3.7 },
  { timestamp: "21:00", parameter: "wind_gust", absoluteError: 4.4, mae: 3.2 },
  { timestamp: "00:00", parameter: "humidity", absoluteError: 3.0, mae: 2.2 },
  { timestamp: "03:00", parameter: "humidity", absoluteError: 5.4, mae: 3.7 },
  { timestamp: "06:00", parameter: "humidity", absoluteError: 8.2, mae: 5.1 },
  { timestamp: "09:00", parameter: "humidity", absoluteError: 6.9, mae: 4.6 },
  { timestamp: "12:00", parameter: "humidity", absoluteError: 4.5, mae: 3.4 },
  { timestamp: "15:00", parameter: "humidity", absoluteError: 4.0, mae: 3.1 },
  { timestamp: "18:00", parameter: "humidity", absoluteError: 5.8, mae: 4.0 },
  { timestamp: "21:00", parameter: "humidity", absoluteError: 4.7, mae: 3.5 },
  { timestamp: "00:00", parameter: "pressure", absoluteError: 3.2, mae: 2.6 },
  { timestamp: "03:00", parameter: "pressure", absoluteError: 5.7, mae: 3.8 },
  { timestamp: "06:00", parameter: "pressure", absoluteError: 9.6, mae: 5.7 },
  { timestamp: "09:00", parameter: "pressure", absoluteError: 7.2, mae: 4.8 },
  { timestamp: "12:00", parameter: "pressure", absoluteError: 5.3, mae: 3.9 },
  { timestamp: "15:00", parameter: "pressure", absoluteError: 4.1, mae: 3.3 },
  { timestamp: "18:00", parameter: "pressure", absoluteError: 6.0, mae: 4.2 },
  { timestamp: "21:00", parameter: "pressure", absoluteError: 4.9, mae: 3.6 }
];

export const stationSeries: StationSeriesPoint[] = [
  { timestamp: "00:00", forecast: 8.2, actual: 8.7, absoluteError: 0.5 },
  { timestamp: "03:00", forecast: 7.8, actual: 12.1, absoluteError: 4.3 },
  { timestamp: "06:00", forecast: 8.2, actual: 29.3, absoluteError: 21.1 },
  { timestamp: "09:00", forecast: 10.0, actual: 17.4, absoluteError: 7.4 },
  { timestamp: "12:00", forecast: 12.5, actual: 14.0, absoluteError: 1.5 },
  { timestamp: "15:00", forecast: 11.9, actual: 13.6, absoluteError: 1.7 },
  { timestamp: "18:00", forecast: 9.4, actual: 8.8, absoluteError: 0.6 },
  { timestamp: "21:00", forecast: 7.1, actual: 7.9, absoluteError: 0.8 }
];

export const history: HistoricalMetric[] = [
  { date: "2026-07-01", regionName: "Northern District", stationName: "Murmansk Port", mae: 5.8, rmse: 7.2, samples: 2304, backfillVersion: "calc-v3" },
  { date: "2026-07-02", regionName: "Central District", stationName: "Ryazan Field", mae: 6.6, rmse: 8.9, samples: 2260, backfillVersion: "calc-v3" },
  { date: "2026-07-03", regionName: "Southern District", stationName: "Sochi Coast", mae: 4.1, rmse: 5.3, samples: 2311, backfillVersion: "calc-v3" },
  { date: "2026-07-04", regionName: "Eastern District", stationName: "Baikal North", mae: 7.5, rmse: 9.7, samples: 2188, backfillVersion: "calc-v2" },
  { date: "2026-07-05", regionName: "Central District", stationName: "Moscow West", mae: 3.8, rmse: 4.9, samples: 2320, backfillVersion: "calc-v3" }
];

export const alerts: Alert[] = [
  { id: "a-1", title: "Kafka lag above SLO", source: "Telemetry Receiver", severity: "warning", status: "open", startedAt: "2026-07-07T08:50:00Z", updatedAt: "2026-07-07T09:10:00Z" },
  { id: "a-2", title: "Ryazan Field station offline", source: "Station Monitor", severity: "critical", status: "acknowledged", startedAt: "2026-07-07T07:36:00Z", updatedAt: "2026-07-07T08:05:00Z" },
  { id: "a-3", title: "Backfill completed for calc-v3", source: "ETL Processor", severity: "info", status: "resolved", startedAt: "2026-07-06T22:00:00Z", updatedAt: "2026-07-07T01:14:00Z" }
];
