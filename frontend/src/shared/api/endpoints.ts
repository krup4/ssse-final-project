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
  Region,
  Station,
  StationSeriesPoint,
  User,
  UserRole
} from "../../entities/types";

export const api = {
  login: (payload: LoginRequest) => apiClient.post<LoginResponse>("/auth/login", payload).then((res) => res.data),
  me: () => apiClient.get<User>("/auth/me").then((res) => res.data),
  regions: () => apiClient.get<Region[]>("/regions").then((res) => res.data),
  stations: () => apiClient.get<Station[]>("/stations").then((res) => res.data),
  overview: () => apiClient.get<OverviewMetrics>("/metrics/overview").then((res) => res.data),
  worstErrors: () => apiClient.get<ForecastErrorRow[]>("/analytics/worst-errors").then((res) => res.data),
  parameterErrors: (parameter = "all") =>
    apiClient.get<ParameterErrorRow[]>("/analytics/parameter-errors", { params: { parameter } }).then((res) => res.data),
  parameterErrorTrend: (parameter = "all") =>
    apiClient.get<ParameterErrorTrendPoint[]>("/analytics/parameter-error-trend", { params: { parameter } }).then((res) => res.data),
  stationSeries: () => apiClient.get<StationSeriesPoint[]>("/analytics/station-series").then((res) => res.data),
  history: () => apiClient.get<HistoricalMetric[]>("/analytics/history").then((res) => res.data),
  alerts: () => apiClient.get<Alert[]>("/alerts").then((res) => res.data),
  users: () => apiClient.get<User[]>("/users").then((res) => res.data),
  updateUserRole: (id: string, role: UserRole) => apiClient.patch<User>(`/users/${id}/role`, { role }).then((res) => res.data)
};
