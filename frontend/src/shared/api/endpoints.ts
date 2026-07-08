import { apiClient } from "./client";
import type {
  Alert,
  ForecastErrorRow,
  HistoricalMetric,
  LoginRequest,
  LoginResponse,
  OverviewMetrics,
  ParameterErrorTrendPoint,
  ParameterErrorRow,
  Station,
  StationSeriesPoint,
  User,
  UserRole,
  AnalyticsFilters,
  ForecastField
} from "../../entities/types";

function filterParams(filters?: AnalyticsFilters) {
  if (!filters) {
    return undefined;
  }
  return {
    dateFrom: filters.dateFrom,
    dateTo: filters.dateTo,
    stationId: filters.stationId,
    metric: filters.metric
  };
}

export const api = {
  login: (payload: LoginRequest) => apiClient.post<LoginResponse>("/auth/login", payload).then((res) => res.data),
  me: () => apiClient.get<User>("/auth/me").then((res) => res.data),
  forecastFields: () => apiClient.get<ForecastField[]>("/forecast-fields").then((res) => res.data),
  stations: () => apiClient.get<Station[]>("/stations").then((res) => res.data),
  overview: (filters?: AnalyticsFilters) => apiClient.get<OverviewMetrics>("/metrics/overview", { params: filterParams(filters) }).then((res) => res.data),
  worstErrors: (filters?: AnalyticsFilters) => apiClient.get<ForecastErrorRow[]>("/analytics/worst-errors", { params: filterParams(filters) }).then((res) => res.data),
  parameterErrors: (parameter = "all", filters?: AnalyticsFilters) =>
    apiClient.get<ParameterErrorRow[]>("/analytics/parameter-errors", { params: { ...filterParams(filters), parameter } }).then((res) => res.data),
  parameterErrorTrend: (parameter = "all", filters?: AnalyticsFilters) =>
    apiClient.get<ParameterErrorTrendPoint[]>("/analytics/parameter-error-trend", { params: { ...filterParams(filters), parameter } }).then((res) => res.data),
  stationSeries: (filters?: AnalyticsFilters) => apiClient.get<StationSeriesPoint[]>("/analytics/station-series", { params: filterParams(filters) }).then((res) => res.data),
  history: (filters?: AnalyticsFilters) => apiClient.get<HistoricalMetric[]>("/analytics/history", { params: filterParams(filters) }).then((res) => res.data),
  alerts: () => apiClient.get<Alert[]>("/alerts").then((res) => res.data),
  users: () => apiClient.get<User[]>("/users").then((res) => res.data),
  updateUserRole: (id: string, role: UserRole) => apiClient.patch<User>(`/users/${id}/role`, { role }).then((res) => res.data)
};
