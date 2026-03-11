# STREAMFORGE — LLM Development Context File
> **For AI/LLM Use Only** — This file is a structured context document intended to be fed to a language model or AI coding assistant to guide development of the StreamForge video streaming platform. Read all sections before generating any code.

---

## DIRECTIVE

You are assisting in building **StreamForge**, a full-stack video streaming web application. Your role is **application development only**. Do NOT generate, suggest, or modify:
- Terraform files
- Kubernetes / Helm manifests
- GitHub Actions workflows
- Azure resource configuration
- Docker Compose or Dockerfile (unless explicitly asked)
- Any CI/CD pipeline configuration

Infrastructure is handled separately. Focus exclusively on **application code**: frontend, backend API, transcoder worker, database schemas, and inter-service communication logic.

---

## PROJECT IDENTITY

| Field | Value |
|---|---|
| **Project Name** | StreamForge |
| **Type** | Video Streaming Platform (Portfolio / DevOps Showcase) |
| **Goal** | Upload, transcode, and stream video content via a browser |
| **Audience** | Single developer portfolio project |
| **Scale** | Small — no need to over-engineer for millions of users |

---

## MONOREPO STRUCTURE

All application code lives in these directories. When generating files, always place them in the correct path.

```
streamforge/
├── frontend/                   # Next.js 14 application
│   ├── app/                    # App Router pages and layouts
│   │   ├── (auth)/             # Auth route group
│   │   │   ├── login/page.tsx
│   │   │   └── register/page.tsx
│   │   ├── dashboard/page.tsx  # Video listing page
│   │   ├── upload/page.tsx     # Video upload page
│   │   ├── watch/[id]/page.tsx # Video player page
│   │   ├── layout.tsx          # Root layout
│   │   └── globals.css
│   ├── components/             # Reusable UI components
│   │   ├── VideoPlayer.tsx     # Video.js wrapper component
│   │   ├── UploadForm.tsx      # Drag-and-drop upload
│   │   ├── VideoCard.tsx       # Card for video listings
│   │   └── NavBar.tsx
│   ├── lib/                    # Shared utilities
│   │   ├── api.ts              # API client (fetch wrappers)
│   │   ├── auth.ts             # JWT helpers (cookie read/write)
│   │   └── types.ts            # Shared TypeScript types
│   ├── middleware.ts            # Next.js middleware for auth protection
│   ├── next.config.js
│   ├── tailwind.config.ts
│   └── package.json
│
├── backend/
│   ├── api/                    # Go REST API service
│   │   ├── cmd/server/main.go  # Entry point
│   │   ├── internal/
│   │   │   ├── handlers/       # HTTP handler functions
│   │   │   │   ├── auth.go
│   │   │   │   ├── videos.go
│   │   │   │   └── watch.go
│   │   │   ├── middleware/     # Auth, logging, CORS middleware
│   │   │   │   ├── auth.go
│   │   │   │   └── cors.go
│   │   │   ├── models/         # Struct definitions
│   │   │   │   ├── user.go
│   │   │   │   └── video.go
│   │   │   ├── db/             # Database connection + queries
│   │   │   │   ├── postgres.go
│   │   │   │   └── cosmos.go
│   │   │   ├── cache/          # Redis client
│   │   │   │   └── redis.go
│   │   │   ├── queue/          # RabbitMQ publisher
│   │   │   │   └── rabbitmq.go
│   │   │   ├── storage/        # Azure Blob client
│   │   │   │   └── blob.go
│   │   │   └── router/         # Route registration
│   │   │       └── router.go
│   │   ├── migrations/         # SQL migration files
│   │   │   ├── 001_create_users.sql
│   │   │   ├── 002_create_videos.sql
│   │   │   └── 003_create_renditions.sql
│   │   └── go.mod
│   │
│   └── transcoder/             # Go transcoding worker service
│       ├── cmd/worker/main.go  # Entry point
│       ├── internal/
│       │   ├── consumer/       # RabbitMQ consumer
│       │   │   └── rabbitmq.go
│       │   ├── ffmpeg/         # FFmpeg wrapper
│       │   │   └── transcode.go
│       │   ├── storage/        # Azure Blob client (shared pattern)
│       │   │   └── blob.go
│       │   └── db/             # PostgreSQL client (update status)
│       │       └── postgres.go
│       └── go.mod
```

