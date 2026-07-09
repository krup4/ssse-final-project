import { Outlet } from "react-router-dom";
import { AuthProvider } from "../features/auth/AuthContext";
import { FiltersProvider } from "../features/filters/FiltersContext";

export function ProvidersLayout() {
  return (
    <AuthProvider>
      <FiltersProvider>
        <Outlet />
      </FiltersProvider>
    </AuthProvider>
  );
}
