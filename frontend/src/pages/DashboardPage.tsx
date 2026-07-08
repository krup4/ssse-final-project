import { useQuery } from "@tanstack/react-query";
import { Activity, Gauge, RadioTower, Server, Timer, TriangleAlert } from "lucide-react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { api } from "../shared/api/endpoints";
import { formatNumber } from "../shared/lib/format";
import { MetricCard } from "../shared/ui/MetricCard";
import { Panel } from "../shared/ui/Panel";

export function DashboardPage() {
  const { data } = useQuery({ queryKey: ["overview"], queryFn: api.overview });

  if (!data) {
    return <div className="page-grid">Loading dashboard...</div>;
  }

  return (
    <div className="page-grid">
      <div className="metrics-grid">
        <MetricCard icon={RadioTower} label="Active stations" value={String(data.activeStations)} hint="5 regions reporting" tone="good" />
        <MetricCard icon={TriangleAlert} label="Degraded/offline" value={String(data.degradedStations + data.offlineStations)} hint="Needs operator review" tone="warn" />
        <MetricCard icon={Server} label="Kafka lag" value={formatNumber(data.kafkaLag)} hint="Telemetry events pending" tone="warn" />
        <MetricCard icon={Timer} label="p95 latency" value={`${data.p95LatencyMs} ms`} hint={`${data.requestRate} req/min`} />
        <MetricCard icon={Gauge} label="Worst error today" value={`${data.worstErrorToday} m/s`} hint="Ryazan Field wind spike" tone="bad" />
        <MetricCard icon={Activity} label="SLO status" value="97.8%" hint="Forecast matching success" tone="good" />
      </div>
      <Panel title="Error trend">
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
              <XAxis dataKey="date" tickLine={false} axisLine={false} />
              <YAxis tickLine={false} axisLine={false} />
              <Tooltip />
              <Area type="monotone" dataKey="mae" stroke="#2f80ed" fill="url(#mae)" strokeWidth={2} />
              <Area type="monotone" dataKey="rmse" stroke="#d64545" fill="transparent" strokeWidth={2} />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </Panel>
    </div>
  );
}
