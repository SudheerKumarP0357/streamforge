# StreamForge — Prompt Playbook
> A structured, dependency-ordered set of prompts for building the StreamForge video streaming platform with an AI coding assistant.
> 
> **How to use:** Start every session by pasting the contents of `STREAMFORGE_DEV_README.md` first, then use the prompts below in order. Each prompt block is self-contained and copy-paste ready.

---

## HOW TO START EVERY SESSION

Paste this at the top of every new AI conversation before any prompt below:

```
I am building StreamForge, a video streaming platform. 
Below is the full project context. Read it carefully before generating any code.
Strictly follow all conventions, file paths, tech choices, and constraints defined in it.
Do NOT generate infrastructure code (Terraform, Kubernetes, Dockerfiles, CI/CD).
Focus only on application code.

[PASTE CONTENTS OF STREAMFORGE_DEV_README.md HERE]

Acknowledge you have read the context and are ready to begin.
```

---

## DEPENDENCY MAP

Read this before starting. Each stage unlocks the next.

```
STAGE 1: Project Scaffolding
    │
    ├── 1A: Monorepo + Go API skeleton
    ├── 1B: Next.js frontend skeleton
    └── 1C: Go Transcoder skeleton
            │
STAGE 2: Database Layer          ← depends on 1A
    │
    ├── 2A: PostgreSQL migrations
    ├── 2B: Go DB clients (pgx)
    └── 2C: Cosmos DB + Redis clients
            │
STAGE 3: Core Backend — Auth     ← depends on 2A, 2B
    │
    ├── 3A: User model + password hashing
    ├── 3B: JWT generation + validation
    ├── 3C: Register + Login handlers
    └── 3D: Auth middleware
            │
STAGE 4: Core Backend — Videos   ← depends on 3D, 2C (blob + queue)
    │
    ├── 4A: Azure Blob client
    ├── 4B: RabbitMQ publisher
    ├── 4C: Video upload handler
    └── 4D: Video list + get handlers
            │
STAGE 5: Transcoder Worker       ← depends on 4B, 4A
    │
    ├── 5A: RabbitMQ consumer
    ├── 5B: FFmpeg wrapper
    ├── 5C: Full pipeline (consume → transcode → upload → update DB)
    └── 5D: Dead letter queue + error handling
            │
STAGE 6: Streaming Endpoint      ← depends on 4D, 5C
    │
    └── 6A: SAS URL generation + /stream endpoint
            │
STAGE 7: Frontend — Auth         ← depends on 3C
    │
    ├── 7A: API client (lib/api.ts)
    ├── 7B: Auth types + cookie helpers
    ├── 7C: Login page
    ├── 7D: Register page
    └── 7E: Next.js auth middleware
            │
STAGE 8: Frontend — Core Pages   ← depends on 6A, 7E
    │
    ├── 8A: Dashboard page (video list)
    ├── 8B: Upload page + UploadForm component
    └── 8C: Watch page + VideoPlayer component
            │
STAGE 9: Watch History           ← depends on 8C, 2C (Cosmos)
    │
    ├── 9A: Cosmos DB write (Go API)
    └── 9B: Watch history page (frontend)
            │
STAGE 10: Health + Polish        ← depends on all above
    │
    ├── 10A: /healthz endpoint
    ├── 10B: Redis rate limiting middleware
    └── 10C: Error boundaries + loading states
```

---

## ════════════════════════════════════════
## STAGE 1 — PROJECT SCAFFOLDING
## ════════════════════════════════════════

### PROMPT 1A — Go API Skeleton

```
Using the monorepo structure and Go conventions in the README context:

Generate the complete skeleton for `backend/api` including:

1. `go.mod` with module name `streamforge/api` and these exact dependencies:
   - github.com/go-chi/chi/v5
   - github.com/golang-jwt/jwt/v5
   - github.com/jackc/pgx/v5
   - go.mongodb.org/mongo-driver
   - github.com/redis/go-redis/v9
   - github.com/rabbitmq/amqp091-go
   - github.com/Azure/azure-sdk-for-go/sdk/storage/azblob
   - github.com/joho/godotenv

2. `cmd/server/main.go` that:
   - Loads .env file using godotenv (only in development)
   - Reads PORT from environment (default 8080)
   - Initialises all dependencies (DB, Redis, RabbitMQ, Blob) — use placeholder init functions for now
   - Registers the router
   - Starts the HTTP server with graceful shutdown on SIGINT/SIGTERM

3. `internal/router/router.go` that:
   - Creates a chi router
   - Adds chi middleware: Logger, Recoverer, RealIP
   - Has placeholder route groups: /api/v1/auth, /api/v1/videos, /api/v1/watch, /healthz
   - Returns the router

4. `internal/config/config.go` that:
   - Defines a Config struct with all fields from the ENV VARIABLES section of the README
   - Has a Load() function that reads all env vars with os.Getenv and returns Config
   - Panics with a clear message if any required variable is missing

Do not implement any handler logic yet. Show only the skeleton.
Generate each file separately with its full path as a header.
```

---

### PROMPT 1B — Next.js Frontend Skeleton

