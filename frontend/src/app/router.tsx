import { createBrowserRouter } from "react-router-dom";
import { AppShell } from "./AppShell";
import { ProvidersLayout } from "./ProvidersLayout";
import { ProtectedRoute } from "../features/auth/ProtectedRoute";
import { AlertsPage } from "../pages/AlertsPage";
import { ChartsPage } from "../pages/ChartsPage";
import { DashboardPage } from "../pages/DashboardPage";
import { ErrorsPage } from "../pages/ErrorsPage";
import { HistoryPage } from "../pages/HistoryPage";
import { LoginPage } from "../pages/LoginPage";
import { ParameterErrorsPage } from "../pages/ParameterErrorsPage";
import { StationsPage } from "../pages/StationsPage";
import { UsersPage } from "../pages/UsersPage";

export const router = createBrowserRouter([
  {
    element: <ProvidersLayout />,
    children: [
      {
        path: "/login",
        element: <LoginPage />
      },
      {
        element: <ProtectedRoute />,
        children: [
          {
            element: <AppShell />,
            children: [
              { path: "/", element: <DashboardPage /> },
              { path: "/map", lazy: () => import("../pages/MapPage").then((module) => ({ Component: module.MapPage })) },
              {
                element: <ProtectedRoute roles={["admin", "analyst", "viewer"]} />,
                children: [
                  { path: "/errors", element: <ErrorsPage /> },
                  { path: "/parameter-errors", element: <ParameterErrorsPage /> },
                  { path: "/charts", element: <ChartsPage /> }
                ]
              },
              {
                element: <ProtectedRoute roles={["admin", "analyst"]} />,
                children: [{ path: "/history", element: <HistoryPage /> }]
              },
              {
                element: <ProtectedRoute roles={["admin", "operator"]} />,
                children: [{ path: "/alerts", element: <AlertsPage /> }]
              },
              {
                element: <ProtectedRoute roles={["admin"]} />,
                children: [
                  { path: "/stations", element: <StationsPage /> },
                  { path: "/users", element: <UsersPage /> }
                ]
              }
            ]
          }
        ]
      }
    ]
  }
]);
