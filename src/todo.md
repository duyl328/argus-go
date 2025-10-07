Argus – Codex TODO (prioritized)

Near-term
- Frontend API base alignment
  - Create .env.local.example with VITE_APP_API_URL=http://127.0.0.1:8080/api
  - Update front docs to mention required /api prefix
  - Acceptance: services resolve correct base; sample calls succeed

- Indexing progress API + SSE
  - Expose ImgTaskManager state: queued, running, doneCount, per-task step/progress
  - Endpoint: GET /api/v1/index/status
  - SSE: event "index-progress" periodic broadcast while indexing
  - Acceptance: UI can show live progress while /library/indexed runs

- EXIF caching
  - ImageAPI.GetExif: first try ExifRepository by hash; only call exiftool if missing
  - Save EXIF during indexing; avoid re-extraction on read paths
  - Acceptance: repeated asset/photo requests do not spawn exiftool

- Converted originals path separation
  - Store converted-original files under AppDir/cache/converted/{hash}.{ext}
  - Keep thumbnails under AppDir/cache/thumbnail
  - Acceptance: paths are separated; cleanup policies can differ

- Robust multipart streaming parser (frontend)
  - Replace text-decoder approach with byte-wise boundary parser for multipart/mixed
  - Handle binary data correctly; progressive UI updates
  - Acceptance: large batches stream reliably; no parse errors on edge cases

Security & Ops
- Gate filesystem APIs behind config flag and optional auth (token)
  - Add app.security.enable_filesystem_api (default true for dev)
  - Return 403 when disabled; add CORS tighten option
  - Acceptance: toggling flag disables routes safely

- Logging/i18n cleanup
  - Normalize log messages (avoid mojibake); optionally dual-language or English-only logs
  - Acceptance: clean logs on Windows console and files

Quality & Documentation
- Expand API_USAGE.md and generate OpenAPI spec
  - Document params/types for photos, assets, filesystem, SSE
  - Acceptance: up-to-date reference usable by frontend or tooling

- Packaging
  - Dockerfile (Linux) with preinstalled exiftool/imagemagick/libvips
  - CI script to build multi-platform bundles with tools
  - Acceptance: one-command image build runs backend

Backlog
- AI integration
  - Define interface to Python service (HTTP/gRPC) for faces/objects/scene/OCR
  - Persist to PhotoAiMetadataTable; expose search endpoints

- Timeline & search
  - Enriched filters (camera/lens/ISO/aperture/GPS bbox), keyword on EXIF+AI

- Auth & multi-user
  - Users/roles; per-library permissions; session/token model

- Tests
  - Repository tests for Photo/Exif timelines & paging
  - Handlers smoke tests (photo/assets/photos) with tool-availability guards