```
Using the monorepo structure and frontend conventions in the README context:

Generate the complete skeleton for `frontend/` including:

1. `package.json` with:
   - Next.js 14, React 18, TypeScript
   - Tailwind CSS + @tailwindcss/forms
   - video.js + @types/video.js
   - clsx (for conditional classnames)
   - No Redux, no Zustand — plain React state only

2. `next.config.js` that:
   - Enables strict mode

3. `tailwind.config.ts` with:
   - Content paths for app/ and components/
   - A dark mode class strategy

4. `app/layout.tsx` (Root layout) that:
   - Sets html lang="en"
   - Imports globals.css
   - Renders a <NavBar /> and {children}
   - Is a Server Component

5. `app/globals.css` with Tailwind base directives

6. `components/NavBar.tsx` that:
   - Is a Client Component ('use client')
   - Shows "StreamForge" brand text on the left
   - Shows "Dashboard", "Upload" nav links
   - Shows "Logout" button that clears the auth cookie and redirects to /login
   - Is minimal — no complex styling needed yet

7. `lib/types.ts` with ALL TypeScript types from the README context exactly as specified

8. `middleware.ts` that:
   - Protects all routes except /login and /register
   - Reads the JWT cookie (name from NEXT_PUBLIC env or hardcoded 'sf_access_token')
   - Redirects to /login if cookie is missing
   - Allows the request through if cookie exists (token validation happens in Go API)

Generate each file separately with its full path as a header.
```

---

### PROMPT 1C — Go Transcoder Skeleton

```
Using the monorepo structure and Go conventions in the README context:

Generate the complete skeleton for `backend/transcoder` including:

1. `go.mod` with module name `streamforge/transcoder` and these dependencies:
   - github.com/jackc/pgx/v5
   - github.com/rabbitmq/amqp091-go
   - github.com/Azure/azure-sdk-for-go/sdk/storage/azblob
   - github.com/joho/godotenv

2. `cmd/worker/main.go` that:
   - Loads .env with godotenv (dev only)
   - Reads all env vars from the TRANSCODER section of the README
   - Initialises RabbitMQ consumer, PostgreSQL client, Azure Blob client
   - Starts consuming — blocks forever
   - Handles SIGINT/SIGTERM gracefully (stop consuming, close connections)

3. `internal/config/config.go` with all transcoder env vars as a struct

4. `internal/consumer/rabbitmq.go` with:
   - A Consumer struct holding the amqp connection and channel
   - NewConsumer(url string) function
   - A Start(handler func(body []byte) error) function skeleton — no logic yet
   - A TranscodeJob struct matching the message format in the README

5. `internal/ffmpeg/transcode.go` with:
   - A Transcoder struct with FfmpegPath and TempDir fields
   - A Transcode(videoID string, inputPath string, outputDir string) error function — skeleton only, no FFmpeg calls yet

6. `internal/storage/blob.go` skeleton (same pattern as API)

7. `internal/db/postgres.go` with:
   - A DB struct wrapping pgx pool
   - Skeleton UpdateVideoStatus(ctx, videoID, status string) error function

Generate each file separately with its full path as a header.
```

---

## ════════════════════════════════════════
## STAGE 2 — DATABASE LAYER
## ════════════════════════════════════════

> **Prerequisite:** Stage 1 complete. You have a running Go skeleton.

### PROMPT 2A — PostgreSQL Migrations

```
Using the DATABASE SCHEMAS section of the README context:

Generate all 4 PostgreSQL migration files for `backend/api/migrations/`:

1. `001_create_users.up.sql` — users table exactly as specified
2. `002_create_videos.up.sql` — videos table with status check constraint
3. `003_create_renditions.up.sql` — hls_renditions AND refresh_tokens tables
4. `001_create_users.down.sql` — DROP TABLE users CASCADE
5. `002_create_videos.down.sql` — DROP TABLE videos CASCADE
6. `003_create_renditions.down.sql` — DROP both tables

Additionally generate `internal/db/migrate.go` in the Go API that:
- Uses golang-migrate/migrate with pgx/v5 driver
- Has a RunMigrations(databaseURL string) error function
- Runs migrations from the migrations/ directory embedded with go:embed
- Is called from main.go before starting the server
- Logs how many migrations were applied

Use the exact schema from the README. Do not add any extra columns.
```

---

### PROMPT 2B — PostgreSQL Client (Go API)

```
Using the Go conventions and database schema in the README context:

Generate `backend/api/internal/db/postgres.go` that:

1. Defines a DB struct wrapping *pgxpool.Pool

2. Has NewDB(ctx context.Context, connString string) (*DB, error) that:
   - Creates a pgxpool with max 10 connections
   - Pings to verify connection
   - Returns the DB

3. Implements these methods (full implementation, not skeletons):

   User methods:
   - CreateUser(ctx, email, passwordHash string) (string, error) — returns user UUID
   - GetUserByEmail(ctx, email string) (*models.User, error)
   - GetUserByID(ctx, id string) (*models.User, error)

   Video methods:
   - CreateVideo(ctx, userID, title, description string) (string, error) — returns video UUID, status='pending'
   - GetVideoByID(ctx, id string) (*models.Video, error)
   - ListVideosByUser(ctx, userID string, page, limit int) ([]models.Video, int, error)
   - UpdateVideoStatus(ctx, videoID, status string) error
   - UpdateVideoRawBlobURL(ctx, videoID, blobURL string) error
   - DeleteVideo(ctx, id string) error

   Rendition methods:
   - CreateRendition(ctx, videoID, resolution, playlistURL string, bandwidth int) error
   - GetRenditionsByVideoID(ctx, videoID string) ([]models.HLSRendition, error)

   Refresh token methods:
   - CreateRefreshToken(ctx, userID, tokenHash string, expiresAt time.Time) error
   - GetRefreshToken(ctx, tokenHash string) (*models.RefreshToken, error)
   - RevokeRefreshToken(ctx, tokenHash string) error

Also generate `internal/models/user.go` and `internal/models/video.go` with all structs.
Use pgx/v5 directly. No ORM. No global variables.
```

