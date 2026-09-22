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
go run main.go
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

# nombresmad
nombres en mdera
