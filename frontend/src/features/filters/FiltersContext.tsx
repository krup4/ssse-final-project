import { createContext, useContext, useMemo, useState } from "react";
import type { AnalyticsFilters } from "../../entities/types";

function dateInputValue(date: Date) {
  return date.toISOString().slice(0, 10);
}

export function defaultFilters(): AnalyticsFilters {
  const dateTo = new Date();
  const dateFrom = new Date(dateTo);
  dateFrom.setUTCDate(dateTo.getUTCDate() - 180);
  return {
    dateFrom: dateInputValue(dateFrom),
    dateTo: dateInputValue(dateTo),
    stationId: "all",
    field: "all",
    metric: "all"
  };
}

interface FiltersContextValue {
  filters: AnalyticsFilters;
  setFilters: React.Dispatch<React.SetStateAction<AnalyticsFilters>>;
}

const FiltersContext = createContext<FiltersContextValue | null>(null);

export function FiltersProvider({ children }: { children: React.ReactNode }) {
  const [filters, setFilters] = useState<AnalyticsFilters>(() => defaultFilters());
  const value = useMemo(() => ({ filters, setFilters }), [filters]);
  return <FiltersContext.Provider value={value}>{children}</FiltersContext.Provider>;
}

export function useFilters() {
  const context = useContext(FiltersContext);
  if (!context) {
    throw new Error("useFilters must be used inside FiltersProvider");
  }
  return context;
}
