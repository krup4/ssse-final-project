/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_ENABLE_MOCKS?: string;
  readonly VITE_YANDEX_MAPS_API_KEY?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare global {
  interface Window {
    ymaps?: YMapsApi;
  }

  interface YMapsApi {
    ready: (callback: () => void) => void;
    Map: new (
      element: HTMLElement,
      state: { center: [number, number]; zoom: number; controls?: string[] },
      options?: Record<string, unknown>
    ) => YandexMap;
    Placemark: new (
      coordinates: [number, number],
      properties: Record<string, unknown>,
      options?: Record<string, unknown>
    ) => YandexGeoObject;
  }

  interface YandexMap {
    geoObjects: {
      add: (object: YandexGeoObject) => void;
      removeAll: () => void;
    };
    destroy: () => void;
  }

  interface YandexGeoObject {}
}

export {};
