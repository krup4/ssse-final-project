import { useQuery } from "@tanstack/react-query";
import { Bar, BarChart, CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { metricLabels, weatherParameterLabels } from "../entities/labels";
import { useFilters } from "../features/filters/FiltersContext";
import { api } from "../shared/api/endpoints";
import { formatChartDate, formatTooltipValue } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";

export function ChartsPage() {
  const { filters } = useFilters();
  const { data = [] } = useQuery({ queryKey: ["station-series", filters], queryFn: () => api.stationSeries(filters) });
  const fieldTitle = filters.field === "all" ? "All parameters" : (weatherParameterLabels[filters.field] ?? filters.field);
  const metricTitle = filters.metric === "all" ? "All metrics" : (metricLabels[filters.metric] ?? filters.metric);

  return (
    <div className="page-grid two-columns">
      <Panel title={`${fieldTitle}, ${metricTitle}: metric trend`}>
        <div className="chart-tall">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
              <XAxis dataKey="timestamp" tickLine={false} axisLine={false} tickFormatter={formatChartDate} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} />
              <Tooltip formatter={formatTooltipValue} labelFormatter={(value) => formatChartDate(String(value))} />
              <Line type="monotone" dataKey="absoluteError" name={metricTitle} stroke="#2f80ed" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Panel>
      <Panel title={`${metricTitle} by time bucket`}>
        <div className="chart-tall">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data}>
              <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
              <XAxis dataKey="timestamp" tickLine={false} axisLine={false} tickFormatter={formatChartDate} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} />
              <Tooltip formatter={formatTooltipValue} labelFormatter={(value) => formatChartDate(String(value))} />
              <Bar dataKey="absoluteError" name={metricTitle} fill="#d64545" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Panel>
    </div>
  );
}
