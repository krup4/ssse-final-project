import { useQuery } from "@tanstack/react-query";
import { api } from "../shared/api/endpoints";
import { formatDateTime } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";
import { AlertSeverityBadge, AlertStatusBadge } from "../shared/ui/StatusBadge";

export function AlertsPage() {
  const { data = [] } = useQuery({ queryKey: ["alerts"], queryFn: api.alerts });

  return (
    <div className="page-grid">
      <Panel title="Alerts">
        <div className="alert-list">
          {data.map((alert) => (
            <article key={alert.id} className="alert-row">
              <div>
                <h3>{alert.title}</h3>
                <p>{alert.source}</p>
              </div>
              <AlertSeverityBadge severity={alert.severity} />
              <AlertStatusBadge status={alert.status} />
              <span>{formatDateTime(alert.startedAt)}</span>
              <span>{formatDateTime(alert.updatedAt)}</span>
            </article>
          ))}
        </div>
      </Panel>
    </div>
  );
}
