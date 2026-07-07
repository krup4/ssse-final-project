import { createContext, useContext, useMemo, useState } from "react";
import type { AnalyticsFilters } from "../../entities/types";

const defaultFilters: AnalyticsFilters = {
  dateFrom: "2026-07-01",
  dateTo: "2026-07-07",
  regionId: "all",
  stationId: "all",
  metric: "wind_speed"
};

interface FiltersContextValue {
  filters: AnalyticsFilters;
  setFilters: React.Dispatch<React.SetStateAction<AnalyticsFilters>>;
}

const FiltersContext = createContext<FiltersContextValue | null>(null);

export function FiltersProvider({ children }: { children: React.ReactNode }) {
  const [filters, setFilters] = useState<AnalyticsFilters>(defaultFilters);
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
