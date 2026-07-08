import type {
  Alert,
  ForecastErrorRow,
  HistoricalMetric,
  OverviewMetrics,
  ParameterErrorRow,
  ParameterErrorTrendPoint,
  Station,
  StationSeriesPoint,
  User,
  ForecastField,
  MetricDefinition
} from "../entities/types";

export const users: User[] = [
  { id: "1", login: "admin", name: "Ada Admin", email: "admin@weather.local", role: "admin", isActive: true, lastSeen: "2026-07-08T09:10:00Z" },
  { id: "2", login: "analyst", name: "Alex Analyst", email: "analyst@weather.local", role: "analyst", isActive: true, lastSeen: "2026-07-08T08:45:00Z" },
  { id: "3", login: "operator", name: "Olga Operator", email: "operator@weather.local", role: "operator", isActive: true, lastSeen: "2026-07-08T07:58:00Z" },
  { id: "4", login: "viewer", name: "Vera Viewer", email: "viewer@weather.local", role: "viewer", isActive: true, lastSeen: "2026-07-07T18:20:00Z" }
];

export const stations: Station[] = [
  { id: "1", name: "Ryazan Field", lat: 54.626, lon: 39.735, isActive: true, status: "online", activeSensors: 1, maxError: 21.1, lastTelemetryAt: "2026-07-08T09:00:00Z" },
  { id: "2", name: "Caspian Steppe", lat: 46.349, lon: 48.041, isActive: true, status: "online", activeSensors: 1, maxError: 16.7, lastTelemetryAt: "2026-07-08T09:00:00Z" }
];

export const forecastFields: ForecastField[] = [
  { id: "1", name: "temperature" },
  { id: "2", name: "wind_speed" },
  { id: "3", name: "humidity" },
  { id: "4", name: "pressure" }
];

export const metrics: MetricDefinition[] = forecastFields.flatMap((field) =>
  ["mae", "mse", "rmse"].map((name, index) => ({
    id: `${field.id}-${index + 1}`,
    forecastFieldId: field.id,
    forecastField: field.name,
    name
  }))
);

export const overview: OverviewMetrics = {
  activeStations: 2,
  degradedStations: 0,
  offlineStations: 0,
  kafkaLag: 42,
  requestRate: 128,
  p95LatencyMs: 83,
  worstErrorToday: 21.1,
  errorTrend: [
    { date: "2026-07-06", mae: 6.4, rmse: 8.1 },
    { date: "2026-07-07", mae: 7.2, rmse: 9.3 },
    { date: "2026-07-08", mae: 5.9, rmse: 7.4 }
  ]
};

export const worstErrors: ForecastErrorRow[] = [
  { id: "e-1", stationId: "1", stationName: "Ryazan Field", parameter: "wind_speed", metric: "mae", forecastValue: 8.2, actualValue: 29.3, absoluteError: 21.1, errorPct: 257, observedAt: "2026-07-08T07:30:00Z" },
  { id: "e-2", stationId: "2", stationName: "Caspian Steppe", parameter: "pressure", metric: "mae", forecastValue: 1008.1, actualValue: 998.5, absoluteError: 9.6, errorPct: 1, observedAt: "2026-07-08T05:10:00Z" }
];

export const parameterErrors: ParameterErrorRow[] = [
  { id: "pe-1", parameter: "temperature", stationId: "1", stationName: "Ryazan Field", forecastValue: 19.4, actualValue: 24.1, absoluteError: 4.7, errorPct: 24, contributionPct: 18, samples: 1, observedAt: "2026-07-08T08:00:00Z" },
  { id: "pe-2", parameter: "wind_speed", stationId: "1", stationName: "Ryazan Field", forecastValue: 8.2, actualValue: 29.3, absoluteError: 21.1, errorPct: 257, contributionPct: 54, samples: 1, observedAt: "2026-07-08T07:30:00Z" },
  { id: "pe-3", parameter: "humidity", stationId: "2", stationName: "Caspian Steppe", forecastValue: 62, actualValue: 70.2, absoluteError: 8.2, errorPct: 13, contributionPct: 28, samples: 1, observedAt: "2026-07-08T04:50:00Z" }
];

export const parameterErrorTrend: ParameterErrorTrendPoint[] = [
  { timestamp: "2026-07-08T00:00:00Z", parameter: "temperature", absoluteError: 4.2, mae: 3.1 },
  { timestamp: "2026-07-08T03:00:00Z", parameter: "temperature", absoluteError: 6.8, mae: 4.4 },
  { timestamp: "2026-07-08T06:00:00Z", parameter: "wind_speed", absoluteError: 21.1, mae: 8.9 },
  { timestamp: "2026-07-08T09:00:00Z", parameter: "humidity", absoluteError: 8.2, mae: 6.2 }
];

export const stationSeries: StationSeriesPoint[] = [
  { timestamp: "2026-07-08T00:00:00Z", forecast: 8, actual: 12, absoluteError: 4 },
  { timestamp: "2026-07-08T03:00:00Z", forecast: 9, actual: 15, absoluteError: 6 },
  { timestamp: "2026-07-08T06:00:00Z", forecast: 8.2, actual: 29.3, absoluteError: 21.1 },
  { timestamp: "2026-07-08T09:00:00Z", forecast: 11, actual: 14, absoluteError: 3 }
];

export const history: HistoricalMetric[] = [
  { date: "2026-07-06", stationName: "Ryazan Field", mae: 5.8, rmse: 7.2, samples: 120, backfillVersion: "calc-v1" },
  { date: "2026-07-07", stationName: "Caspian Steppe", mae: 6.6, rmse: 8.9, samples: 120, backfillVersion: "calc-v1" },
  { date: "2026-07-08", stationName: "Ryazan Field", mae: 4.1, rmse: 5.3, samples: 80, backfillVersion: "calc-v1" }
];

export const alerts: Alert[] = [
  { id: "a-1", title: "Kafka lag above SLO", source: "Telemetry Receiver", severity: "warning", status: "open", startedAt: "2026-07-08T08:20:00Z", updatedAt: "2026-07-08T08:55:00Z" }
];
