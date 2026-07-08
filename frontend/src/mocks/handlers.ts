import { HttpResponse, delay, http } from "msw";
import type { LoginRequest, StationInput, UserRole } from "../entities/types";
import { alerts, forecastFields, history, metrics, overview, parameterErrorTrend, parameterErrors, stationSeries, stations, users, worstErrors } from "./data";

const apiBase = import.meta.env.VITE_API_BASE_URL || "";

function endpoint(path: string) {
  return `${apiBase}${path}`;
}

function tokenFor(userId: string) {
  return `mock-jwt:${userId}:${Date.now()}`;
}

function userFromAuthHeader(request: Request) {
  const auth = request.headers.get("authorization") || "";
  const userId = auth.replace("Bearer mock-jwt:", "").split(":")[0];
  return users.find((user) => user.id === userId) ?? users[1];
}

export const handlers = [
  http.post(endpoint("/auth/login"), async ({ request }) => {
    await delay(250);
    const body = (await request.json()) as LoginRequest;
    const roleFromLogin = body.login as UserRole;
    const user = users.find((item) => item.login === body.login) ?? users.find((item) => item.role === roleFromLogin) ?? users[1];
    return HttpResponse.json({ token: tokenFor(user.id), user });
  }),
  http.get(endpoint("/auth/me"), ({ request }) => HttpResponse.json(userFromAuthHeader(request))),
  http.get(endpoint("/forecast-fields"), () => HttpResponse.json(forecastFields)),
  http.get(endpoint("/metrics/catalog"), () => HttpResponse.json(metrics)),
  http.get(endpoint("/stations"), ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    if (status === "online") {
      return HttpResponse.json(stations.filter((station) => station.isActive));
    }
    if (status === "offline") {
      return HttpResponse.json(stations.filter((station) => !station.isActive));
    }
    return HttpResponse.json(stations);
  }),
  http.post(endpoint("/stations"), async ({ request }) => {
    const body = (await request.json()) as StationInput;
    const station = {
      id: String(Date.now()),
      name: body.name,
      lat: body.lat,
      lon: body.lon,
      isActive: body.isActive,
      status: body.isActive ? "online" as const : "offline" as const,
      activeSensors: body.isActive ? 1 : 0,
      maxError: 0,
      lastTelemetryAt: new Date().toISOString()
    };
    stations.push(station);
    return HttpResponse.json(station, { status: 201 });
  }),
  http.patch(endpoint("/stations/:id"), async ({ params, request }) => {
    const body = (await request.json()) as StationInput;
    const station = stations.find((item) => item.id === params.id);
    if (!station) {
      return new HttpResponse(null, { status: 404 });
    }
    station.name = body.name;
    station.lat = body.lat;
    station.lon = body.lon;
    station.isActive = body.isActive;
    station.status = body.isActive ? "online" : "offline";
    station.activeSensors = body.isActive ? 1 : 0;
    return HttpResponse.json(station);
  }),
  http.get(endpoint("/metrics/overview"), () => HttpResponse.json(overview)),
  http.get(endpoint("/analytics/worst-errors"), () => HttpResponse.json(worstErrors)),
  http.get(endpoint("/analytics/parameter-errors"), ({ request }) => {
    const url = new URL(request.url);
    const parameter = url.searchParams.get("parameter");
    const result = parameter && parameter !== "all" ? parameterErrors.filter((row) => row.parameter === parameter) : parameterErrors;
    return HttpResponse.json(result);
  }),
  http.get(endpoint("/analytics/parameter-error-trend"), ({ request }) => {
    const url = new URL(request.url);
    const parameter = url.searchParams.get("parameter");
    const result = parameter && parameter !== "all" ? parameterErrorTrend.filter((row) => row.parameter === parameter) : parameterErrorTrend;
    return HttpResponse.json(result);
  }),
  http.get(endpoint("/analytics/station-series"), () => HttpResponse.json(stationSeries)),
  http.get(endpoint("/analytics/history"), () => HttpResponse.json(history)),
  http.get(endpoint("/alerts"), () => HttpResponse.json(alerts)),
  http.get(endpoint("/users"), () => HttpResponse.json(users)),
  http.patch(endpoint("/users/:id/role"), async ({ params, request }) => {
    const body = (await request.json()) as { role: UserRole };
    const user = users.find((item) => item.id === params.id);
    if (!user) {
      return new HttpResponse(null, { status: 404 });
    }
    user.role = body.role;
    return HttpResponse.json(user);
  })
];
