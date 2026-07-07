import {
  Activity,
  AlertTriangle,
  BarChart3,
  History,
  LayoutDashboard,
  LogOut,
  Map,
  ListFilter,
  Shield,
  Target
} from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";
import type { UserRole } from "../entities/types";
import { GlobalFilters } from "../features/filters/GlobalFilters";
import { useAuth } from "../features/auth/AuthContext";
import { RoleBadge } from "../shared/ui/StatusBadge";

const navItems: Array<{ path: string; label: string; icon: typeof LayoutDashboard; roles: UserRole[] }> = [
  { path: "/", label: "Overview", icon: LayoutDashboard, roles: ["admin", "analyst", "operator", "viewer"] },
  { path: "/map", label: "Station map", icon: Map, roles: ["admin", "analyst", "operator", "viewer"] },
  { path: "/errors", label: "Top errors", icon: Target, roles: ["admin", "analyst", "viewer"] },
  { path: "/parameter-errors", label: "Parameter errors", icon: ListFilter, roles: ["admin", "analyst", "viewer"] },
  { path: "/charts", label: "Station charts", icon: BarChart3, roles: ["admin", "analyst", "viewer"] },
  { path: "/history", label: "History", icon: History, roles: ["admin", "analyst"] },
  { path: "/alerts", label: "Alerts", icon: AlertTriangle, roles: ["admin", "operator"] },
  { path: "/users", label: "Access", icon: Shield, roles: ["admin"] }
];

export function AppShell() {
  const { user, logout, hasRole } = useAuth();
  const visibleItems = navItems.filter((item) => hasRole(item.roles));

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">
            <Activity size={22} />
          </div>
          <div>
            <strong>Weather Accuracy</strong>
            <span>Forecast error analytics</span>
          </div>
        </div>
        <nav className="nav-list">
          {visibleItems.map((item) => (
            <NavLink key={item.path} to={item.path} end={item.path === "/"}>
              <item.icon size={18} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-user">
          <div>
            <strong>{user?.name}</strong>
            {user ? <RoleBadge role={user.role} /> : null}
          </div>
          <button type="button" className="icon-button" onClick={logout} aria-label="Logout" title="Logout">
            <LogOut size={18} />
          </button>
        </div>
      </aside>
      <main className="main">
        <header className="topbar">
          <div>
            <p>Presentation Layer</p>
            <h1>Forecast accuracy control center</h1>
          </div>
          <GlobalFilters />
        </header>
        <Outlet />
      </main>
    </div>
  );
}
