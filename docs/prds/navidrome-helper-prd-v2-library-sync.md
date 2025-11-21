Navidrome Import Helper PRD (Library Sync)
==========================================

Overview
--------
- Extend the existing React + Go import helper so the frontend knows which albums already exist in `NAVIDROME_MUSIC_PATH` and can avoid re-imports.
- Provide API endpoints to list/refresh the Navidrome library index and enrich search results with `exists` flags.

Goals
-----
- Maintain a lightweight library index (artist/album/track count/path) refreshed from `NAVIDROME_MUSIC_PATH`.
- Expose library entries and on-demand refresh via API for the frontend.
- Enrich Amazon Music search results with an `exists` boolean (album-aware, song → parent album).

Functional Requirements
-----------------------
- Library Index:
  - Scan `NAVIDROME_MUSIC_PATH` for `Artist/Album` directories; count audio tracks (common extensions).
  - Store normalized artist/album keys and track count in SQLite; refreshable on demand.
  - Skip hidden folders and non-audio files; tolerate missing metadata.
- API:
  - `GET /api/library` -> returns cached library entries (artist, album, trackCount, path, updatedAt).
  - `POST /api/library/refresh` -> triggers a rescan + cache update; returns updated entries.
  - `GET /api/search` -> same as before but returns `exists` per item by matching normalized artist/album (song normalized to album).
- Matching:
  - Normalize artist+album to lower-case, trimmed, punctuation-light for matching.
  - Song searches map to parent album for `exists` logic.

Non-Functional Requirements
---------------------------
- Scans should be bounded (ignore hidden paths); keep in-memory operations simple.
- Do not block search; use cached index and let `/api/library/refresh` refresh when desired.
- Keep all file system paths sanitized and avoid traversal.

Out of Scope
------------
- Deep metadata/tag parsing.
- Navidrome API calls; we only scan filesystem paths.
- Partial track-level diffing.

Open Questions
--------------
- Refresh trigger cadence: manual only vs. timed background refresh?
- Should we include file sizes/durations in the index?
