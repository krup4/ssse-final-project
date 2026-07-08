export type UserRole = "admin" | "analyst" | "operator" | "viewer";

export type MetricKey = "temperature" | "wind_speed" | "humidity" | "pressure" | "precipitation";

export type WeatherParameterKey = string;

export type StationStatus = "online" | "degraded" | "offline";

export type AlertSeverity = "critical" | "warning" | "info";

export type AlertStatus = "open" | "acknowledged" | "resolved";

export interface User {
  id: string;
  login: string;
  name: string;
  email: string;
  role: UserRole;
  isActive: boolean;
  lastSeen: string;
}

export interface Station {
  id: string;
  name: string;
  lat: number;
  lon: number;
  status: StationStatus;
  activeSensors: number;
  maxError: number;
  lastTelemetryAt: string;
}

export interface ForecastField {
  id: string;
  name: WeatherParameterKey;
}

export interface OverviewMetrics {
  activeStations: number;
  degradedStations: number;
  offlineStations: number;
  kafkaLag: number;
  requestRate: number;
  p95LatencyMs: number;
  worstErrorToday: number;
  errorTrend: Array<{ date: string; mae: number; rmse: number }>;
}

export interface ForecastErrorRow {
  id: string;
  stationId: string;
  stationName: string;
  metric: MetricKey;
  forecastValue: number;
  actualValue: number;
  absoluteError: number;
  errorPct: number;
  observedAt: string;
}

export interface ParameterErrorRow {
  id: string;
  parameter: WeatherParameterKey;
  stationId: string;
  stationName: string;
  forecastValue: number;
  actualValue: number;
  absoluteError: number;
  errorPct: number;
  contributionPct: number;
  samples: number;
  observedAt: string;
}

export interface ParameterErrorTrendPoint {
  timestamp: string;
  parameter: WeatherParameterKey;
  absoluteError: number;
  mae: number;
}

export interface StationSeriesPoint {
  timestamp: string;
  forecast: number;
  actual: number;
  absoluteError: number;
}

export interface HistoricalMetric {
  date: string;
  stationName: string;
  mae: number;
  rmse: number;
  samples: number;
  backfillVersion: string;
}

export interface Alert {
  id: string;
  title: string;
  source: string;
  severity: AlertSeverity;
  status: AlertStatus;
  startedAt: string;
  updatedAt: string;
}

export interface LoginRequest {
  login: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface AnalyticsFilters {
  dateFrom: string;
  dateTo: string;
  stationId: string;
  metric: MetricKey;
}
