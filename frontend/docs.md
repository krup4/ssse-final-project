# ssse-final-project

Frontend for a weather forecast accuracy analytics system.

## Local frontend

```bash
cd frontend
npm install
cat > public/config.js <<'EOF'
window.__APP_CONFIG__ = {
  yandexMapsApiKey: "your-yandex-maps-api-key"
};
EOF
npm run dev
```

Demo accounts:

- `admin`
- `analyst`
- `operator`
- `viewer`

Password is `password` for the real local Core API. In mock mode, any
non-empty value is accepted.

## Docker

Create `.env` in the repository root:

```bash
cp .env.example .env
```

Then set:

```dotenv
YANDEX_MAPS_API_KEY=your-yandex-maps-api-key
```

Build and run:

```bash
docker compose up --build
```

Frontend will be available at `http://localhost:3000`; API calls to `/api/v1`
are proxied to `core-api`.

Optional local-only demo data and Kafka telemetry scripts are documented in
`dev/README.md`. Do not run those scripts for production deploys.

The station map uses Yandex Maps JavaScript API. The production image writes
`/config.js` at container startup from `YANDEX_MAPS_API_KEY`, so the same image
can be reused across environments. In Kubernetes, provide this value through
`frontend-secret`.
