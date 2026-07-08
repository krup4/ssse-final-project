import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ArrowDownUp } from "lucide-react";
import { metricLabels, weatherParameterLabels, weatherParameterUnits } from "../entities/labels";
import { useFilters } from "../features/filters/FiltersContext";
import { api } from "../shared/api/endpoints";
import { formatDateTime, formatNumber } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";

export function ErrorsPage() {
  const [descending, setDescending] = useState(true);
  const { filters } = useFilters();
  const { data = [] } = useQuery({ queryKey: ["worst-errors", filters], queryFn: () => api.worstErrors(filters) });
  const metricTitle = filters.metric === "all" ? "Metric value" : (metricLabels[filters.metric] ?? filters.metric);
  const rows = useMemo(
    () => [...data].sort((a, b) => (descending ? b.absoluteError - a.absoluteError : a.absoluteError - b.absoluteError)),
    [data, descending]
  );

  return (
    <div className="page-grid">
      <Panel
        title="Top forecast errors"
        action={
          <button type="button" className="ghost-button" onClick={() => setDescending((value) => !value)}>
            <ArrowDownUp size={16} /> Sort
          </button>
        }
      >
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Station</th>
                <th>Parameter</th>
                <th>Metric</th>
                <th>Forecast</th>
                <th>{metricTitle}</th>
                <th>Observed</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id}>
                  <td>{row.stationName}</td>
                  <td>{weatherParameterLabels[row.parameter] ?? row.parameter}</td>
                  <td>{metricLabels[row.metric] ?? row.metric}</td>
                  <td>{formatNumber(row.forecastValue)} {weatherParameterUnits[row.parameter] ?? ""}</td>
                  <td><strong>{formatNumber(row.absoluteError)}</strong></td>
                  <td>{formatDateTime(row.observedAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  );
}