---

### PROMPT 2C — Cosmos DB + Redis Clients

```
Using the Go conventions and Redis key patterns in the README context:

1. Generate `backend/api/internal/db/cosmos.go` that:
   - Wraps *mongo.Client and *mongo.Database
   - Has NewCosmos(ctx, connectionString, dbName string) (*CosmosDB, error)
   - Implements:
     * SaveWatchEvent(ctx, event WatchEvent) error  — inserts to video_events collection
     * UpsertWatchHistory(ctx, history WatchHistory) error — upsert by user_id+video_id
     * GetWatchHistory(ctx, userID string, limit int) ([]WatchHistory, error)
     * UpsertUserPreferences(ctx, prefs UserPreferences) error
     * GetUserPreferences(ctx, userID string) (*UserPreferences, error)
   - Define WatchEvent, WatchHistory, UserPreferences structs matching the README shapes

2. Generate `backend/api/internal/cache/redis.go` that:
   - Wraps *redis.Client
   - Has NewRedis(url, password string) (*Cache, error)
   - Implements:
     * SetSession(ctx, userID string, data SessionData, ttl time.Duration) error
     * GetSession(ctx, userID string) (*SessionData, error)
     * DeleteSession(ctx, userID string) error
     * IncrRateLimit(ctx, ip, endpoint string) (int64, error) — key pattern from README, TTL 60s
     * GetRateLimit(ctx, ip, endpoint string) (int64, error)
     * SetVideoMeta(ctx, videoID string, video interface{}, ttl time.Duration) error
     * GetVideoMeta(ctx, videoID string) (string, error)
   - Define SessionData struct: { UserID, Email, Role string }

Use exact key patterns from the README Redis section. No global state.
```

---

## ════════════════════════════════════════
## STAGE 3 — CORE BACKEND: AUTH
## ════════════════════════════════════════

> **Prerequisite:** Stage 2 complete. DB, Redis clients implemented.

### PROMPT 3A — Password Hashing + User Model

```
Generate `backend/api/internal/auth/password.go` that:

1. Has HashPassword(password string) (string, error) using bcrypt cost 12
2. Has CheckPassword(password, hash string) bool
3. Validates password minimum length of 8 characters

Generate `backend/api/internal/auth/validate.go` that:
1. Has ValidateEmail(email string) bool — basic format check using regexp
2. Has ValidateRegistration(email, password string) error — returns descriptive error messages

No external validation libraries. Standard library only.
```

---

### PROMPT 3B — JWT Service

```
Using the auth flow described in the README context:

Generate `backend/api/internal/auth/jwt.go` that:

1. Defines a JWTService struct with fields: secret string, accessExpiry time.Duration, refreshExpiry time.Duration

2. Has NewJWTService(secret string) *JWTService

3. Implements:

   GenerateAccessToken(userID, email, role string) (string, error)
   - Uses HS256
   - Claims: sub=userID, email, role, exp, iat, jti (random UUID for uniqueness)
   - Returns signed token string

   GenerateRefreshToken() (string, error)
   - Returns a cryptographically random 32-byte hex string (not a JWT)
   - This gets stored hashed in PostgreSQL

   ValidateAccessToken(tokenString string) (*Claims, error)
   - Parses and validates signature + expiry
   - Returns custom Claims struct: { UserID, Email, Role string, StandardClaims }

   HashRefreshToken(token string) string
   - SHA-256 hash of the raw token (for DB storage)

Use golang-jwt/jwt/v5 only. No other JWT libraries.
```

---

### PROMPT 3C — Auth Handlers

```
Using the API contract and Go conventions in the README context:

Generate `backend/api/internal/handlers/auth.go` with a Handler struct and these methods:

1. Register(w, r):
   - Decode JSON body: { email, password }
   - Validate with auth.ValidateRegistration
   - Check if email already exists (GetUserByEmail) — return 409 if exists
   - Hash password, create user in DB
   - Return 201: { user_id, email }

2. Login(w, r):
   - Decode JSON body: { email, password }
   - GetUserByEmail — return 401 if not found
   - Check password hash — return 401 if wrong (same error message as not found for security)
   - Generate access token + refresh token
   - Hash refresh token, store in DB via CreateRefreshToken
   - Store session in Redis
   - Return 200: { access_token, refresh_token, expires_in: 3600 }

3. Refresh(w, r):
   - Decode body: { refresh_token }
   - Hash the token, look up in DB
   - Check: not revoked, not expired
   - Generate new access token
   - Return 200: { access_token, expires_in: 3600 }

4. Logout(w, r) [protected route]:
   - Get userID from request context
   - Delete Redis session
   - Return 200: { message: "logged out" }

Use respond() and respondError() helpers. No global state. 
Handler struct takes DB, Cache, JWTService as constructor arguments.
```

---

### PROMPT 3D — Auth Middleware

