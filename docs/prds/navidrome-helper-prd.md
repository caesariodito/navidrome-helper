Navidrome Import Helper PRD
===========================

Overview
--------
- Build a React + Go service that searches Amazon Music for albums/songs, lets a user select items to import, and orchestrates downloading, extracting, and placing audio files into `NAVIDROME_MUSIC_PATH` under artist-organized folders.
- Out of scope for now: full Navidrome integration APIs (we only place files), user auth/roles, and advanced library management (playlists, metadata editing).

Goals
-----
- Simple UI to search Amazon Music content and preview key details (art, artist, album, track count).
- Allow selecting albums/songs, confirming the import, and showing live progress per job.
- Backend reliably fetches pixeldrain links via doubledouble.top, downloads zips, extracts, and organizes files by artist path.
- Track job state so the frontend can poll/stream loading + completion/errors.

Users
-----
- Navidrome admins or power users who want a fast import flow without manual downloads or file moves.

User Stories
------------
- As a user, I search for an album/song by keywords and see relevant Amazon Music results.
- As a user, I select one or more items and get a confirmation dialog before import starts.
- As a user, if I pick a song/single, the system imports the containing album and makes that clear before I confirm.
- As a user, I see progress (queued → fetching → downloading → extracting → placing → done/error) for each import job.
- As a user, I can view job outcomes (successes, failures, and messages) and retry failed imports.

Assumptions
-----------
- Valid Amazon Music API access/credentials are available (rate limits known).
- doubledouble.top remains reachable and produces pixeldrain URLs given the target album.
- pixeldrain download is stable and allowed for automated access.
- The server has write access to `NAVIDROME_MUSIC_PATH` and enough disk for temp zips + extraction.
- No user authentication initially (single trusted user).

Functional Requirements
-----------------------
- Search:
  - Input box with debounced search; list results with title, artist, cover, type (album/single/song), track count/duration.
  - Show both album entries and song entries; when a song is selected, the import resolves to its parent album.
  - Backend endpoint wraps Amazon Music API (avoid CORS secrets in frontend).
- Selection + Confirmation:
  - Multi-select albums/songs; confirmation dialog shows count, indicates album-level import if a song was picked, and shows target path root.
  - Allow cancel; on confirm, POST starts an import job.
- Import Job Handling:
  - Backend creates an async job with ID and initial status.
  - Status includes phase, percent (if estimable), and log/error messages.
  - Frontend polls or subscribes to status to show progress and completion.
- Download & File Placement:
  - Fetch pixeldrain URL via doubledouble.top for each selected item.
  - Download zip to temp dir, extract, validate audio files, and move under `NAVIDROME_MUSIC_PATH/<Artist>/<Album>/`.
  - Ensure safe paths (sanitize names), avoid overwriting by default; if conflict, skip or create unique path and record in job log.
- History & Observability:
  - List recent jobs with status, timestamps, counts, and error detail.
  - Basic server logs per job for debugging.

Non-Functional Requirements
---------------------------
- Reliability: failed steps surface clear errors; temp files cleaned.
- Performance: do not block server while waiting on downloads; allow concurrent jobs (configurable limit).
- Security: keep API keys server-side; sanitize filenames to prevent path traversal.
- Configurability: `NAVIDROME_MUSIC_PATH`, temp dir, concurrency limits, timeouts.
- Testability: unit tests for job state machine/path handling; integration test for job lifecycle with mocked external calls.

Data Model (initial)
--------------------
- Job: id, type (album/song), source ids, status, phase, progress, message, created_at, updated_at, finished_at.
- JobItem (optional): job_id, item_id, artist, album, status, message.
- Stored locally (e.g., SQLite or filesystem JSON) to survive restarts; in-memory cache acceptable for first iteration if persistence is optional.

API Endpoints (proposed)
------------------------
- `GET /api/search?q=` → Amazon Music search results (server-side call).
- `POST /api/import` → body: selected item ids + metadata; returns job_id.
- `GET /api/jobs/:id` → job status/progress.
- `GET /api/jobs` → recent jobs summary.
- (Optional) `POST /api/jobs/:id/retry` to rerun failed jobs.

System/Workflow
---------------
1) Frontend search calls backend search endpoint.
2) User selects items → confirmation dialog → POST /api/import.
3) Backend enqueues job; worker processes phases: fetch pixeldrain → download → extract → place files → cleanup.
4) Frontend polls/streams job status; shows progress and final state.

Decisions From Review
---------------------
- Amazon Music API/SDK: any free-accessible option is acceptable; expect server-side token/credentials.
- Import scope: whole albums only (no per-track selection).
- Overwrite/conflicts: skip existing files/folders rather than overwrite or version.
- Metadata handling: trust zip contents from doubledouble.top (no normalization).
- Persistence: SQLite for job storage.

Success Metrics
---------------
- Import job success rate and average completion time.
- Time saved vs. manual download/move flow.
- User-visible errors per import (should be minimal and actionable).
