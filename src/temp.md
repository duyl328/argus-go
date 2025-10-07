Scratchpad (transient)

Indexing progress design sketch
- ImgTaskManager: expose snapshot { queued=len(queue), running=cap(workerPool)-len(workerPool), done=doneCount, total=queued+running+done }
- Per-task fields: { id, path, status, progress, current_step, updated_at }
- Handler: GET /api/v1/index/status returns { summary, tasks: [] (limited N) }
- SSE: broadcast event type "index-progress" every 1–2s while indexing active

Multipart streaming client (frontend)
- Implement streaming multipart parser on bytes: scan for `--boundary\r\n`, headers (CRLF), then raw body until `\r\n--boundary` or `\r\n--boundary--`
- Build Blob URLs per part; emit progress callbacks

Paths
- Keep converted originals under cache/converted; thumbnails under cache/thumbnail. Update HashThumbPath or add helper for converted path.

Notes
- Front default base URL is 3001; ensure .env.local overrides to backend http://127.0.0.1:8080/api
- Prefer caching EXIF at index-time; PhotoHandler should not re-run exiftool paths

