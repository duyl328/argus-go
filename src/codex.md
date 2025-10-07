Argus (go-argus/src) — Codex Context

Purpose
- Internal memory of the repo to accelerate future work: architecture, data flow, dependencies, conventions, pitfalls, and roadmap.

High‑Level Summary
- Project: Argus — cross‑platform photo management and image processing system.
- Back end (rear, Go 1.24): REST API with Gin, GORM (SQLite by default / MySQL optional), image pipeline via external tools (ExifTool, ImageMagick, LibVips), async task workflow, SSE real‑time events, filesystem browsing/ops.
- Front end (front/argus-front, Vue 3 + Vite + TS): File manager UI, timeline, SSE subscription, API services.
- Experimental python service (service/src-python-v1): YOLO/ArcFace demos; planned future AI integration.

Top‑Level Structure
- rear/                 Go backend
- front/argus-front/    Vue 3 frontend
- service/src-python-v1 Python experiments (YOLO/ArcFace)
- design/               HTML design prototypes

Backend (rear) Overview
- Entry: rear/main.go
  - Loads config (internal/config), initializes logging, optional dev-time tool extraction (tools/), detects tool paths, prints versions.
  - Initializes DB (internal/db), automigrates tables, sets up DI containers (internal/container), repositories, and image task manager (workflow/ImgTaskManager).
  - Creates cache dirs (cache/thumbnail, temp/png), then starts HTTP server with Gin; WriteTimeout=0 to support SSE streaming; pprof in debug.

- Config: internal/config
  - loader.go: FileConfig YAML schema (app, tools, paths, logging, database, image, task); GetDefaultFileConfig + MergeWithDefaults.
  - config.go: Populates global CONFIG with derived fields (ImageCompressionOption, PathConfig). Determines AppDir from executable dir. Development flag via app.development.
  - database.go: DatabaseConfig (sqlite/mysql). SQLite path processing: default AppDir/db/db.sqlite; ensures dir exists.
  - Utility getters: GetDatabaseConfig(), GetTaskConfig() (fills concurrency defaults), GetDefaultThumbnailSize().

- HTTP Layer
  - Router: internal/router/router.go
    - GET /, /health
    - /api/v1
      - /users (CRUD, basic example)
      - /library: GET/POST/PUT/DELETE; POST /indexed triggers indexing workflow
      - /exif: list, stats, cameras, search, gps, iso, aperture, camera
      - /photo: GET /:hash (original|thumbnail), GET /preview?path=...&size=...
      - /photo/batch-preview (multipart streaming thumbnails)
      - /photos: GET timeline, list (columnar format for performance)
      - /assets/:hash returns combined Photo + EXIF details
      - /filesystem: browse, disk-usage, item, search; file ops (create dir, delete, move, copy), and watcher management (watch/unwatch/watched)
      - /sse: connect, broadcast/send, clients, stats, subscriptions (subscribe/unsubscribe/get)
    - /dev/tool/exiftool/get_exif (dev endpoint)
  - Middleware: internal/service/service.go
    - RequestID, Logger (zap), ErrorHandler; CORS via gin-contrib/cors

- Models & DB
  - Tables: internal/model/tables
    - Photo: basic photo info, times, rating, view_count
    - PhotoExif: wide EXIF table with JSONMap OtherFields and helpers (ToParsedExif/FromParsedExif)
    - LibraryTable: paths to scan; enabled flag
    - PhotoAiMetadataTable (placeholder for future AI)
    - BaseModel with soft delete; User example
  - EXIF parse model: internal/model/exif_model.go — BaseImageInfo, ExifInfo, ParsedExif; SplitExifData() robustly splits exiftool output into the three structs.
  - Unified API response: internal/model/response.go
  - DB init & manager: internal/db
    - InitDatabase(): GORM open sqlite/mysql; when sqlite, set WAL and single-connection tuning.
    - AutoMigrate() registers tables.
    - DatabaseManager: write queue with retries; SubmitWriteTask/Sync; limited workers (1 for SQLite, 10 for MySQL).
    - Write tasks defined for Photo/Exif/User & view count updates.
  - Repositories: internal/repositories
    - PhotoRepository: CRUD, columnar list, timeline aggregation, date‑range queries, CreatePhotoFromImageAPI() to map exif into Photo; existence checks by hash/path.
    - ExifRepository: CRUD + filters (camera, iso, aperture), stats, keyword search, GPS; CreateOrUpdate.
    - LibraryRepository: CRUD; uses BaseService (single-threaded write queue) to avoid SQLite lock issues; GetAllLibrary.
    - base_service.go: ExecuteWrite/ExecuteRead; serializes writes only for SQLite via an internal channel worker.

