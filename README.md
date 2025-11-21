# Navidrome Import Helper

React + Go scaffold to search Amazon Music (proxy/stub), select albums or singles, confirm import, and track job progress while the backend downloads/extracts (stubbed) and places files under `NAVIDROME_MUSIC_PATH`.

## Getting Started

Prereqs: Go 1.22+, Node 18+.

### Backend

```bash
cd backend
# adjust environment or copy ../.env.example and export values
go run ./...
```

Key env vars:
- `PORT`: HTTP port (default `8080`)
- `NAVIDROME_MUSIC_PATH`: destination root for imported music (default `./navidrome_music`)
- `DATA_DIR`: where the SQLite DB lives (default `./data`)
- `TEMP_DIR`: temp download/extract area (default `./tmp`)
- `CONCURRENT_JOBS`: worker concurrency (default `2`)
- `ENABLE_DOWNLOADS`: future switch to enable real downloads (current pipeline is stubbed)
- `AMAZON_API_BASE_URL`: optional override when wiring real Amazon Music API

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Set `VITE_API_BASE` in a `.env` file (default empty uses same origin).

## Notes
- The backend job runner currently stubs doubledouble.top/pixeldrain and writes a placeholder file into the target album folder; swap in real fetch/download/extract logic where marked.
- Song selections are normalized to their parent albums on import.
- SQLite persistence is used for jobs/logs/items; tables bootstrap automatically in `DATA_DIR`.
