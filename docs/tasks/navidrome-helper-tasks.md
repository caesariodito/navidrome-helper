Navidrome Import Helper - Tasks
===============================

Status Legend: [ ] TODO  [-] IN PROGRESS  [x] DONE

Planning & Foundations
- [x] Confirm Amazon Music API access method/keys and regional support. (Use any free-accessible option; handle server-side credentials)
- [x] Decide persistence for jobs (choose SQLite).
- [x] Define overwrite/duplicate handling policy for destination folders. (Skip conflicts)
- [ ] Set config surface: `NAVIDROME_MUSIC_PATH`, temp dir, concurrency, timeouts, log level.

Backend (Go)
- [ ] Scaffold Go service with config loading and logging (env-based).
- [ ] Implement job model/state machine (phases, timestamps, messages, progress).
- [ ] Add persistence layer (chosen option) for jobs + optional job items.
- [ ] Implement Amazon Music search proxy endpoint (`GET /api/search?q`), with rate-limit handling (free API compatible).
- [ ] Implement import creation endpoint (`POST /api/import`) that validates selection, normalizes song selections to parent albums, creates job(s), enqueues worker task.
- [ ] Implement job status endpoints (`GET /api/jobs`, `GET /api/jobs/:id`); include recent logs and progress fields.
- [ ] Build worker pipeline: fetch pixeldrain URL via doubledouble.top; download zip; extract; sanitize paths; move to `NAVIDROME_MUSIC_PATH/<Artist>/<Album>/`; handle conflicts per policy; cleanup temp files.
- [ ] Add retry mechanism for failed jobs (optional endpoint or auto-retry limit).
- [ ] Add structured logging per job and error propagation to status messages.
- [ ] Add unit tests for path sanitization, job transitions, and error handling; integration test for job lifecycle with mocks.

Frontend (React)
- [ ] Bootstrap React app structure (routing, state mgmt, design tokens).
- [ ] Build search UI with debounced input and results list (title, artist, cover, type, track count), showing albums and singles/songs.
- [ ] Implement selection model (multi-select albums/songs) with clear indicators; if a song is selected, display that the parent album will be imported.
- [ ] Add confirmation dialog summarizing selections and target path root.
- [ ] Implement POST to start import and capture job_id.
- [ ] Build job status view (polling or SSE) showing phases, progress, and errors; allow retry trigger if available.
- [ ] Add recent jobs/history list for quick visibility.
- [ ] Basic styling consistent with PRD scope (no auth for now).

Integration & Ops
- [ ] Wire frontend to backend endpoints; centralize API client with error handling.
- [ ] Configure temp storage and ensure write access to `NAVIDROME_MUSIC_PATH`.
- [ ] Add `.env` example documenting required variables and secrets handling.
- [ ] Provide scripts/Makefile targets for dev (run frontend + backend), tests, and linting.
- [ ] Document runbook: how to start services, where files land, how to retry/clean temp.

Validation
- [ ] Manual E2E test: search → select → confirm → import → files appear in Navidrome path.
- [ ] Record known limitations and follow-ups after first run.