- Image Pipeline: internal/api/img_api.go
  - NewImageAPI(path): reads file bytes, detects format with h2non/filetype, hashes file (SHA256), stores path/bytes/format/hash.
  - GetExif(): runs exiftool, splits to ParsedExif, logs key fields.
  - GetSupportOriginalImagePath(ctx): if format already jpg/jpeg/webp returns original path; else converts to target configured format (CONFIG.ImageCompressionOption.ThumbnailFormat) and stores under AppDir/cache/thumbnail/{hash}_original.{ext}.
  - GetImagePath(longSide): ensures supported original, returns original if longSide<=0 or requested size >= long side; else generates thumbnail by LibVips (fallback to ImageMagick→PNG→Vips).
  - Thumbnails: GetThumbnailPath() = AppDir/cache/thumbnail + HashUtils.HashThumbPath(hash, size, ext); generateStandardThumbnail/SmartThumbnails; special cropping for extreme ratios.
  - Smart generation: GetSmartThumbnailSizes(fileSize) => {512}, {512,1080}, {512,1080,2048}, {512,1080,2048,4096} based on size thresholds. Tiles (vips dzsave) for very large images.
  - ProcessImage(): EXIF → ensure supported original → image info → maybe tiles → smart thumbnails.

- External Tools Integration: internal/utils
  - tool_manager.go: detects tool paths (ImageMagick `magick` or `convert`, `exiftool`, `vips`) by scanning AppDir and PATH; ensureExecutable on Unix; ExecuteCommand() wrapper returns stdout/stderr/exit/status with structured logging.
  - tools/exif_tool.go: GetExifData/GetExifField/RemoveExif/CopyExif/SetExif; calls EnsureInitialized(); parses JSON.
  - tools/image_magick_tool.go: Convert/Resize/Crop/Rotate, IsImageMagickAvailable.
  - tools/libvips_tool.go: vipsheader info; Convert/Resize/Crop; CompressVipsImage; dzsave tiling; strip metadata; Quick* helpers; IsVipsAvailable/VipsVersion.
  - utils/filesystem_watcher.go: fsnotify watcher with subscription counts; normalizes events; event JSON marshalling; Start/Stop; used by FileSystemHandler to broadcast via SSE.
  - utils/task_scheduler.go: generic task scheduler (worker pool, stats) mainly as a utility; image/exif example tasks provided (not central to current pipeline which uses ImgTaskManager and db manager).
  - pkg/utils: rich file utils (exists/dir/file checks, recursive walks, copy/move/delete dir/file with options, disk/memory info, hashing SHA256/MD5, format size, image hashing etc.).
  - pkg/logger: zap wrapper with lumberjack rotation; global logger.* convenience functions; configurable via config.logging.
  - pkg/sse: robust SSE Manager with client registry, buffer, keepalive, ping, and subscription mapping. BroadcastToSubscribers(path, event, data) used by FS watcher.

- Handlers
  - PhotoHandler: GET /api/v1/photo/:hash (format=original|thumbnail&size=...); increments view_count async; GetPhotoAssets returns merged Photo + EXIF; GetPhotos returns columnar arrays with pagination & order; GetPhotoPreview supports path preview.
  - LibraryHandler: manage library paths; POST /indexed scans enabled directories and enqueue tasks to ImgTaskManager.
  - ExifHandler: query EXIF by hash; stats (cameras), filters (iso, aperture), GPS, search, list.
  - FileSystemHandler: wraps FileSystemService for browse & ops; creates FileSystemWatcher that streams events through SSE to subscribed clients; exposes watch/unwatch endpoints.
  - BatchThumbnailHandler: POST /api/v1/photo/batch-preview → multipart/mixed streaming; each part contains a thumbnail or an error JSON; concurrency limited to 5.
  - SSEHandler: /api/v1/sse/connect plus broadcast/send; subscription endpoints (subscribe/unsubscribe/get); emits keepalive every 5s; WriteTimeout=0 on server.

