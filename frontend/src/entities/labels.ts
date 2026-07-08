import type { MetricKey, UserRole, WeatherParameterKey } from "./types";

export const metricLabels: Record<MetricKey, string> = {
  temperature: "Temperature",
  wind_speed: "Wind speed",
  humidity: "Humidity",
  pressure: "Pressure",
  precipitation: "Precipitation"
};

export const metricUnits: Record<MetricKey, string> = {
  temperature: "C",
  wind_speed: "m/s",
  humidity: "%",
  pressure: "hPa",
  precipitation: "mm"
};

export const weatherParameterLabels: Record<string, string> = {
  temperature: "Temperature",
  precipitation: "Precipitation",
  precipitation_total: "Precipitation",
  wind_speed: "Wind speed",
  wind_gust: "Wind gust",
  humidity: "Humidity",
  pressure: "Pressure"
};

export const weatherParameterUnits: Record<string, string> = {
  temperature: "C",
  precipitation: "mm",
  precipitation_total: "mm",
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
