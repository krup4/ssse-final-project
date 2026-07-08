import { HttpResponse, delay, http } from "msw";
import type { LoginRequest, UserRole } from "../entities/types";
import { alerts, forecastFields, history, overview, parameterErrorTrend, parameterErrors, stationSeries, stations, users, worstErrors } from "./data";

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
  http.get(endpoint("/stations"), () => HttpResponse.json(stations)),
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
