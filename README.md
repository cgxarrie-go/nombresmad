# nombresmad

This repository contains a Go backend and React frontend for the NombresMad app.

Quick start (requires Docker/Postgres):

1. Create Postgres DB `nombresmad` or use Docker Compose below.
2. Start services with Docker Compose:

```bash
docker-compose up --build
```

3. Or run locally:

From `backend` folder:

```bash
go run ./cmd
```

Then in another terminal from `backend` run:

```bash
./import.sh
```

From `frontend` folder:

```bash
npm install
npm start
```

If using Docker Compose the frontend is available at http://localhost:3000 and backend at http://localhost:8080

Frontend backend configuration:

- `REACT_APP_API_URL` sets the browser-facing API base URL. It defaults to `/api`, which keeps local requests on the frontend origin.
- `API_PROXY_TARGET` sets the Create React App development proxy target. It defaults to `http://localhost:8080`; Docker Compose sets it to `http://backend:8080`.

Railway deployment:

- Set `API_PROXY_TARGET` on the frontend service to `http://<backend-private-domain>:<backend-port>`. Use the backend's Railway private domain and its assigned `PORT`; do not append `/api`.
- The production frontend proxies `/api` through this variable, so the backend can remain private and session cookies stay same-origin.
- The backend listens on Railway's `PORT` and defaults to `8080` for local development and Docker Compose.

# nombresmad
nombres en mdera