---

## TECH STACK — APPLICATION LAYER ONLY

### Frontend
| Item | Detail |
|---|---|
| **Framework** | Next.js 14 with App Router |
| **Language** | TypeScript (strict mode) |
| **Styling** | Tailwind CSS |
| **Video Player** | Video.js + `videojs-contrib-hls` / HLS.js |
| **State** | React `useState` / `useEffect` — no Redux needed at this scale |
| **Auth** | JWT stored in HTTP-only cookies; read server-side in layouts |
| **API calls** | Custom fetch wrapper in `lib/api.ts` |

### Backend — Go API (`backend/api`)
| Item | Detail |
|---|---|
| **Language** | Go 1.22 |
| **HTTP Router** | `chi` (lightweight, idiomatic Go router) |
| **Auth** | JWT using `golang-jwt/jwt/v5` |
| **PostgreSQL** | `pgx/v5` driver (NOT `database/sql` + `lib/pq`) |
| **Cosmos DB** | `mongo-driver` (Cosmos DB exposes Mongo-compatible API) |
| **Redis** | `go-redis/redis/v9` |
| **RabbitMQ** | `rabbitmq/amqp091-go` |
| **Azure Blob** | `azure-sdk-for-go` — `azstorage` / `azblob` package |
| **Migrations** | `golang-migrate/migrate` (run at startup) |
| **Config** | Environment variables loaded via `os.Getenv` — secrets come from Azure Key Vault via env injection at runtime |

### Backend — Go Transcoder (`backend/transcoder`)
| Item | Detail |
|---|---|
| **Language** | Go 1.22 |
| **Queue Consumer** | `rabbitmq/amqp091-go` |
| **Transcoding** | Shell out to `ffmpeg` binary via `os/exec` |
| **Storage** | Azure Blob Storage (`azblob`) |
| **DB** | `pgx/v5` — only used to update video status |

---

## DATABASE SCHEMAS

### PostgreSQL (Source of Truth for structured data)

```sql
-- 001_create_users.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 002_create_videos.sql
CREATE TABLE videos (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title            TEXT NOT NULL,
    description      TEXT,
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
    raw_blob_url     TEXT,
    duration_seconds INTEGER,
    file_size_bytes  BIGINT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 003_create_renditions.sql
CREATE TABLE hls_renditions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    video_id     UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    resolution   TEXT NOT NULL CHECK (resolution IN ('360p', '480p', '720p', '1080p')),
    playlist_url TEXT NOT NULL,
    bandwidth    INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Cosmos DB / MongoDB Collections

```
Collection: watch_history
Document shape:
{
  "_id": ObjectId,
  "user_id": "uuid-string",
  "video_id": "uuid-string",
  "watched_at": ISODate,
  "progress_seconds": 142,
  "completed": false
}

Collection: user_preferences
Document shape:
{
  "_id": ObjectId,
  "user_id": "uuid-string",
  "theme": "dark",
  "autoplay": true,
  "quality_preference": "auto",
  "updated_at": ISODate
}

Collection: video_events
Document shape:
{
  "_id": ObjectId,
  "video_id": "uuid-string",
  "user_id": "uuid-string",
  "event_type": "play" | "pause" | "seek" | "end",
  "timestamp": ISODate,
  "session_id": "uuid-string",
  "position_seconds": 30
}
```

### Redis Key Patterns

```
session:{user_id}              → JSON string of session data, TTL 86400 (24h)
ratelimit:{ip}:{endpoint}      → Integer counter, TTL 60
video:meta:{video_id}          → JSON string of video metadata, TTL 300 (5m)
trending:videos                → Sorted set, score = view count, TTL 900 (15m)
```

---

## API CONTRACT

Base URL: `http://localhost:8080` (local) / `https://api.streamforge.internal` (AKS)

All protected routes require `Authorization: Bearer <jwt_token>` header.

### Auth Routes