- Workflow: internal/workflow/img_task.go
  - PictureTask runs a multi‑step pipeline (validating → exif → convert → thumbnails → save db → intelligent analysis placeholder) with progress and status; saves EXIF and Photo via repositories.
  - ImgTaskManager queues PictureTask, controls concurrency (from config.task), supports global pause/resume and dynamic concurrency; background monitor to auto‑adjust in future.
  - Future intelligent analysis hooks into Python/AI service (faces, objects, scene, OCR) → PhotoAiMetadataTable.

- Containers: internal/container
  - DbContainer with repos; TaskContainer with ImgTaskManager (wires exif/photo repos).

- Services: internal/service/filesystem_service.go
  - Cross‑platform browse root/dirs with drive discovery (pkg/system/DeviceManager), disk usage based on OS; create/delete/move/copy with FileUtils; summarizes counts and formats sizes for UI.

- Config/Tooling/Build
  - rear/README.md explains tool dependencies, platform packaging, and quickstart; Windows zips are included under rear/tools/windows_amd64/; main.go can auto‑extract into AppDir/tools in development.
  - scripts/build.sh builds multi‑platform artifacts and assembles release bundles.
  - docs under rear/doc explain choice to avoid go‑vips on Windows and potential Docker+govips future.
  - REST .http tests under rear/tests/ for quick manual validation.

