import { useQuery } from "@tanstack/react-query";
import { Bar, BarChart, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { metricLabels } from "../entities/labels";
import { useFilters } from "../features/filters/FiltersContext";
import { api } from "../shared/api/endpoints";
import { formatChartDate } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";

export function ChartsPage() {
  const { filters } = useFilters();
  const { data = [] } = useQuery({ queryKey: ["station-series", filters], queryFn: () => api.stationSeries(filters) });

  return (
    <div className="page-grid two-columns">
      <Panel title={`${metricLabels[filters.metric]}: forecast vs actual`}>
        <div className="chart-tall">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
              <XAxis dataKey="timestamp" tickLine={false} axisLine={false} tickFormatter={formatChartDate} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} />
              <Tooltip labelFormatter={(value) => formatChartDate(String(value))} />
              <Legend />
              <Line type="monotone" dataKey="forecast" stroke="#2f80ed" strokeWidth={2} dot={false} />
              <Line type="monotone" dataKey="actual" stroke="#1f9d68" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Panel>
      <Panel title="Absolute error by time bucket">
        <div className="chart-tall">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data}>
              <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
              <XAxis dataKey="timestamp" tickLine={false} axisLine={false} tickFormatter={formatChartDate} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} />
              <Tooltip labelFormatter={(value) => formatChartDate(String(value))} />
              <Bar dataKey="absoluteError" fill="#d64545" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Panel>
    </div>
  );
}