```
POST /api/v1/auth/register
Body: { "email": string, "password": string }
Response 201: { "user_id": string, "email": string }

POST /api/v1/auth/login
Body: { "email": string, "password": string }
Response 200: { "access_token": string, "refresh_token": string, "expires_in": 3600 }

POST /api/v1/auth/refresh
Body: { "refresh_token": string }
Response 200: { "access_token": string, "expires_in": 3600 }

POST /api/v1/auth/logout    [protected]
Response 200: { "message": "logged out" }
```

### Video Routes

```
GET /api/v1/videos           [protected]
Query: ?page=1&limit=20&status=ready
Response 200: { "videos": [...], "total": int, "page": int }

GET /api/v1/videos/:id       [protected]
Response 200: { "id", "title", "description", "status", "renditions": [...], "created_at" }

POST /api/v1/videos          [protected]
Content-Type: multipart/form-data
Fields: title (string), description (string), file (video file)
Response 201: { "video_id": string, "status": "pending" }
Side effect: uploads raw file to Azure Blob, publishes transcode job to RabbitMQ

DELETE /api/v1/videos/:id    [protected]
Response 204: (no body)

GET /api/v1/videos/:id/stream  [protected]
Response 200: { "master_playlist_url": string }
Note: returns Azure Blob SAS URL for the HLS master playlist
```

### Watch History Routes

```
POST /api/v1/watch/:video_id/event   [protected]
Body: { "event_type": "play"|"pause"|"seek"|"end", "position_seconds": int }
Response 201: (writes to Cosmos DB)

GET /api/v1/watch/history    [protected]
Response 200: { "history": [{ "video_id", "title", "watched_at", "progress_seconds" }] }
```

### Health

```
GET /healthz
Response 200: { "status": "ok", "db": "ok", "redis": "ok", "rabbitmq": "ok" }
```

---

## VIDEO PIPELINE — TRANSCODER LOGIC

The transcoder worker is a **long-running Go process** that:

1. Connects to RabbitMQ on startup and begins consuming from queue `transcoder.jobs`
2. For each message received:

```json
{
  "video_id": "uuid-string",
  "raw_blob_url": "https://storageaccount.blob.core.windows.net/raw/video-uuid.mp4",
  "user_id": "uuid-string"
}
```

3. Downloads raw file from Azure Blob to a temp directory (`/tmp/{video_id}/input.mp4`)
4. Updates video status in PostgreSQL: `status = 'processing'`
5. Runs FFmpeg to produce HLS at 4 resolutions:

```bash
# Master playlist + 4 renditions
ffmpeg -i /tmp/{video_id}/input.mp4 \
  -filter_complex \
    "[v:0]split=4[v1][v2][v3][v4]" \
  -map "[v1]" -map a:0 -c:v libx264 -b:v 5000k -s 1920x1080 \
    -c:a aac -b:a 128k \
    -hls_time 6 -hls_playlist_type vod \
    -hls_segment_filename /tmp/{video_id}/1080p_%03d.ts \
    /tmp/{video_id}/1080p.m3u8 \
  -map "[v2]" -map a:0 -c:v libx264 -b:v 2800k -s 1280x720 \
    -c:a aac -b:a 128k \
    -hls_time 6 -hls_playlist_type vod \
    -hls_segment_filename /tmp/{video_id}/720p_%03d.ts \
    /tmp/{video_id}/720p.m3u8 \
  -map "[v3]" -map a:0 -c:v libx264 -b:v 1400k -s 854x480 \
    -c:a aac -b:a 96k \
    -hls_time 6 -hls_playlist_type vod \
    -hls_segment_filename /tmp/{video_id}/480p_%03d.ts \
    /tmp/{video_id}/480p.m3u8 \
  -map "[v4]" -map a:0 -c:v libx264 -b:v 700k -s 640x360 \
    -c:a aac -b:a 64k \
    -hls_time 6 -hls_playlist_type vod \
    -hls_segment_filename /tmp/{video_id}/360p_%03d.ts \
    /tmp/{video_id}/360p.m3u8
```

6. Generates `master.m3u8` combining all renditions:

```m3u8
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=5128000,RESOLUTION=1920x1080
1080p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2928000,RESOLUTION=1280x720
720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1496000,RESOLUTION=854x480
480p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=764000,RESOLUTION=640x360
360p.m3u8
```

