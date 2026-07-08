import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../shared/api/endpoints";
import { Panel } from "../shared/ui/Panel";
import { formatDateTime } from "../shared/lib/format";
import type { Station } from "../entities/types";

const yandexMapsApiKey = import.meta.env.VITE_YANDEX_MAPS_API_KEY;
let yandexMapsPromise: Promise<YMapsApi> | null = null;

function severityColor(error: number) {
  if (error >= 16) return "#d64545";
  if (error >= 10) return "#e6a700";
  return "#1f9d68";
}

function escapeHtml(value: string) {
  return value.replace(/[&<>"']/g, (char) => {
    const entities: Record<string, string> = {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;"
    };
    return entities[char];
  });
}

function loadYandexMaps() {
  if (window.ymaps) {
    return Promise.resolve(window.ymaps);
  }

  if (!yandexMapsApiKey) {
    return Promise.reject(new Error("VITE_YANDEX_MAPS_API_KEY is not configured"));
  }

  if (!yandexMapsPromise) {
    yandexMapsPromise = new Promise((resolve, reject) => {
      const script = document.createElement("script");
      script.src = `https://api-maps.yandex.ru/2.1/?apikey=${encodeURIComponent(yandexMapsApiKey)}&lang=ru_RU`;
      script.async = true;
      script.onload = () => {
        if (!window.ymaps) {
          reject(new Error("Yandex Maps API did not initialize"));
          return;
        }
        window.ymaps.ready(() => resolve(window.ymaps!));
      };
      script.onerror = () => reject(new Error("Failed to load Yandex Maps API"));
      document.head.appendChild(script);
    });
  }

  return yandexMapsPromise;
}

function balloonHtml(station: Station) {
  return [
    `<div class="map-popup">`,
    `<strong>${escapeHtml(station.name)}</strong>`,
    `<span class="badge badge-${station.status === "online" ? "good" : station.status === "degraded" ? "warn" : "bad"}">${station.status}</span>`,
    `<span>Max error: ${station.maxError}</span>`,
    `<span>Last telemetry: ${formatDateTime(station.lastTelemetryAt)}</span>`,
    `</div>`
  ].join("");
}

export function MapPage() {
  const { data: stations = [] } = useQuery({ queryKey: ["stations"], queryFn: api.stations });
  const mapContainerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<YandexMap | null>(null);
  const ymapsRef = useRef<YMapsApi | null>(null);
  const [isMapReady, setIsMapReady] = useState(false);

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) {
      return;
    }

    let isMounted = true;

    loadYandexMaps()
      .then((ymaps) => {
        if (!isMounted || !mapContainerRef.current) {
          return;
        }
        ymapsRef.current = ymaps;
        mapRef.current = new ymaps.Map(
          mapContainerRef.current,
          {
            center: [55.75, 61.0],
            zoom: 4,
            controls: ["zoomControl", "fullscreenControl"]
          },
          {
            suppressMapOpenBlock: true,
            yandexMapDisablePoiInteractivity: true
          }
        );
        setIsMapReady(true);
      })
      .catch(() => {
        mapRef.current = null;
      });

    return () => {
      isMounted = false;
      mapRef.current?.destroy();
      mapRef.current = null;
      setIsMapReady(false);
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    const ymaps = ymapsRef.current;
    if (!map || !ymaps || !isMapReady) {
      return;
    }

    map.geoObjects.removeAll();
    stations.forEach((station) => {
      const marker = new ymaps.Placemark(
        [station.lat, station.lon],
        {
          hintContent: station.name,
          balloonContent: balloonHtml(station)
        },
        {
          preset: "islands#circleDotIcon",
          iconColor: severityColor(station.maxError)
        }
      );
      map.geoObjects.add(marker);
    });
  }, [isMapReady, stations]);

  return (
    <div className="page-grid">
      <Panel title="Station map" className="map-panel">
        {yandexMapsApiKey ? (
          <div ref={mapContainerRef} className="station-map" />
        ) : (
          <div className="map-config-empty">
            <strong>Yandex Maps API key is not configured</strong>
            <span>Set VITE_YANDEX_MAPS_API_KEY in frontend environment to render the station map.</span>
          </div>
        )}
      </Panel>
    </div>
  );
}
