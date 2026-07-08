import { useEffect, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { RotateCcw } from "lucide-react";
import { metricLabels, weatherParameterLabels } from "../../entities/labels";
import { api } from "../../shared/api/endpoints";
import { defaultFilters, useFilters } from "./FiltersContext";

export function GlobalFilters() {
  const { filters, setFilters } = useFilters();
  const { data: stations = [] } = useQuery({ queryKey: ["stations", "online"], queryFn: () => api.stations("online") });
  const { data: forecastFields = [] } = useQuery({ queryKey: ["forecast-fields"], queryFn: api.forecastFields });
  const { data: metrics = [] } = useQuery({ queryKey: ["metrics-catalog"], queryFn: api.metrics });
  const metricOptions = useMemo(() => {
    const allowed = filters.field === "all" ? metrics : metrics.filter((metric) => metric.forecastField === filters.field);
    return [...new Set(allowed.map((metric) => metric.name))].sort();
  }, [filters.field, metrics]);

  useEffect(() => {
    if (filters.field !== "all" && forecastFields.length > 0 && !forecastFields.some((field) => field.name === filters.field)) {
      setFilters((current) => ({ ...current, field: "all" }));
    }
  }, [filters.field, forecastFields, setFilters]);

  useEffect(() => {
    if (filters.metric !== "all" && metricOptions.length > 0 && !metricOptions.includes(filters.metric)) {
      setFilters((current) => ({ ...current, metric: "all" }));
    }
  }, [filters.metric, metricOptions, setFilters]);

  useEffect(() => {
    if (filters.stationId !== "all" && stations.length > 0 && !stations.some((station) => station.id === filters.stationId)) {
      setFilters((current) => ({ ...current, stationId: "all" }));
    }
  }, [filters.stationId, setFilters, stations]);

  return (
    <div className="global-filters">
      <label>
        From
        <input
          type="date"
          value={filters.dateFrom}
          onChange={(event) => setFilters((current) => ({ ...current, dateFrom: event.target.value }))}
        />
      </label>
      <label>
        To
        <input
          type="date"
          value={filters.dateTo}
          onChange={(event) => setFilters((current) => ({ ...current, dateTo: event.target.value }))}
        />
      </label>
      <label>
        Station
        <select
          value={filters.stationId}
          onChange={(event) => setFilters((current) => ({ ...current, stationId: event.target.value }))}
        >
          <option value="all">All stations</option>
          {stations.map((station) => (
            <option key={station.id} value={station.id}>
              {station.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Parameter
        <select
          value={filters.field}
          onChange={(event) => setFilters((current) => ({ ...current, field: event.target.value }))}
        >
          <option value="all">All parameters</option>
          {forecastFields.map((field) => (
            <option key={field.id} value={field.name}>
              {weatherParameterLabels[field.name] ?? field.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Metric
        <select
          value={filters.metric}
          onChange={(event) => setFilters((current) => ({ ...current, metric: event.target.value }))}
        >
          <option value="all">All metrics</option>
          {metricOptions.map((metric) => (
            <option key={metric} value={metric}>
              {metricLabels[metric] ?? metric}
            </option>
          ))}
        </select>
      </label>
      <button
        type="button"
        className="ghost-button filter-reset-button"
        onClick={() => setFilters(defaultFilters())}
      >
        <RotateCcw size={16} />
        Reset
      </button>
    </div>
  );
}
