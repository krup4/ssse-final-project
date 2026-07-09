import React from "react";
import ReactDOM from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "react-router-dom";
import { router } from "./app/router";
import "./styles.css";

async function enableMocks() {
  if (import.meta.env.VITE_ENABLE_MOCKS !== "true") {
    return;
  }
  const { worker } = await import("./mocks/browser");
  await worker.start({ onUnhandledRequest: "bypass" });
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      refetchOnWindowFocus: false
    }
  }
});

enableMocks().then(() => {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider
          router={router}
          fallbackElement={<div className="route-loading">Loading...</div>}
        />
      </QueryClientProvider>
    </React.StrictMode>
  );
});
