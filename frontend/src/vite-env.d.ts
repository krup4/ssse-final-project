/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_ENABLE_MOCKS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare global {
  interface Window {
    __APP_CONFIG__?: {
      yandexMapsApiKey?: string;
    };
    ymaps?: YMapsApi;
  }

  interface YMapsApi {
    ready: (callback: () => void) => void;
    Map: new (
      element: HTMLElement,
      state: { center: [number, number]; zoom: number; controls?: string[] },
      options?: Record<string, unknown>
    ) => YandexMap;
    Clusterer: new (options?: Record<string, unknown>) => YandexClusterer;
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

  interface YandexClusterer extends YandexGeoObject {
    add: (objects: YandexGeoObject[]) => void;
  }
}

export {};
