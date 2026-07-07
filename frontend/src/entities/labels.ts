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

export const weatherParameterLabels: Record<WeatherParameterKey, string> = {
  temperature_min: "Minimum temperature",
  temperature_max: "Maximum temperature",
  precipitation_total: "Precipitation",
  wind_speed: "Wind speed",
  wind_gust: "Wind gust",
  humidity: "Humidity",
  pressure: "Pressure"
};

export const weatherParameterUnits: Record<WeatherParameterKey, string> = {
  temperature_min: "C",
  temperature_max: "C",
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
