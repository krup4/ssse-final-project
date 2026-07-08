import { useQuery } from "@tanstack/react-query";
import { metricLabels } from "../../entities/labels";
import type { MetricKey } from "../../entities/types";
import { api } from "../../shared/api/endpoints";
import { useFilters } from "./FiltersContext";

const metricOptions: MetricKey[] = ["temperature", "wind_speed", "humidity", "pressure", "precipitation"];

export function GlobalFilters() {
  const { filters, setFilters } = useFilters();
  const { data: regions = [] } = useQuery({ queryKey: ["regions"], queryFn: api.regions });
  const { data: stations = [] } = useQuery({ queryKey: ["stations"], queryFn: api.stations });

  const visibleStations = filters.regionId === "all" ? stations : stations.filter((station) => station.regionId === filters.regionId);

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
        Region
        <select
          value={filters.regionId}
          onChange={(event) => setFilters((current) => ({ ...current, regionId: event.target.value, stationId: "all" }))}
        >
          <option value="all">All regions</option>
          {regions.map((region) => (
            <option key={region.id} value={region.id}>
              {region.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Station
        <select
          value={filters.stationId}
          onChange={(event) => setFilters((current) => ({ ...current, stationId: event.target.value }))}
        >
          <option value="all">All stations</option>
          {visibleStations.map((station) => (
            <option key={station.id} value={station.id}>
              {station.name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Metric
        <select
          value={filters.metric}
          onChange={(event) => setFilters((current) => ({ ...current, metric: event.target.value as MetricKey }))}
        >
          {metricOptions.map((metric) => (
            <option key={metric} value={metric}>
              {metricLabels[metric]}
            </option>
          ))}
        </select>
      </label>
    </div>
  );
}
