Navidrome Import Helper - Library Sync Tasks
============================================

Status Legend: [ ] TODO  [-] IN PROGRESS  [x] DONE

Backend
- [x] Add SQLite table for library index (artist/album/path/track_count/updated_at) with normalized keys.
- [x] Implement filesystem scanner for `NAVIDROME_MUSIC_PATH` (two-level Artist/Album, audio extensions) that refreshes the index.
- [x] Expose `GET /api/library` (cached) and `POST /api/library/refresh` (rescan + return updated entries).
- [x] Enrich `/api/search` results with `exists` flags using normalized artist/album matching; song → parent album.
- [x] Add minimal logging/messages around refresh and matching.

Frontend
- [x] Use enriched `exists` flag in search results to show “Already in library” badges / disable selection.
- [x] Add library view (consume `/api/library`) showing indexed albums and a refresh trigger.

Docs/Config
- [x] Document the new endpoints and scanner behavior in README or PRD notes.
