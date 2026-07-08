import { useQuery } from "@tanstack/react-query";
import { api } from "../shared/api/endpoints";
import { useFilters } from "../features/filters/FiltersContext";
import { formatNumber } from "../shared/lib/format";
import { Panel } from "../shared/ui/Panel";

export function HistoryPage() {
  const { filters } = useFilters();
  const { data = [] } = useQuery({ queryKey: ["history", filters], queryFn: () => api.history(filters) });

  return (
    <div className="page-grid">
      <Panel title="Historical analytics">
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Station</th>
                <th>MAE</th>
                <th>RMSE</th>
                <th>Samples</th>
                <th>Backfill</th>
              </tr>
            </thead>
            <tbody>
              {data.map((row) => (
                <tr key={`${row.date}-${row.stationName}`}>
                  <td>{row.date}</td>
                  <td>{row.stationName}</td>
                  <td>{formatNumber(row.mae)}</td>
                  <td>{formatNumber(row.rmse)}</td>
                  <td>{row.samples.toLocaleString("en")}</td>
                  <td>{row.backfillVersion}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  );
}
