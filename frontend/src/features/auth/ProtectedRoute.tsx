import { Navigate, Outlet } from "react-router-dom";
import type { UserRole } from "../../entities/types";
import { useAuth } from "./AuthContext";

export function ProtectedRoute({ roles }: { roles?: UserRole[] }) {
  const { token, hasRole } = useAuth();
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  if (roles && !hasRole(roles)) {
    return <Navigate to="/" replace />;
  }
  return <Outlet />;
}
