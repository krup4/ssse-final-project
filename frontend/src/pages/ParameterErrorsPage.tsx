import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Bar, BarChart, CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { SlidersHorizontal } from "lucide-react";
import { weatherParameterLabels, weatherParameterUnits } from "../entities/labels";
import type { WeatherParameterKey } from "../entities/types";
import { api } from "../shared/api/endpoints";
import { formatDateTime, formatNumber } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";

const parameterOptions: Array<WeatherParameterKey | "all"> = [
  "all",
  "temperature_min",
  "temperature_max",
  "precipitation_total",
  "wind_speed",
  "wind_gust",
  "humidity",
  "pressure"
];

function optionLabel(parameter: WeatherParameterKey | "all") {
  return parameter === "all" ? "All parameters" : weatherParameterLabels[parameter];
}

export function ParameterErrorsPage() {
  const [parameter, setParameter] = useState<WeatherParameterKey | "all">("all");
  const { data = [] } = useQuery({
    queryKey: ["parameter-errors", parameter],
    queryFn: () => api.parameterErrors(parameter)
  });
  const { data: trend = [] } = useQuery({
    queryKey: ["parameter-error-trend", parameter],
    queryFn: () => api.parameterErrorTrend(parameter)
  });

  const summaryRows = useMemo(() => {
    const groups = new Map<WeatherParameterKey, { parameter: WeatherParameterKey; maxError: number; avgError: number; contributionPct: number; count: number }>();

    data.forEach((row) => {
      const current = groups.get(row.parameter) ?? {
        parameter: row.parameter,
        maxError: 0,
        avgError: 0,
        contributionPct: 0,
        count: 0
      };
      current.maxError = Math.max(current.maxError, row.absoluteError);
      current.avgError += row.absoluteError;
      current.contributionPct += row.contributionPct;
      current.count += 1;
      groups.set(row.parameter, current);
    });

    return [...groups.values()]
      .map((row) => ({
        ...row,
        label: weatherParameterLabels[row.parameter],
        avgError: row.count > 0 ? row.avgError / row.count : 0
      }))
      .sort((a, b) => b.contributionPct - a.contributionPct);
  }, [data]);

  const worstRow = data.reduce((best, row) => (row.absoluteError > (best?.absoluteError ?? -1) ? row : best), data[0]);

  const trendRows = useMemo(() => {
    if (parameter !== "all") {
      return trend.map((row) => ({
        timestamp: row.timestamp,
        absoluteError: row.absoluteError,
        mae: row.mae
      }));
    }

    const grouped = new Map<string, { timestamp: string; maxError: number; mae: number; count: number }>();
    trend.forEach((row) => {
      const current = grouped.get(row.timestamp) ?? { timestamp: row.timestamp, maxError: 0, mae: 0, count: 0 };
      current.maxError = Math.max(current.maxError, row.absoluteError);
      current.mae += row.mae;
      current.count += 1;
      grouped.set(row.timestamp, current);
    });

    return [...grouped.values()].map((row) => ({
      timestamp: row.timestamp,
      absoluteError: row.maxError,
      mae: row.count > 0 ? row.mae / row.count : 0
    }));
  }, [parameter, trend]);

  return (
    <div className="page-grid">
      <Panel
        title="Errors by weather parameter"
        action={
          <div className="panel-action-label">
            <SlidersHorizontal size={16} />
            Parameter details
          </div>
        }
      >
        <div className="segmented-control" role="tablist" aria-label="Weather parameters">
          {parameterOptions.map((option) => (
            <button
              key={option}
              type="button"
              className={option === parameter ? "active" : ""}
              onClick={() => setParameter(option)}
            >
              {optionLabel(option)}
            </button>
          ))}
        </div>
      </Panel>

      <div className="parameter-summary-grid">
        <section className="parameter-summary">
          <span>Worst parameter error</span>
          <strong>{worstRow ? formatNumber(worstRow.absoluteError) : "0"}</strong>
          <p>{worstRow ? `${weatherParameterLabels[worstRow.parameter]} at ${worstRow.stationName}` : "No data"}</p>
        </section>
        <section className="parameter-summary">
          <span>Rows in selection</span>
          <strong>{data.length}</strong>
          <p>Filtered by selected parameter and global period</p>
        </section>
        <section className="parameter-summary">
          <span>Highest contribution</span>
          <strong>{summaryRows[0] ? `${formatNumber(summaryRows[0].contributionPct)}%` : "0%"}</strong>
          <p>{summaryRows[0]?.label ?? "No data"}</p>
        </section>
      </div>

      <div className="page-grid two-columns">
        <Panel title="Contribution by parameter">
          <div className="chart-tall">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={summaryRows}>
                <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
                <XAxis dataKey="label" tickLine={false} axisLine={false} interval={0} angle={-18} textAnchor="end" height={78} />
                <YAxis tickLine={false} axisLine={false} />
                <Tooltip />
                <Bar dataKey="contributionPct" name="Contribution, %" fill="#2f80ed" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Panel>

        <Panel title={parameter === "all" ? "Aggregate error trend" : `${weatherParameterLabels[parameter]} error trend`}>
          <div className="chart-tall">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trendRows}>
                <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
                <XAxis dataKey="timestamp" tickLine={false} axisLine={false} />
                <YAxis tickLine={false} axisLine={false} />
                <Tooltip />
                <Line type="monotone" dataKey="absoluteError" name="Max absolute error" stroke="#d64545" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="mae" name="MAE" stroke="#2f80ed" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Panel>
      </div>

      <div className="page-grid">
        <Panel title="Parameter error rows">
          <div className="table-wrap compact-table">
            <table>
              <thead>
                <tr>
                  <th>Parameter</th>
                  <th>Station</th>
                  <th>Forecast</th>
                  <th>Actual</th>
                  <th>Error</th>
                  <th>Contribution</th>
                </tr>
              </thead>
              <tbody>
                {data.map((row) => (
                  <tr key={row.id}>
                    <td>{weatherParameterLabels[row.parameter]}</td>
                    <td>{row.stationName}</td>
                    <td>{formatNumber(row.forecastValue)} {weatherParameterUnits[row.parameter]}</td>
                    <td>{formatNumber(row.actualValue)} {weatherParameterUnits[row.parameter]}</td>
                    <td><strong>{formatNumber(row.absoluteError)}</strong></td>
                    <td>{formatNumber(row.contributionPct)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Panel>
      </div>

      <Panel title="Detailed observations">
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Parameter</th>
                <th>Region</th>
                <th>Station</th>
                <th>Forecast</th>
                <th>Actual</th>
                <th>Abs. error</th>
                <th>Error %</th>
                <th>Samples</th>
                <th>Observed</th>
              </tr>
            </thead>
            <tbody>
              {data.map((row) => (
                <tr key={`detail-${row.id}`}>
                  <td>{weatherParameterLabels[row.parameter]}</td>
                  <td>{row.regionName}</td>
                  <td>{row.stationName}</td>
                  <td>{formatNumber(row.forecastValue)} {weatherParameterUnits[row.parameter]}</td>
                  <td>{formatNumber(row.actualValue)} {weatherParameterUnits[row.parameter]}</td>
                  <td><strong>{formatNumber(row.absoluteError)}</strong></td>
                  <td>{formatNumber(row.errorPct)}%</td>
                  <td>{row.samples.toLocaleString("en")}</td>
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
