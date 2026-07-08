import { useQuery } from "@tanstack/react-query";
import { Activity, Gauge, RadioTower, Server, Timer, TriangleAlert } from "lucide-react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { api } from "../shared/api/endpoints";
import { useFilters } from "../features/filters/FiltersContext";
import { metricLabels } from "../entities/labels";
import { formatChartDate, formatNumber, formatTooltipValue } from "../shared/lib/format";
import { MetricCard } from "../shared/ui/MetricCard";
import { Panel } from "../shared/ui/Panel";

export function DashboardPage() {
  const { filters } = useFilters();
  const { data } = useQuery({ queryKey: ["overview", filters], queryFn: () => api.overview(filters) });
  const metricTitle = filters.metric === "all" ? "All metrics" : (metricLabels[filters.metric] ?? filters.metric);

  if (!data) {
    return <div className="page-grid">Loading dashboard...</div>;
  }

  return (
    <div className="page-grid">
      <div className="metrics-grid">
        <MetricCard icon={RadioTower} label="Active stations" value={String(data.activeStations)} hint="Reporting now" tone="good" />
        <MetricCard icon={TriangleAlert} label="Degraded/offline" value={String(data.degradedStations + data.offlineStations)} hint="Needs operator review" tone="warn" />
        <MetricCard icon={Server} label="Kafka lag" value={formatNumber(data.kafkaLag)} hint="Telemetry events pending" tone="warn" />
        <MetricCard icon={Timer} label="p95 latency" value={`${formatNumber(data.p95LatencyMs)} ms`} hint={`${formatNumber(data.requestRate)} req/min`} />
        <MetricCard icon={Gauge} label="Worst error today" value={formatNumber(data.worstErrorToday)} hint="Max aggregated forecast error" tone="bad" />
        <MetricCard icon={Activity} label="SLO status" value="97.8%" hint="Forecast matching success" tone="good" />
      </div>
      <Panel title={`${metricTitle} trend`}>
        <div className="chart-tall">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={data.errorTrend}>
              <defs>
                <linearGradient id="mae" x1="0" x2="0" y1="0" y2="1">
                  <stop offset="5%" stopColor="#2f80ed" stopOpacity={0.35} />
                  <stop offset="95%" stopColor="#2f80ed" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#d9e2ec" />
              <XAxis dataKey="date" tickLine={false} axisLine={false} tickFormatter={formatChartDate} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} />
              <Tooltip formatter={formatTooltipValue} labelFormatter={(value) => formatChartDate(String(value))} />
              <Area type="monotone" dataKey="mae" name={filters.metric === "all" ? "Avg metric value" : metricTitle} stroke="#2f80ed" fill="url(#mae)" strokeWidth={2} />
              {filters.metric === "all" ? <Area type="monotone" dataKey="rmse" name="Metric spread" stroke="#d64545" fill="transparent" strokeWidth={2} /> : null}
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </Panel>
    </div>
  );
}