Frontend (front/argus-front) Overview
- Stack: Vue 3 + Vite + TypeScript + Pinia + Naive UI (seen in components).
- Config: src/config/httpConfig.ts with baseURL from VITE_APP_API_URL (default http://127.0.0.1:3001/api – override in .env to match backend /api/v1 base).
- HTTP: src/utils/http.ts wraps axios with interceptors, abort duplicate requests, unified ApiResponse handling; plugin registers $http and config.
- Services (src/services/*):
  - sseService.ts: EventSource to /v1/sse/connect; parses custom events (connected/keepalive/filesystem-change); supports subscribe/unsubscribe and reconnection/backoff.
  - fileSystemService.ts: browse, item info, file ops, disk usage, SSE integration for live updates; used by FileManager UI.
  - timelineService.ts: fetch /photos/timeline and per‑day photos using columnar list endpoints.
  - photoPreviewService.ts & batchThumbnailService.ts: preview by path; batch thumbnails via multipart/mixed streaming parser.
- Components: FileManager (FileManager.vue, FilePane.vue, DebugPanel.vue, PhotoPreviewModal.vue, ContextMenu.vue, etc.); composables for keyboard nav, drag select, virtual scroll, file selection; docs under src/docs about performance and watcher solution.

Data Flow (End‑to‑End)
1) Library Indexing
   - Add library: POST /api/v1/library { path }
   - Start indexing: POST /api/v1/library/indexed → scans enabled directories (BaseSupportedFileTypes), filters duplicates by PhotoRepo.GetByPath, enqueues PictureTask for new files.
   - PictureTask: ImageAPI.ProcessImage → EXIF via exiftool → convert originals to configured format if needed → generate smart thumbnails → save EXIF(PhotoExif) + Photo rows via repos (which use db task queue/manager for writes).

2) Photo Retrieval
   - GET /api/v1/photo/:hash?format=original|thumbnail&size=… → PhotoHandler fetches Photo by hash, increments view_count via async DB task, uses ImageAPI to provide file path and returns it as file; sets content-type by resulting format.
   - GET /api/v1/assets/:hash → merges Photo + EXIF into PhotoDetailResponse for UI detail panels.
   - GET /api/v1/photos (columnar lists) + /photos/timeline for home/timeline views.

3) Filesystem + SSE
   - Front opens SSE to /api/v1/sse/connect; on connected event, gets client_id; subscribes to a path via POST /api/v1/sse/subscribe.
   - Backend FileSystemWatcher emits events (create/modify/delete/rename) with watched_path; SSE Manager broadcasts to subscribers of that path as event type filesystem-change.
   - Front updates FileManager panes in real‑time.

Conventions & Patterns
- Responses use model.Response { code, message, data } consistently.
- Logging via rear/pkg/logger (zap) across all layers; Info/Warn/Error and Infof/Warnf wrappers.
- Tool execution wrappers call utils.EnsureInitialized() once (tool detection) and utils.ExecuteCommand() with structured logs and timeouts.
- Thumbnails & converted originals stored under AppDir/cache/thumbnail using HashUtils.HashThumbPath(hash, sizeTag, ext). Default sizes 512/1080/2048/4096. Default thumbnail format from config (jpg default).
- SQLite: single connection with WAL; writes funneled through a queue (DatabaseManager + base_service) to avoid lock contention.
- SSE: server WriteTimeout=0, keepalive every 5s; CORS allowed for all by default.

External Dependencies (selected)
- Gin, cors, pprof; GORM with glebarez/sqlite and MySQL; zap + lumberjack; fsnotify; filetype; google/uuid; yaml.v3; bytedance/sonic (indirect), etc.
- External tools: ExifTool, ImageMagick (magick/convert), LibVips (vips, vipsheader).

Notable Design Choices
- Avoid go‑vips on Windows (doc/argus.md) due to toolchain complexity; prefer calling platform tools directly and bundling zips for Windows.
- Keep WriteTimeout=0 for long‑lived SSE instead of WebSocket, with note to place behind reverse proxy in production.
- Columnar photo list responses to minimize payload and optimize virtual scroll in UI.

Setup Hints
- Backend default port 8080 (rear/config.yaml). Ensure external tools are available:
  - Dev on Windows: zip bundles under rear/tools/windows_amd64 auto‑extracted to AppDir/tools when IsDevelopment().
  - macOS/Linux: install via package manager (see rear/README.md) or configure CONFIG.tools paths.
- Frontend must set VITE_APP_API_URL to match backend prefix (e.g. http://127.0.0.1:8080/api). The default in httpConfig.ts is http://127.0.0.1:3001/api, so override in .env.local.

Quality / Tech Debt / Risks
- Multipart thumbnail streaming parser in front/services is text‑decoder based; binary parts may be fragile if boundaries appear in decoded text. Consider a robust multipart reader using ReadableStream and boundary parsing on bytes.
- ImageAPI stores converted “original” under cache/thumbnail; consider a separate cache path for converted originals to avoid mixing with thumbnails.
- Some strings in logs/comments show mojibake (encoding artifacts), likely Windows console encoding but worth normalizing.
- EXIF fetch does not check DB cache before running exiftool (TODO in code). Consider caching/extracting only once during indexing.
- SSE GetStats uptime uses a placeholder. Consider tracking real uptime; also expose queue/dequeue metrics from ImgTaskManager and DatabaseManager.
- Tool detection assumes vipsheader adjacent to vips; confirm with bundled layout. Add path validation and clearer error diagnostics in production mode.
- Filesystem operations do not currently enforce access control. Add auth/ACL if running beyond trusted local use.

Performance Notes
- SQLite tuned with WAL + single connection; write operations queued, minimizing “database is locked”. For large scale, switch to MySQL with higher write workers in DatabaseManager.
- Thumbnails generated smartly by file size; tiles for very large images to support zoomable views.
- Columnar API payload shapes reduce JSON overhead for large lists; UI relies on virtual scroll (see front/docs/virtual-scroll-performance.md).

Roadmap / Future Extensions (implied by code/docs)
- AI metadata: integrate service/src-python-v1 (YOLO, ArcFace) to populate PhotoAiMetadataTable (nsfw score, faces, objects, scenes, OCR) in workflow.performIntelligentAnalysis.
- Search & filters: build search endpoints across EXIF + AI metadata; geo clustering; camera/lens dashboards.
- Job orchestration & progress: expose ImgTaskManager task progress via SSE for indexing status; pause/resume controls in UI.
- Caching & CDN: cache headers with ETags, pre‑generate common thumbnails, serve via CDN.
- Importers: sidecar reading (XMP/JSON), video support (transcodes, thumbnails), RAW pipeline improvements (dcraw/rawtherapee integration).
- Auth & multi‑user: add users/roles, libraries per user, sessions.
- Packaging: cross‑platform release bundles with embedded tools per platform; Docker image with preinstalled tools for Linux/macOS; possibly containerized govips.

Key Files (fast recall)
- rear/main.go: startup, tool extraction/detection, HTTP server.
- rear/internal/router/router.go: route registrations; returns cleanup for SSE and watcher.
- rear/internal/api/img_api.go: image pipeline, conversion, thumbnails, EXIF, tiles, smart sizes.
- rear/internal/workflow/img_task.go: PictureTask + ImgTaskManager orchestration.
- rear/internal/repositories/*: DB access; base_service for SQLite serialization.
- rear/internal/handler/*: HTTP handlers (photo, library, exif, filesystem, sse, batch thumbnail).
- rear/internal/utils/tools/*.go: exiftool, imagemagick, libvips wrappers; tool_manager + ExecuteCommand.
- rear/pkg/sse/*.go: SSE manager, events, subscriptions.
- rear/pkg/utils/*.go: file ops, hashing, disk/mem info.
- front/argus-front/src/services/*: http client, sse, file system, timeline, thumbnails.

Operational Notes
- Ensure rear/config.yaml exists (else defaults are used). Paths in config are relative to AppDir when not absolute.
- Cache directories are created at startup: AppDir/cache/thumbnail and AppDir/temp/png.
- WriteTimeout=0 is intentional for SSE; in production, put behind nginx/traefik with sensible upstream timeouts.
- For Windows dev, tool zips are included and auto‑extracted; for other OS, install & set PATH or config.tools.*.

Scratchpad / TODO for future me
- Add /api/v1/index/status to expose indexing progress from ImgTaskManager (counts, queue length, history), stream updates via SSE.
- Normalize file extension detection vs stored Format to avoid MIME mismatches.
- Consider moving converted originals to AppDir/cache/converted/ to separate from thumbnails.
- Front: centralize API base in .env and ensure /api/v1 prefix is consistently used; audit services for hardcoded paths.

Git 提交规范（Codex 使用）

目标
- 统一、可读、可自动化解析的提交信息，便于回溯、自动生成变更日志、代码审查与持续集成。

规范基础
- 采用 Conventional Commits 1.0 风格，中文书写规范化。
- Header 结构：<type>(<scope>)!?: <subject>
  - type：变更类型（必填）
  - scope：影响范围/模块（可选），如 rear、front、api、handler、repo、config、db、workflow、utils、sse、scripts、docs 等
  - !：破坏性变更（可选），出现时必须在 Footer 提供 BREAKING CHANGE 说明
  - subject：一句话概述（必填，使用祈使语气，≤ 50 字符）
- Body：详细说明（可选），描述动机、实现要点、影响面、风险与回滚方式，72 字符换行
- Footer：关联信息（可选）
  - BREAKING CHANGE: 描述破坏性变更与迁移指引
  - Closes #123 / Refs #123 等 issue 关联

变更类型（type）
- feat：新功能
- fix：缺陷修复
- perf：性能优化
- refactor：重构（非功能、非修复）
- docs：文档与注释（含内部文档）
- style：代码风格（格式化、空格、分号等，无逻辑变更）
- test：测试用例
- build：构建系统或依赖调整（go.mod、vite、工具链）
- ci：持续集成配置
- chore：杂项（脚本、发布、依赖升级等）
- revert：回滚提交

书写要点
- subject 使用动词开头的祈使句（如：新增、修复、移除、重构、对齐、提高）
- 避免情绪词与无信息词（如：更新了、修改一下）
- Body 说明“为什么/影响/如何验证”，而非仅“做了什么”
- 提交粒度尽量小且完整可编译/可运行
- PR 标题也遵循该规范

示例
- feat(rear/api): 新增批量缩略图接口 multipart/mixed 流式返回

  - 支持并发处理与分块刷写，降低大批量请求延迟
  - Body 附：接口路径、请求体、响应格式与前端解析注意点
  - Closes #128

- fix(utils/tools): 修复 vipsheader 路径拼接在部分平台失败的问题

- docs(repo): 新增 Codex 自主记忆文件并添加提交规范

  - 新增 codex.md、todo.md、temp.md
  - 在 codex.md 记录架构、数据流、工具依赖与提交规范

破坏性变更示例
- refactor(api)!: 统一 /api/v1 路由前缀并调整响应结构

  BREAKING CHANGE: 原 /photo 与 /assets 等接口迁移至 /api/v1/*，响应字段 code/message/data 统一；前端需同步升级服务调用与类型定义

提交检查清单（建议）
- [ ] 类型与范围是否准确
- [ ] subject 是否简洁、命令式
- [ ] Body 是否说明动机/影响/验证
- [ ] 是否存在破坏性变更且已在 Footer 说明
- [ ] 是否需要关联 issue/需求编号

在本仓库的实践
- 常用 scope：rear、front、api、handler、repo、config、db、workflow、utils、sse、scripts、docs
- 文档/内部知识（如 codex.md、todo.md、temp.md）使用 docs(repo) 类型
- 仅工具链或脚手架改动使用 build/* 或 chore/*

End of codex.
