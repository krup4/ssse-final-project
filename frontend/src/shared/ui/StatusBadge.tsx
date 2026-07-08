import type { AlertSeverity, AlertStatus, StationStatus, UserRole } from "../../entities/types";

type BadgeTone = "good" | "warn" | "bad" | "neutral" | "accent";

const stationTone: Record<StationStatus, BadgeTone> = {
  online: "good",
  degraded: "warn",
  offline: "bad"
};

const severityTone: Record<AlertSeverity, BadgeTone> = {
  critical: "bad",
  warning: "warn",
  info: "accent"
};

const statusTone: Record<AlertStatus, BadgeTone> = {
  open: "bad",
  acknowledged: "warn",
  resolved: "good"
};

const roleTone: Record<UserRole, BadgeTone> = {
  admin: "bad",
  analyst: "accent",
  operator: "warn",
  viewer: "neutral"
};

export function StatusBadge({
  label,
  tone
}: {
  label: string;
  tone: BadgeTone;
}) {
  return <span className={`badge badge-${tone}`}>{label}</span>;
}

export function StationStatusBadge({ status }: { status: StationStatus }) {
  return <StatusBadge label={status} tone={stationTone[status]} />;
}

export function AlertSeverityBadge({ severity }: { severity: AlertSeverity }) {
  return <StatusBadge label={severity} tone={severityTone[severity]} />;
}

export function AlertStatusBadge({ status }: { status: AlertStatus }) {
  return <StatusBadge label={status} tone={statusTone[status]} />;
}

export function RoleBadge({ role }: { role: UserRole }) {
  return <StatusBadge label={role} tone={roleTone[role]} />;
}
