import type { UserRole } from "./types";

export const metricLabels: Record<string, string> = {
  mae: "MAE",
  mse: "MSE",
  rmse: "RMSE"
};

export const metricUnits: Record<string, string> = {};

export const weatherParameterLabels: Record<string, string> = {
  temperature: "Temperature",
  wind_speed: "Wind speed",
  wind_gust: "Wind gust",
  humidity: "Humidity",
  pressure: "Pressure"
};

export const weatherParameterUnits: Record<string, string> = {
  temperature: "C",
  wind_speed: "m/s",
  wind_gust: "m/s",
  humidity: "%",
  pressure: "hPa"
};

export const roleLabels: Record<UserRole, string> = {
  admin: "Admin",
  analyst: "Analyst",
  operator: "Operator",
  viewer: "Viewer"
};