```
Using the auth flow in the README context:

Generate `backend/api/internal/middleware/auth.go` that:

1. Defines an AuthMiddleware struct with JWTService and Cache as fields

2. Implements Authenticate(next http.Handler) http.Handler that:
   - Reads Authorization header
   - Strips "Bearer " prefix
   - Validates token using JWTService.ValidateAccessToken
   - Checks Redis: GetSession(userID) — if missing, treat as logged out, return 401
   - On success: adds userID, email, role to request context using typed context keys
   - On any failure: return 401 JSON error immediately

3. Implements RequireRole(role string) func(http.Handler) http.Handler that:
   - Reads role from context (set by Authenticate)
   - Returns 403 if role doesn't match

4. Exports context key helpers: GetUserID(r *http.Request) string, GetUserRole(r *http.Request) string

Generate `backend/api/internal/middleware/cors.go` that:
- Sets Access-Control-Allow-Origin from CORS_ALLOWED_ORIGIN env var
- Allows methods: GET, POST, PUT, DELETE, OPTIONS
- Allows headers: Content-Type, Authorization
- Handles preflight OPTIONS requests

Generate `backend/api/internal/middleware/ratelimit.go` that:
- Reads client IP from X-Real-IP or RemoteAddr
- Calls Cache.IncrRateLimit with the endpoint path
- Returns 429 with Retry-After header if count > 100
```

---

## ════════════════════════════════════════
## STAGE 4 — CORE BACKEND: VIDEOS
## ════════════════════════════════════════

> **Prerequisite:** Stage 3 complete. Auth working end-to-end.

### PROMPT 4A — Azure Blob Client

```
Using the Azure Blob URL patterns in the README context:

Generate `backend/api/internal/storage/blob.go` that:

1. Defines a BlobStorage struct with: accountName, accountKey, rawContainer, hlsContainer strings and azblob.Client

2. Has NewBlobStorage(accountName, accountKey, rawContainer, hlsContainer string) (*BlobStorage, error)

3. Implements:

   UploadRawVideo(ctx, videoID string, reader io.Reader, size int64, contentType string) (string, error)
   - Uploads to container: rawContainer, blob name: {videoID}.mp4
   - Returns the full blob URL (no SAS — raw URL)
   - Uses azblob.UploadStreamOptions with block size 4MB

   GenerateHLSSASURL(ctx, videoID string, filename string) (string, error)
   - Generates a SAS URL for hls/{videoID}/{filename}
   - Expiry: 4 hours from now
   - Permissions: read only
   - Returns the full SAS URL

   UploadHLSFile(ctx, videoID, filename string, data []byte, contentType string) error
   - Uploads to hlsContainer, path: {videoID}/{filename}
   - Content type: application/x-mpegURL for .m3u8, video/MP2T for .ts

   DeleteRawVideo(ctx, videoID string) error
   - Deletes rawContainer/{videoID}.mp4

   DeleteHLSFolder(ctx, videoID string) error
   - Lists and deletes all blobs with prefix hls/{videoID}/

Use azure-sdk-for-go azblob package. Return meaningful wrapped errors.
```

---

### PROMPT 4B — RabbitMQ Publisher

```
Using the RabbitMQ configuration in the README context:

Generate `backend/api/internal/queue/rabbitmq.go` that:

1. Defines a Publisher struct with amqp connection and channel

2. Has NewPublisher(url string) (*Publisher, error) that:
   - Connects to RabbitMQ
   - Declares the exchange: name=transcoder, type=direct, durable=true
   - Declares the queue: transcoder.jobs, durable=true, with x-dead-letter-exchange arg
   - Binds queue to exchange with routing key: transcode
   - Declares DLX exchange: transcoder.dlx
   - Declares DLQ: transcoder.jobs.dlq bound to transcoder.dlx

3. Implements PublishTranscodeJob(ctx, videoID, rawBlobURL, userID string) error that:
   - Creates the JSON message matching the README TranscodeJob format
   - Publishes to transcoder exchange with routing key transcode
   - Sets delivery mode: Persistent (2)
   - Sets content type: application/json

4. Implements Close() error

Include reconnection logic: if publish fails with connection error, attempt reconnect once.
```

---

### PROMPT 4C — Video Upload Handler

```
Using the API contract (POST /api/v1/videos) in the README context:

Generate the Upload method in `backend/api/internal/handlers/videos.go`:

The VideoHandler struct takes: DB *db.DB, Storage *storage.BlobStorage, Queue *queue.Publisher, Cache *cache.Cache as fields.

Implement Upload(w, r) that:
1. Gets userID from request context (set by auth middleware)
2. Parses multipart form with max 2GB limit
3. Reads fields: title (required), description (optional)
4. Reads the file field named "file" — validate it is a video/* content type
5. Creates video record in PostgreSQL: status='pending'
6. Uploads raw file to Azure Blob using Storage.UploadRawVideo
7. Updates video record with raw_blob_url
8. Publishes transcode job to RabbitMQ via Queue.PublishTranscodeJob
9. Returns 201: { video_id, status: "pending" }

Error handling:
- If blob upload fails: update video status to 'failed', return 500
- If queue publish fails: log the error but still return 201 (job can be requeued manually)
- All errors must be wrapped with context

Use XMLHttpRequest on the client side for upload progress — the handler just needs to stream the multipart correctly.
Implement Delete(w, r) that:
- Gets video by ID, verifies it belongs to the authenticated user
- Deletes from PostgreSQL (cascade handles renditions)
- Deletes blob files (raw + HLS folder)
- Returns 204
```

---

### PROMPT 4D — Video List + Get Handlers