7. Uploads all `.m3u8` and `.ts` files to Azure Blob container `hls/{video_id}/`
8. Inserts rows into `hls_renditions` table in PostgreSQL
9. Updates video status: `status = 'ready'`
10. Acknowledges the RabbitMQ message (`ack`)
11. Cleans up temp directory
12. On any error: updates status to `'failed'`, `nack` the message (do not requeue — send to dead-letter queue)

---

## ENVIRONMENT VARIABLES

All services read config from environment variables. These are injected at runtime by the infrastructure layer (from Azure Key Vault). When developing locally, use a `.env` file (never commit it).

### Go API (`backend/api`)

```env
# Server
PORT=8080
ENV=development

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=streamforge
POSTGRES_PASSWORD=<secret>
POSTGRES_DB=streamforge
POSTGRES_SSL_MODE=disable

# Cosmos DB (Mongo connection string)
COSMOS_CONNECTION_STRING=mongodb://...
COSMOS_DATABASE=streamforge

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=<secret>

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Azure Blob Storage
AZURE_STORAGE_ACCOUNT_NAME=<name>
AZURE_STORAGE_ACCOUNT_KEY=<secret>
AZURE_BLOB_RAW_CONTAINER=raw
AZURE_BLOB_HLS_CONTAINER=hls

# JWT
JWT_SECRET=<secret>
JWT_EXPIRY_HOURS=1
REFRESH_TOKEN_EXPIRY_DAYS=7

# CORS
CORS_ALLOWED_ORIGIN=http://localhost:3000
```

### Go Transcoder (`backend/transcoder`)

```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=streamforge
POSTGRES_PASSWORD=<secret>
POSTGRES_DB=streamforge
POSTGRES_SSL_MODE=disable

RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_QUEUE=transcoder.jobs
RABBITMQ_DEAD_LETTER_QUEUE=transcoder.jobs.dlq

AZURE_STORAGE_ACCOUNT_NAME=<name>
AZURE_STORAGE_ACCOUNT_KEY=<secret>
AZURE_BLOB_RAW_CONTAINER=raw
AZURE_BLOB_HLS_CONTAINER=hls

FFMPEG_PATH=/usr/bin/ffmpeg
TEMP_DIR=/tmp
```

### Next.js Frontend (`frontend`)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
JWT_COOKIE_NAME=sf_access_token
```

---

## AUTHENTICATION FLOW

**Implementation pattern: stateless JWT with Redis-backed refresh token invalidation.**

### Access Token
- Algorithm: `HS256`
- Expiry: `1 hour`
- Claims: `{ sub: user_id, email, role, exp, iat }`
- Sent as: `Authorization: Bearer <token>` header from frontend

### Refresh Token
- Random UUID, stored as SHA-256 hash in `refresh_tokens` PostgreSQL table
- Expiry: `7 days`
- Sent as HTTP-only cookie: `sf_refresh_token`

### Auth Middleware (Go)
```
1. Read Authorization header
2. Parse and validate JWT signature using JWT_SECRET
3. Check exp claim — reject if expired
4. Extract user_id from sub claim
5. Check Redis for session:{user_id} — if revoked (logout), reject
6. Attach user to request context
7. Pass to handler
```

### Next.js Middleware
```
1. Read sf_access_token cookie
2. If missing → redirect to /login
3. If present → allow request to continue
4. Token validation happens server-side in Go API (not in Next.js)
```

---

## FRONTEND PAGES & COMPONENTS

### Pages

| Route | File | Description |
|---|---|---|
| `/login` | `app/(auth)/login/page.tsx` | Login form, calls `POST /auth/login`, stores tokens in cookies |
| `/register` | `app/(auth)/register/page.tsx` | Registration form, calls `POST /auth/register` |
| `/dashboard` | `app/dashboard/page.tsx` | Lists user's videos with status badges, links to watch/upload |
| `/upload` | `app/upload/page.tsx` | Drag-and-drop upload form with progress bar |
| `/watch/[id]` | `app/watch/[id]/page.tsx` | Full video player page with Video.js |

### Key Components

#### `VideoPlayer.tsx`
- Wraps Video.js with HLS support
- Accepts `src` (master playlist URL as prop)
- On mount: initialise Video.js with `{ sources: [{ src, type: 'application/x-mpegURL' }] }`
- Emit `play`, `pause`, `ended` events → call `POST /watch/:id/event` via API client
- Must clean up Video.js instance in `useEffect` return

```tsx
// Usage
<VideoPlayer
  videoId="uuid"
  src="https://blob.../hls/uuid/master.m3u8"
