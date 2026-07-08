# ssse-final-project

Frontend for a weather forecast accuracy analytics system.

## Local frontend

```bash
cd frontend
npm install
export VITE_YANDEX_MAPS_API_KEY=your-yandex-maps-api-key
npm run dev
```

Demo accounts:

- `admin@weather.local`
- `analyst@weather.local`
- `operator@weather.local`
- `viewer@weather.local`

Password can be any non-empty value in mock mode.

## Docker

Create `.env` in the repository root:

```bash
cp .env.example .env
```

Then set:

```dotenv
VITE_YANDEX_MAPS_API_KEY=your-yandex-maps-api-key
```

Build and run:

```bash
docker compose up --build
```

Frontend will be available at `http://localhost:3000`.

The station map uses Yandex Maps JavaScript API. Because this is a Vite static
frontend, `VITE_*` values are embedded during Docker image build. Rebuild the
image after changing `.env`.
