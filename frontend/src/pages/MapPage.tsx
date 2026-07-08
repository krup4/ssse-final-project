import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Save } from "lucide-react";
import { api } from "../shared/api/endpoints";
import { Panel } from "../shared/ui/Panel";
import { formatDateTime, formatNumber } from "../shared/lib/format";
import type { Station, StationInput } from "../entities/types";
import { StationStatusBadge } from "../shared/ui/StatusBadge";
import { useAuth } from "../features/auth/AuthContext";

const yandexMapsApiKey = import.meta.env.VITE_YANDEX_MAPS_API_KEY;
let yandexMapsPromise: Promise<YMapsApi> | null = null;

function severityColor(error: number) {
  if (error >= 16) return "#d64545";
  if (error >= 10) return "#e6a700";
  return "#1f9d68";
}

function stationColor(station: Station) {
  if (!station.isActive) return "#6f8090";
  return severityColor(station.maxError);
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

function balloonBodyHtml(station: Station) {
  return [
    `<div class="map-popup">`,
    `<span class="badge badge-${station.status === "online" ? "good" : station.status === "degraded" ? "warn" : "bad"}">${station.status}</span>`,
    `<span>Max error: ${formatNumber(station.maxError)}</span>`,
    `<span>Last telemetry: ${formatDateTime(station.lastTelemetryAt)}</span>`,
    `</div>`
  ].join("");
}

function StationRow({
  station,
  disabled,
  onSave
}: {
  station: Station;
  disabled: boolean;
  onSave: (station: Station, payload: StationInput) => void;
}) {
  const [draft, setDraft] = useState<StationInput>({
    name: station.name,
    lat: station.lat,
    lon: station.lon,
    isActive: station.isActive
  });

  useEffect(() => {
    setDraft({
      name: station.name,
      lat: station.lat,
      lon: station.lon,
      isActive: station.isActive
    });
  }, [station]);

  return (
    <tr>
      <td>
        <input
          value={draft.name}
          onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))}
        />
      </td>
      <td>
        <input
          type="number"
          min="-90"
          max="90"
          step="0.000001"
          value={draft.lat}
          onChange={(event) => setDraft((current) => ({ ...current, lat: Number(event.target.value) }))}
        />
      </td>
      <td>
        <input
          type="number"
          min="-180"
          max="180"
          step="0.000001"
          value={draft.lon}
          onChange={(event) => setDraft((current) => ({ ...current, lon: Number(event.target.value) }))}
        />
      </td>
      <td><StationStatusBadge status={station.status} /></td>
      <td>
        <input
          className="table-checkbox"
          type="checkbox"
          checked={draft.isActive}
          onChange={(event) => setDraft((current) => ({ ...current, isActive: event.target.checked }))}
        />
      </td>
      <td>
        <button type="button" className="icon-button" disabled={disabled} onClick={() => onSave(station, draft)} aria-label="Save station">
          <Save size={16} />
        </button>
      </td>
    </tr>
  );
}

export function MapPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const { data: stations = [] } = useQuery({ queryKey: ["stations"], queryFn: () => api.stations() });
  const drawableStations = useMemo(
    () => stations.filter((station) => Number.isFinite(station.lat) && Number.isFinite(station.lon)),
    [stations]
  );
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
    const markers = drawableStations.map((station) => (
      new ymaps.Placemark(
        [station.lat, station.lon],
        {
          hintContent: station.name,
          balloonContentHeader: escapeHtml(station.name),
          balloonContentBody: balloonBodyHtml(station)
        },
        {
          preset: "islands#circleDotIcon",
          iconColor: stationColor(station)
        }
      )
    ));
    const clusterer = new ymaps.Clusterer({
      clusterDisableClickZoom: false,
      clusterOpenBalloonOnClick: true,
      groupByCoordinates: true,
      clusterBalloonContentLayout: "cluster#balloonAccordion",
      clusterBalloonPanelMaxMapArea: 0,
      preset: "islands#invertedBlueClusterIcons"
    });
    clusterer.add(markers);
    map.geoObjects.add(clusterer);
  }, [drawableStations, isMapReady]);

  return (
    <div className="page-grid">
      <Panel
        title="Station map"
        className="map-panel"
        action={<span className="panel-caption">Shown {drawableStations.length} of {stations.length} stations</span>}
      >
        {yandexMapsApiKey ? (
          <div ref={mapContainerRef} className="station-map" />
        ) : (
          <div className="map-config-empty">
            <strong>Yandex Maps API key is not configured</strong>
            <span>Set VITE_YANDEX_MAPS_API_KEY in frontend environment to render the station map.</span>
          </div>
        )}
      </Panel>
      {isAdmin ? <StationManagementPanel /> : null}
    </div>
  );
}

export function StationManagementPanel() {
  const queryClient = useQueryClient();
  const { data: stations = [] } = useQuery({ queryKey: ["stations"], queryFn: () => api.stations() });
  const [newStation, setNewStation] = useState<StationInput>({ name: "", lat: 55.7558, lon: 37.6173, isActive: true });
  const createStation = useMutation({
    mutationFn: api.createStation,
    onSuccess: () => {
      setNewStation({ name: "", lat: 55.7558, lon: 37.6173, isActive: true });
      queryClient.invalidateQueries({ queryKey: ["stations"] });
    }
  });
  const updateStation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: StationInput }) => api.updateStation(id, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["stations"] })
  });

  function submitStation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    createStation.mutate(newStation);
  }

  function saveStation(station: Station, payload: StationInput) {
    updateStation.mutate({
      id: station.id,
      payload
    });
  }

  return (
    <Panel title="Station management">
      <form className="station-form" onSubmit={submitStation}>
        <label>
          Name
          <input
            value={newStation.name}
            onChange={(event) => setNewStation((current) => ({ ...current, name: event.target.value }))}
            required
          />
        </label>
        <label>
          Latitude
          <input
            type="number"
            min="-90"
            max="90"
            step="0.000001"
            value={newStation.lat}
            onChange={(event) => setNewStation((current) => ({ ...current, lat: Number(event.target.value) }))}
            required
          />
        </label>
        <label>
          Longitude
          <input
            type="number"
            min="-180"
            max="180"
            step="0.000001"
            value={newStation.lon}
            onChange={(event) => setNewStation((current) => ({ ...current, lon: Number(event.target.value) }))}
            required
          />
        </label>
        <label className="checkbox-label">
          <input
            type="checkbox"
            checked={newStation.isActive}
            onChange={(event) => setNewStation((current) => ({ ...current, isActive: event.target.checked }))}
          />
          Active
        </label>
        <button type="submit" className="primary-button station-submit" disabled={createStation.isPending}>
          <Plus size={16} />
          Add station
        </button>
      </form>

      <div className="table-wrap compact-table">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Lat</th>
              <th>Lon</th>
              <th>Status</th>
              <th>Active</th>
              <th>Save</th>
            </tr>
          </thead>
          <tbody>
            {stations.map((station) => (
              <StationRow key={station.id} station={station} disabled={updateStation.isPending} onSave={saveStation} />
            ))}
          </tbody>
        </table>
      </div>
    </Panel>
  );
}