/>
```

#### `UploadForm.tsx`
- Drag-and-drop file input (accept `video/*`)
- Show file name + size preview
- On submit: `POST /api/v1/videos` as `multipart/form-data`
- Show progress bar using `XMLHttpRequest` with `onprogress` (fetch doesn't support upload progress)
- After upload success: redirect to dashboard with toast notification

#### `VideoCard.tsx`
- Props: `{ id, title, status, created_at, duration_seconds }`
- Show status badge: `pending` (gray), `processing` (yellow spinner), `ready` (green), `failed` (red)
- If status is `ready`: link to `/watch/{id}`
- If status is `processing`: auto-refresh every 5 seconds using `setInterval`

### API Client (`lib/api.ts`)
```typescript
// All API calls go through this client
// Automatically attaches Authorization header from cookie
// Handles 401 → attempts token refresh → retries once → redirects to login

const apiClient = {
  get: (path: string) => fetchWithAuth(path, 'GET'),
  post: (path: string, body: unknown) => fetchWithAuth(path, 'POST', body),
  delete: (path: string) => fetchWithAuth(path, 'DELETE'),
}
```

---

## GO CODE CONVENTIONS

Follow these patterns consistently across all Go code.

### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("fetchVideo: failed to query postgres: %w", err)
}
```

### HTTP Handlers
```go
// Handlers receive (w http.ResponseWriter, r *http.Request)
// Use chi for routing
// Return JSON always
// Use a shared respond() helper

func respond(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respond(w, status, map[string]string{"error": message})
}
```

### Context Usage
```go
// User ID is attached to request context by auth middleware
type contextKey string
const userIDKey contextKey = "userID"

// In middleware:
ctx := context.WithValue(r.Context(), userIDKey, userID)

// In handler:
userID := r.Context().Value(userIDKey).(string)
```

### Package Structure Rule
- `handlers/` — only HTTP logic, call into service layer
- `internal/` — all business logic (no direct handler-to-db calls)
- `models/` — structs only, no methods
- Keep `main.go` minimal: wire dependencies, start server

---

## TYPESCRIPT CONVENTIONS

### Shared Types (`lib/types.ts`)
```typescript
export type VideoStatus = 'pending' | 'processing' | 'ready' | 'failed'

export interface Video {
  id: string
  title: string
  description: string | null
  status: VideoStatus
  duration_seconds: number | null
  created_at: string
  renditions?: HLSRendition[]
}

export interface HLSRendition {
  resolution: '360p' | '480p' | '720p' | '1080p'
  playlist_url: string
  bandwidth: number
}

export interface User {
  id: string
  email: string
  role: 'user' | 'admin'
}

export interface ApiError {
  error: string
}
```

### Server Components vs Client Components
- Pages that fetch data server-side: use `async` Server Components, call API with server-side fetch (include cookie from `cookies()`)
- Interactive components (forms, video player, status polling): mark with `'use client'`
- Keep `'use client'` boundary as low as possible

---

## RABBITMQ CONFIGURATION

Queue setup (created by Go API on startup if not exists):

```
Exchange:   transcoder (type: direct, durable: true)
Queue:      transcoder.jobs (durable: true, x-dead-letter-exchange: transcoder.dlx)
Binding:    transcoder.jobs → transcoder exchange, routing key: transcode

DLX Exchange: transcoder.dlx (type: direct, durable: true)
DLQ Queue:    transcoder.jobs.dlq (durable: true)
```

Message format published by Go API:
```json
{
  "video_id": "550e8400-e29b-41d4-a716-446655440000",
  "raw_blob_url": "https://{account}.blob.core.windows.net/raw/550e8400.mp4",
  "user_id": "user-uuid",
  "uploaded_at": "2026-03-06T10:00:00Z"
}
```

Consumer prefetch: `1` (process one job at a time per worker instance)

---

## AZURE BLOB STORAGE — URL PATTERNS

```
Raw uploads container:  raw/
  URL pattern: https://{account}.blob.core.windows.net/raw/{video_id}.mp4

HLS output container:   hls/
  URL pattern: https://{account}.blob.core.windows.net/hls/{video_id}/master.m3u8
               https://{account}.blob.core.windows.net/hls/{video_id}/1080p.m3u8
               https://{account}.blob.core.windows.net/hls/{video_id}/1080p_000.ts
```

When serving HLS to the frontend, generate a **SAS (Shared Access Signature) URL** with 4-hour expiry rather than exposing raw blob URLs. The Go API generates and returns this SAS URL on `GET /videos/:id/stream`.

---

## WHAT TO BUILD — PRIORITY ORDER

When generating code, follow this priority sequence. Do not skip ahead.

### P0 — Must Have (Core Functionality)
- [ ] PostgreSQL schema migrations (all 4 SQL files)
- [ ] Go API: user registration + login with JWT
- [ ] Go API: video upload endpoint (multipart → Azure Blob → RabbitMQ publish)
- [ ] Go API: video list + get by ID endpoint
- [ ] Go Transcoder: RabbitMQ consumer + FFmpeg execution + Blob upload
- [ ] Go API: `/videos/:id/stream` SAS URL endpoint
- [ ] Next.js: login and register pages
- [ ] Next.js: dashboard page (list videos)
- [ ] Next.js: upload page with progress bar
- [ ] Next.js: watch page with Video.js HLS player
- [ ] Auth middleware (Go) + Next.js middleware

### P1 — Should Have
- [ ] Go API: watch history endpoints (Cosmos DB read/write)
- [ ] Redis session cache + rate limiting middleware
- [ ] VideoCard status polling (auto-refresh when status = processing)
- [ ] JWT refresh token flow
- [ ] `/healthz` endpoint with all dependency checks
- [ ] Go API: video delete endpoint (cascade Blob + DB)

### P2 — Nice to Have
- [ ] User preferences (Cosmos DB)
- [ ] Video event tracking (play/pause/seek events to Cosmos DB)
- [ ] Trending videos from Redis sorted set
- [ ] Admin role: view all videos from all users
- [ ] Search videos by title (PostgreSQL ILIKE)

---

## CONSTRAINTS & GUARDRAILS

When generating code, strictly follow these rules:

1. **No infrastructure code** — Do not generate Terraform, Kubernetes manifests, Dockerfiles, or GitHub Actions YAML unless explicitly asked.
2. **No mocking Azure in app code** — The app reads real environment variables. For local dev, variables are set via `.env`. No `if os.Getenv("ENV") == "local"` Azure mocks in application code.
3. **pgx/v5 only** — Do not use `database/sql` + `lib/pq` for PostgreSQL. Use `pgx/v5` directly.
4. **chi router** — Do not use Gin, Fiber, Echo, or any other Go HTTP framework. Use `chi`.
5. **No ORM** — Do not use GORM or any ORM. Write raw SQL queries with `pgx`.
6. **No global state in Go** — Pass dependencies (db, redis, etc.) through constructor injection, not global variables.
7. **TypeScript strict mode** — No `any` types. No `@ts-ignore`. Properly type all API responses.
8. **No `useEffect` for data fetching** — Use Next.js Server Components for initial data loads. `useEffect` is for side effects only (e.g., Video.js init, polling).
9. **Error boundaries** — All pages must handle loading and error states.
10. **Never log secrets** — No `log.Printf("password: %s", password)` or similar.

---

## EXAMPLE: HOW TO USE THIS FILE

When prompting an LLM, prepend or include this file and then ask specific questions:

```
# Example prompts that work well with this context:

"Using the API contract and Go conventions in this README, generate the 
 video upload handler in backend/api/internal/handlers/videos.go"

"Generate the VideoPlayer.tsx component using the spec in the Frontend 
 Components section, including the watch event emission logic"

"Write the RabbitMQ consumer in backend/transcoder/internal/consumer/rabbitmq.go 
 following the pipeline logic described in the VIDEO PIPELINE section"

"Generate the PostgreSQL migration files as defined in the DATABASE SCHEMAS section"

"Create the Next.js dashboard page that fetches and displays the user's videos 
 with status badges and auto-refresh for processing videos"
```

---

*StreamForge — LLM Development Context | Generated March 2026 | Application layer only*