```
Continuing `backend/api/internal/handlers/videos.go`:

Implement these additional methods on VideoHandler:

1. List(w, r):
   - Gets userID from context
   - Reads query params: page (default 1), limit (default 20, max 100), status (optional filter)
   - Calls DB.ListVideosByUser
   - For each video with status='ready': fetches renditions from DB, attaches to response
   - Checks Redis cache first (GetVideoMeta) — if hit, return cached response
   - On cache miss: fetch from DB, set cache with 5 min TTL
   - Returns 200: { videos: [...], total: int, page: int, limit: int }

2. GetByID(w, r):
   - Gets videoID from chi URL param
   - Gets userID from context
   - Fetches video — return 404 if not found
   - Return 403 if video.UserID != userID (and user is not admin)
   - Fetches renditions if status='ready'
   - Returns 200 with full video object + renditions

3. GetStreamURL(w, r):
   - Gets videoID from URL param
   - Fetches video — return 404 if not found, 403 if not owner
   - Return 400 if status != 'ready'
   - Generates SAS URL for master.m3u8 via Storage.GenerateHLSSASURL
   - Returns 200: { master_playlist_url: string }

Wire all handlers into the router in `internal/router/router.go`.
Protected routes use the Authenticate middleware from Stage 3D.
```

---

## ════════════════════════════════════════
## STAGE 5 — TRANSCODER WORKER
## ════════════════════════════════════════

> **Prerequisite:** Stage 4 complete. Messages are being published to RabbitMQ.

### PROMPT 5A — RabbitMQ Consumer

```
Using the RabbitMQ configuration and pipeline logic in the README context:

Generate `backend/transcoder/internal/consumer/rabbitmq.go` with full implementation:

1. Consumer struct with: conn *amqp.Connection, ch *amqp.Channel, queueName string

2. NewConsumer(url, queueName string) (*Consumer, error):
   - Connects and creates channel
   - Sets prefetch count to 1 (QoS)
   - Does NOT re-declare queues (API service already declared them)

3. TranscodeJob struct exactly matching the README message format

4. Start(ctx context.Context, handler func(ctx context.Context, job TranscodeJob) error) error:
   - Calls ch.Consume with autoAck=false
   - For each delivery:
     * Unmarshal body into TranscodeJob
     * Call handler(ctx, job)
     * If handler returns nil: d.Ack(false)
     * If handler returns error: d.Nack(false, false) — sends to DLQ, does NOT requeue
   - Blocks until ctx is cancelled or channel closes
   - Returns when done

5. Close() error — closes channel and connection

Handle amqp.ErrClosed gracefully — log and return.
```

---

### PROMPT 5B — FFmpeg Wrapper

```
Using the VIDEO PIPELINE section and exact FFmpeg command in the README context:

Generate `backend/transcoder/internal/ffmpeg/transcode.go` with full implementation:

1. Transcoder struct with FfmpegPath, TempDir string

2. TranscodeResult struct:
   { VideoID string, OutputDir string, Renditions []RenditionResult }
   RenditionResult: { Resolution string, PlaylistFile string, SegmentFiles []string, Bandwidth int }

3. Transcode(ctx context.Context, videoID, inputPath string) (*TranscodeResult, error):
   - Creates output directory: {TempDir}/{videoID}/output/
   - Builds the exact FFmpeg command from the README (4 renditions: 1080p, 720p, 480p, 360p)
   - Runs ffmpeg using exec.CommandContext(ctx, ...) so it respects cancellation
   - Captures stderr for error logging
   - On exit code != 0: return error with ffmpeg stderr output included
   - After success: scans output directory, collects all .m3u8 and .ts files per resolution
   - Returns TranscodeResult with all file paths

4. GenerateMasterPlaylist(result *TranscodeResult, outputDir string) error:
   - Generates master.m3u8 exactly as shown in the README
   - Writes to {outputDir}/master.m3u8
   - Bandwidth values: 1080p=5128000, 720p=2928000, 480p=1496000, 360p=764000

5. Cleanup(videoID string) error:
   - Removes entire {TempDir}/{videoID}/ directory

Use os/exec only. No FFmpeg Go bindings.
```

---

### PROMPT 5C — Full Transcoder Pipeline

```
Using the complete 12-step VIDEO PIPELINE in the README context:

Generate `cmd/worker/main.go` and tie everything together:

Create a Pipeline struct in `internal/pipeline/pipeline.go` with fields:
  consumer *consumer.Consumer
  transcoder *ffmpeg.Transcoder  
  storage *storage.BlobStorage
  db *db.DB

Implement Process(ctx context.Context, job consumer.TranscodeJob) error that executes ALL 12 steps from the README:

Step 1: Download raw file from Azure Blob to /tmp/{videoID}/input.mp4
  - Use azblob client to download the raw_blob_url
  - Create temp directory first

Step 2: Update PostgreSQL video status to 'processing'

Step 3: Run ffmpeg.Transcoder.Transcode()

Step 4: Generate master.m3u8 with GenerateMasterPlaylist()

Step 5: Upload all HLS files to Azure Blob
  - Upload master.m3u8 first
  - Upload each rendition .m3u8
  - Upload all .ts segment files
  - Use correct content types (application/x-mpegURL for m3u8, video/MP2T for ts)

Step 6: Insert hls_renditions rows in PostgreSQL for each rendition

Step 7: Update PostgreSQL video status to 'ready'

Step 8: Clean up temp directory

Error handling at each step:
  - If any step fails: update status to 'failed', log the error with videoID context, return error
  - Cleanup temp dir even on failure (defer)

In main.go: wire Pipeline, start consumer, pass Pipeline.Process as the handler.
Log each step with structured logging (log/slog).
```

---

### PROMPT 5D — Dead Letter Queue Handler

