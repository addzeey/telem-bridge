# Telemetry Bridge

A cross-platform bridge for forwarding game telemetry data via UDP and OSC, with a modern React dashboard for configuration and live monitoring.

## Features

- Real-time telemetry forwarding via UDP and OSC
- Web dashboard for configuration, live data, and service control
- Packet forwarding and OSC address mapping
- Service restart endpoints for hot-reloading without full restart

## Project Structure

```
.
├── backend/      # Go backend (API, UDP/OSC, static file server)
│   ├── main.go
│   ├── config.go
│   ├── ...
│   └── dist/     # Built frontend assets (output from frontend build)
├── frontend/     # React + TypeScript + Vite frontend
│   ├── src/
│   ├── index.html
│   └── ...
├── package.json
└── README.md
```

## Prerequisites

- [Go](https://golang.org/dl/) 1.20+ (for backend)
- [Node.js](https://nodejs.org/) 18+ and [npm](https://www.npmjs.com/) (for frontend)

---

## Building the Application

1. **Delete old frontend build (recommended before rebuilding):**

   ```sh
   rmdir /s /q backend\dist
   ```

2. **Build the frontend (only if changes were made):**

   ```sh
   cd frontend
   npm install
   npm run build
   cd ..
   ```
   This outputs static files to `backend/dist`.

3. **Change the version:**

   - Update the `Version` variable in `backend/main.go` as needed.
   - Optionally update the version in `frontend/package.json`.

4. **Build the backend application:**

   ```sh
   cd backend
   go build -o f1-telem-bridge.exe
   cd ..
   ```

   The resulting binary will be `backend/f1-telem-bridge.exe`.

---

## Running the Application

### Option 1: Run the built binary (recommended for production)

```sh
backend\f1-telem-bridge.exe
```

- The dashboard will open automatically at [http://localhost:1337](http://localhost:1337).

### Option 2: Run backend and frontend in development mode (hot reload frontend)

1. **Start the frontend dev server:**

   ```sh
   cd frontend
   npm run dev
   ```

2. **Run the backend (serves API only):**

   ```sh
   cd backend
   go run .
   ```

- Visit the frontend at the URL shown by Vite (usually [http://localhost:5173](http://localhost:5173)).
- The backend API will be available at [http://localhost:1337](http://localhost:1337).

---

## Service Restart Endpoints

You can restart UDP, OSC, or all services via the dashboard or by calling:

- `POST /api/restart/udp`
- `POST /api/restart/osc`
- `POST /api/restart/all`

---

## Configuration & Logs

- Config files and logs are stored in your user config directory (e.g. `%APPDATA%\f1-telem-bridge` on Windows).
- Logs are written to `app.log`.

---

## License

MIT © 2025 Adam Ashdown