```
Generate `backend/transcoder/internal/consumer/dlq.go` that:

1. DLQConsumer struct — same pattern as Consumer but for transcoder.jobs.dlq

2. Start(ctx, handler func(job TranscodeJob, failCount int) error) error:
   - Consumes from DLQ
   - Reads x-death header from amqp delivery to get failure count
   - If failCount < 3: calls handler (allowing retry logic)
   - If failCount >= 3: logs "job permanently failed" with videoID, acks message (no more retries)

3. In main.go: optionally start DLQ consumer alongside main consumer

Also add a DLQ monitoring log in the main consumer:
  - Every 60 seconds, log the current queue depth of both queues (use amqp channel QueueInspect)

This demonstrates understanding of message queue failure modes — important for DevOps portfolio.
```

---

## ════════════════════════════════════════
## STAGE 6 — STREAMING ENDPOINT VERIFICATION
## ════════════════════════════════════════

> **Prerequisite:** Stage 5 complete. Full pipeline working.

### PROMPT 6A — End-to-End API Test

```
Generate a shell script `scripts/test_pipeline.sh` that manually tests the full pipeline:

1. Register a user: POST /auth/register
2. Login: POST /auth/login — capture access_token
3. Upload a test video: POST /videos with a small test file (use curl)
4. Poll GET /videos/{id} every 3 seconds until status='ready' (max 10 polls)
5. Get stream URL: GET /videos/{id}/stream — capture master_playlist_url
6. Verify the SAS URL returns HTTP 200 with content-type application/x-mpegURL

Use curl and jq. Print clear pass/fail for each step.
Print the final master playlist URL for manual VLC testing.

Also generate `scripts/seed_test_video.sh` that downloads a small public domain video
(e.g., from archive.org) for use as a test file.

This script is not application code — it is a developer utility only.
```

---

## ════════════════════════════════════════
## STAGE 7 — FRONTEND: AUTH
## ════════════════════════════════════════

> **Prerequisite:** Stage 3 complete. Auth API endpoints working and tested.

### PROMPT 7A — API Client

```
Using the API contract, auth flow, and TypeScript conventions in the README context:

Generate `frontend/lib/api.ts` with a full API client implementation:

1. Base fetch wrapper fetchWithAuth(path, method, body?) that:
   - Reads JWT from cookie named 'sf_access_token'
   - Attaches Authorization: Bearer {token} header
   - On 401 response: calls attemptRefresh(), retries the original request once
   - On second 401: clears cookies, redirects to /login
   - Returns typed response or throws ApiError

2. attemptRefresh() that:
   - Reads 'sf_refresh_token' cookie
   - POSTs to /auth/refresh
   - On success: writes new access token to cookie
   - On failure: throws (caller will redirect to login)

3. Cookie helpers: getTokenCookie(), setTokenCookie(token), clearAuthCookies()
   - Work in both browser (document.cookie) and can be read server-side

4. Typed API methods:
   - auth.register(email, password): Promise<{user_id, email}>
   - auth.login(email, password): Promise<{access_token, refresh_token, expires_in}>
   - auth.logout(): Promise<void>
   - videos.list(page?, status?): Promise<{videos: Video[], total: number, page: number}>
   - videos.getById(id): Promise<Video>
   - videos.getStreamUrl(id): Promise<{master_playlist_url: string}>
   - videos.delete(id): Promise<void>
   - watch.saveEvent(videoId, eventType, positionSeconds): Promise<void>
   - watch.getHistory(): Promise<WatchHistory[]>

Use the Video and other types from lib/types.ts.
No axios. Fetch API only. TypeScript strict — no any.
```

---

### PROMPT 7B — Login Page

```
Using the API contract and frontend conventions in the README context:

Generate `frontend/app/(auth)/login/page.tsx` as a Client Component that:

1. Has a form with: email input, password input, submit button
2. On submit:
   - Calls api.auth.login(email, password)
   - On success: stores access_token in cookie, stores refresh_token in cookie, redirects to /dashboard
   - On error: shows error message below the form (e.g. "Invalid email or password")
3. Shows a loading spinner on the button while submitting
4. Has a link to /register ("Don't have an account? Register")
5. Validates client-side: email format, password not empty — before making API call
6. Uses Tailwind for styling — clean, minimal dark theme

Generate `frontend/app/(auth)/register/page.tsx` with:
1. Email, password, confirm password fields
2. Client-side: passwords must match, password >= 8 chars
3. Calls api.auth.register, then auto-logs in (calls api.auth.login), redirects to /dashboard
4. Shows specific error if email already taken (409 from API)
5. Same styling pattern as login page
```

---

## ════════════════════════════════════════
## STAGE 8 — FRONTEND: CORE PAGES
## ════════════════════════════════════════

> **Prerequisite:** Stage 7 complete. Auth cookies working. API client ready.

### PROMPT 8A — Dashboard Page

```
Using the frontend conventions and Video types in the README context:

Generate `frontend/app/dashboard/page.tsx` as a Server Component that:
1. Reads the auth cookie server-side using Next.js cookies()
2. Fetches videos from the Go API server-side (no client-side fetch)
3. Passes data to a client component for interactivity

Generate `frontend/components/VideoCard.tsx` as a Client Component that:
1. Props: Video object from types.ts
2. Shows: thumbnail placeholder (gray box with play icon), title, status badge, created_at formatted as relative time
3. Status badges:
   - pending: gray pill "Pending"
   - processing: yellow pill with animated spinner "Processing..."
   - ready: green pill "Ready" — entire card is clickable, links to /watch/{id}
   - failed: red pill "Failed" with a retry hint
4. When status is 'processing': polls GET /videos/{id} every 5 seconds using setInterval
   - Stops polling when status changes to 'ready' or 'failed'
   - Updates the card UI in place (no full page reload)
5. Has a delete button (trash icon) that calls api.videos.delete, removes card on success

Generate `frontend/components/VideoGrid.tsx`:
- Renders a responsive grid of VideoCards (3 cols desktop, 2 tablet, 1 mobile)
- Shows empty state: "No videos yet. Upload your first video →" with link to /upload
- Shows skeleton loading state

The dashboard page uses VideoGrid. Include a "Upload Video" button in the page header.
```

---

### PROMPT 8B — Upload Page

```
Using the API contract (POST /api/v1/videos multipart) in the README context:

Generate `frontend/app/upload/page.tsx` as a Client Component.

Generate `frontend/components/UploadForm.tsx` as a Client Component that:

1. Drag-and-drop zone:
   - Accepts video/* files only
   - Shows drag-over highlight state
   - Shows selected file name + size (formatted: "124.5 MB")
   - Has a fallback "Browse files" button

2. Form fields:
   - Title (required, max 100 chars)
   - Description (optional, textarea, max 500 chars)

3. Upload with progress:
   - MUST use XMLHttpRequest (not fetch) for upload progress events
   - Shows a progress bar: 0% → 100% as bytes upload
   - Shows upload speed estimate (MB/s)
   - Disables form during upload

4. After successful upload:
   - Shows success message: "Video uploaded! Transcoding will begin shortly."
   - Shows a "Go to Dashboard" button
   - Shows a "Upload Another" button that resets the form

5. Error states:
   - File too large (> 2GB): show error before upload starts
   - Network error during upload: show retry button
   - API error (4xx): show the error message from the API response

Use Tailwind only. No external upload libraries. No drag-and-drop libraries.
```

---

### PROMPT 8C — Watch Page + Video Player

```
Using the VideoPlayer spec in the README context:

Generate `frontend/app/watch/[id]/page.tsx` as a Server Component that:
1. Fetches video metadata server-side (title, description, status)
2. If status != 'ready': shows a "Video is still processing" message with back button
3. Fetches stream URL server-side: calls GET /videos/{id}/stream
4. Passes master_playlist_url to the VideoPlayer client component

Generate `frontend/components/VideoPlayer.tsx` as a Client Component that:

1. Props: { videoId: string, src: string, title: string }

2. Initialises Video.js on mount:
   - Options: { controls: true, autoplay: false, preload: 'auto', fluid: true }
   - Source: { src: props.src, type: 'application/x-mpegURL' }
   - Attach to a <div> ref using videojs(ref.current, options)

3. Event tracking — on each event, calls api.watch.saveEvent():
   - 'play' event: saveEvent(videoId, 'play', player.currentTime())
   - 'pause' event: saveEvent(videoId, 'pause', player.currentTime())
   - 'ended' event: saveEvent(videoId, 'end', player.currentTime())

4. Cleanup on unmount:
   - player.dispose() in useEffect return function
   - Remove all event listeners

5. Shows title above the player
6. Shows a back button to /dashboard

Important: Video.js must be imported dynamically (next/dynamic, ssr: false) because it uses browser APIs.
Include the Video.js CSS import: 'video.js/dist/video-js.css'
```

---

## ════════════════════════════════════════
## STAGE 9 — WATCH HISTORY
## ════════════════════════════════════════

> **Prerequisite:** Stage 8 complete. VideoPlayer emitting events.

### PROMPT 9A — Watch History API Handlers

```
Generate `backend/api/internal/handlers/watch.go` with WatchHandler struct (takes CosmosDB, DB as fields):

1. SaveEvent(w, r) [protected]:
   - Gets userID from context
   - Gets videoID from URL param
   - Decodes body: { event_type, position_seconds }
   - Validates event_type is one of: play, pause, seek, end
   - Saves to Cosmos DB video_events collection
   - If event_type is 'end': also upsert watch_history with completed=true
   - If event_type is 'play' or 'pause': upsert watch_history with progress_seconds
   - Returns 201

2. GetHistory(w, r) [protected]:
   - Gets userID from context
   - Reads limit query param (default 20, max 100)
   - Fetches from Cosmos DB watch_history collection for this user
   - For each history item: joins with PostgreSQL to get video title and status
   - Returns 200: { history: [{ video_id, title, watched_at, progress_seconds, completed }] }

Wire to router: POST /api/v1/watch/:video_id/event and GET /api/v1/watch/history
Both routes require Authenticate middleware.
```

---

### PROMPT 9B — Watch History Page

```
Generate `frontend/app/history/page.tsx` as a Server Component that:
1. Fetches watch history from GET /watch/history server-side
2. Shows list of watched videos: title, progress bar (progress_seconds / duration), watched_at date
3. Each item links to /watch/{video_id}
4. Shows "completed" badge if completed=true
5. Shows empty state: "You haven't watched any videos yet"

Add "History" link to NavBar.tsx.
```

---

## ════════════════════════════════════════
## STAGE 10 — HEALTH + POLISH
## ════════════════════════════════════════

> **Prerequisite:** All stages complete. Full application working.

### PROMPT 10A — Health Endpoint

```
Generate `backend/api/internal/handlers/health.go` with:

Health(w, r) handler that:
1. Checks PostgreSQL: runs SELECT 1, sets "db": "ok" or "db": "error: {msg}"
2. Checks Redis: runs PING, sets "redis": "ok" or "redis": "error"
3. Checks RabbitMQ: checks if channel is not nil and not closed, sets "rabbitmq": "ok" or "rabbitmq": "error"
4. Checks Cosmos DB: runs a Ping(), sets "cosmos": "ok" or "cosmos": "error"
5. Overall status: "ok" if all pass, "degraded" if any fail
6. HTTP status: 200 if all ok, 503 if any degraded
7. Includes: version (from env var APP_VERSION or "dev"), uptime in seconds

Returns JSON:
{
  "status": "ok",
  "version": "dev",
  "uptime_seconds": 142,
  "dependencies": {
    "db": "ok",
    "redis": "ok",
    "rabbitmq": "ok",
    "cosmos": "ok"
  }
}

Wire to GET /healthz — no auth middleware on this route.
```

---

### PROMPT 10B — Rate Limiting Middleware

```
Using the Redis rate limit implementation from Stage 2C:

Update `backend/api/internal/middleware/ratelimit.go` to be fully wired:

1. NewRateLimiter(cache *cache.Cache, limit int) func(http.Handler) http.Handler

2. The middleware:
   - Extracts client IP from X-Real-IP header, fallback to RemoteAddr (strip port)
   - Gets current route pattern from chi.RouteContext
   - Calls cache.IncrRateLimit(ctx, ip, route)
   - Sets response headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
   - If over limit: return 429 with JSON { "error": "rate limit exceeded", "retry_after": 60 }

Apply rate limiting in router.go:
- Global: 200 req/min on all routes
- Stricter: 10 req/min on POST /auth/login and POST /auth/register (brute force protection)

Add a chi middleware that logs each request: method, path, status code, duration, userID (if authenticated).
```

---

### PROMPT 10C — Error Handling + Loading States

```
Generate frontend error handling and loading polish:

1. `frontend/app/error.tsx` (Next.js error boundary):
   - 'use client'
   - Shows a friendly error message with a "Try again" button and "Go to Dashboard" link
   - Logs error to console in development

2. `frontend/app/not-found.tsx`:
   - Shows 404 message with link back to dashboard

3. `frontend/app/dashboard/loading.tsx`:
   - Shows skeleton grid (3x2 gray boxes animating with pulse)

4. `frontend/app/watch/[id]/loading.tsx`:
   - Shows a gray rectangle placeholder where the video player will be

5. Update `frontend/lib/api.ts` to:
   - Add request timeout: abort fetch after 30 seconds using AbortController
   - Add request ID header: X-Request-ID with a random UUID for tracing
   - Log all API errors to console in development with request details

6. In `frontend/components/UploadForm.tsx`:
   - Add client-side file validation: warn if file > 500MB (large file warning, not a block)
   - Add accepted format hint: "Supported formats: MP4, MOV, AVI, MKV"
```

---

## ════════════════════════════════════════
## BONUS PROMPTS — AFTER COMPLETION
## ════════════════════════════════════════

Use these only after all 10 stages are working.

### PROMPT B1 — Search

```
Add video search to the Go API and frontend:

Go API: GET /api/v1/videos/search?q={query}&page=1&limit=20 [protected]
- Uses PostgreSQL ILIKE: WHERE title ILIKE '%{query}%' AND user_id = $2
- Same response shape as /videos list
- Add index: CREATE INDEX idx_videos_title ON videos USING gin(to_tsvector('english', title));
- Use full text search: to_tsquery instead of ILIKE for better results

Frontend: Add a search input on the dashboard page
- Debounce input 400ms before making API call
- Show results inline, replacing the video grid
- Show "No results for '{query}'" if empty
- Clear search button to return to full list
```

---

### PROMPT B2 — Admin View

```
Add an admin role feature:

Go API:
- GET /api/v1/admin/videos [admin only, RequireRole("admin") middleware]
  Returns ALL videos from ALL users, paginated, with user email included
- GET /api/v1/admin/stats [admin only]
  Returns: { total_users, total_videos, videos_by_status: {pending, processing, ready, failed} }

Frontend:
- /admin/videos page (Server Component) — only renders if user role is 'admin' (check cookie claims)
- Table view of all videos with user email, status, created_at, delete button
- /admin/stats page — simple stat cards showing the aggregate numbers

Seed an admin user in a script: `scripts/create_admin.sh`
that calls register then runs a raw SQL UPDATE to set role='admin'
```

---

### PROMPT B3 — Prometheus Metrics

```
Add Prometheus instrumentation to the Go API:

1. Add dependency: github.com/prometheus/client_golang

2. Create `internal/metrics/prometheus.go` that defines and registers:
   - http_request_duration_seconds (histogram, labels: method, path, status)
   - http_requests_total (counter, labels: method, path, status)
   - active_connections (gauge)
   - blob_upload_duration_seconds (histogram, labels: container)
   - rabbitmq_published_total (counter, labels: queue, status)

3. Update the request logging middleware to observe http_request_duration_seconds and increment http_requests_total on each request

4. Wrap BlobStorage.UploadRawVideo to record timing in blob_upload_duration_seconds

5. Add GET /metrics endpoint (no auth) that serves Prometheus format using promhttp.Handler()

Do the same for the transcoder service:
- transcode_jobs_total (counter, labels: status)
- transcode_duration_seconds (histogram)
- rabbitmq_consumed_total (counter, labels: queue, status)
- Expose GET /metrics on a separate port (9090)
```

---

*StreamForge Prompt Playbook | 10 Stages | 25 Prompts | March 2026*
*Start every session with STREAMFORGE_DEV_README.md as context.*
