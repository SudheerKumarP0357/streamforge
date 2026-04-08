# StreamForge

A cloud-native video streaming platform built on **Azure**, following microservices architecture with full Infrastructure-as-Code, CI/CD automation, and Kubernetes-first deployment.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
- [Repository Structure](#repository-structure)
- [Infrastructure (Terraform)](#infrastructure-terraform)
- [Kubernetes Deployment](#kubernetes-deployment)
- [CI/CD Pipelines](#cicd-pipelines)
- [Containerization](#containerization)
- [Environment Variables Reference](#environment-variables-reference)
- [Local Development (Docker Compose)](#local-development-docker-compose)
- [Running Services Individually](#running-services-individually)
- [Secrets Management](#secrets-management)
- [Network Policies & Security](#network-policies--security)
- [Health Checks & Probes](#health-checks--probes)
- [Observability](#observability)
- [Required External Dependencies](#required-external-dependencies)

---

## Architecture Overview

StreamForge follows a **microservices architecture** deployed on **Azure Kubernetes Service (AKS)**. The platform ingests raw video uploads, transcodes them to HLS format via an asynchronous worker pipeline, and serves them through a server-side-rendered Next.js frontend.

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│   Frontend  │────▶│  API Server │────▶│   PostgreSQL    │
│  (Next.js)  │     │   (Go 1.24)   │     │   (v17+)        │
│  Port: 3000 │     │  Port: 8080   │     │  Port: 5432     │
└─────────────┘     └──────┬───────┘     └────────────────┘
                           │
                    ┌──────┴───────┐
                    │              │
              ┌─────▼─────┐  ┌────▼─────────┐
              │  RabbitMQ  │  │  Redis Cache  │
              │ Port: 5672 │  │  Port: 6379   │
              └─────┬──────┘  └──────────────┘
                    │
              ┌─────▼──────────┐
              │  Transcoder     │     ┌──────────────────┐
              │  Worker (Go)    │────▶│  Azure Blob       │
              │  Port: 9090     │     │  Storage (HLS)    │
              └────────────────┘     └──────────────────┘
                                           │
                                     ┌─────▼──────────┐
                                     │  Cosmos DB       │
                                     │  (MongoDB API)   │
                                     └────────────────┘
```

### Service Communication Flow

1. **Frontend** → API Server (REST over HTTP)
2. **API Server** → RabbitMQ (publishes transcode jobs)
3. **Transcoder Worker** ← RabbitMQ (consumes jobs from `transcoder.jobs` queue)
4. **Transcoder** → Azure Blob Storage (uploads HLS segments)
5. **API Server** → PostgreSQL (users, content, subscriptions)
6. **API Server** → Cosmos DB (watch history, preferences)
7. **API Server** → Redis (session cache, rate limiting)

---

## Technology Stack

| Layer | Technology | Version | Azure Service |
|-------|-----------|---------|---------------|
| **Frontend** | Next.js + React | Node 24.11.0 | AKS Pod |
| **API Gateway** | Go | 1.24.0 | AKS Pod |
| **Transcoder Worker** | Go + FFmpeg | 1.24.0 / 8.0.1 | AKS Pod (async) |
| **Message Queue** | RabbitMQ | 4.2.4 | AKS StatefulSet |
| **Relational DB** | PostgreSQL | 17+ | Azure DB for PostgreSQL Flexible Server |
| **Document DB** | MongoDB API | — | Azure Cosmos DB |
| **Cache** | Redis | Alpine | Azure Cache for Redis |
| **Object Storage** | Blob Storage | — | Azure Storage Account |
| **Secrets** | Key Vault + CSI Driver | — | Azure Key Vault |
| **Container Registry** | ACR / GHCR | — | Azure Container Registry |
| **IaC** | Terraform | azurerm 4.65.0 | Azure Resource Manager |
| **CI/CD** | GitHub Actions | — | GitHub-hosted runners |
| **Networking** | Calico CNI (Overlay) | — | AKS Network Profile |
| **Observability** | Prometheus + Grafana | — | Azure Managed Grafana |

---

## Repository Structure

```
streamforge/
├── .github/
│   ├── actions/
│   │   └── docker-build-push/       # Reusable composite action for Docker builds
│   └── workflows/
│       ├── api-ci.yml                # Backend API build + push
│       ├── frontend-ci.yml           # Frontend build + push
│       ├── transcoder-ci.yml         # Transcoder build + push
│       └── infra-plan.yml            # Terraform plan/apply pipeline
├── backend/
│   ├── api/                          # Go API service
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   ├── migrations/
│   │   └── Dockerfile
│   └── transcoder/                   # Go transcoder worker
│       ├── cmd/worker/main.go
│       ├── internal/
│       └── Dockerfile
├── frontend/                         # Next.js SSR application
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── Dockerfile
├── infra/
│   └── terraform/                    # All Azure infrastructure
│       ├── main.tf                   # Resource group
│       ├── aks.tf                    # AKS cluster + identities
│       ├── postgres.tf               # PostgreSQL Flexible Server
│       ├── cosmos.tf                 # Cosmos DB (MongoDB)
│       ├── redis.tf                  # Azure Cache for Redis
│       ├── acr.tf                    # Azure Container Registry
│       ├── keyvault.tf               # Key Vault + secrets
│       ├── network.tf                # VNets
│       ├── subnets.tf                # Subnet definitions
│       ├── monitor.tf                # Azure Monitor / Grafana
│       ├── jumpserver.tf             # Bastion jump server
│       ├── variables.tf
│       ├── terraform.tfvars
│       ├── outputs.tf
│       └── versions.tf
├── k8s/
│   ├── raw-manifests/                # Plain YAML manifests (local/dev)
│   │   ├── 01-namespace.yml
│   │   ├── 02-storage.yml
│   │   ├── 03-configmaps.yml
│   │   ├── 04-secrets.yml
│   │   ├── 05-postgres.yml
│   │   ├── 06-redis.yml
│   │   ├── 07-rabbitmq.yml
│   │   ├── 08-api.yml
│   │   ├── 09-transcoder.yml
│   │   ├── 10-frontend.yml
│   │   └── 11-network-policy.yml
│   ├── azure-raw/                    # Azure-specific manifests (Key Vault CSI, workload identity)
│   │   ├── 01-streamforge-ns.yml
│   │   ├── 02-secret-provider-class.yml
│   │   ├── 02-sf-workload-sa.yml
│   │   ├── 03-sc-pvc-rabbitmq.yml
│   │   ├── 04-sf-configmaps.yml
│   │   ├── 05-sf-rabbitmq-*.yml
│   │   ├── 06-sf-api-*.yml
│   │   ├── 07-sf-transcoder-*.yml
│   │   ├── 08-sf-frontend-*.yml
│   │   └── 09-network-policy.yml
│   ├── kustomize/                    # Kustomize overlays
│   │   ├── base/                     # Base resources
│   │   └── envs/
│   │       ├── dev/
│   │       └── prod/
│   └── helm/                         # Helm charts (WIP)
├── docs/
├── docker-compose.yml                # Full local dev stack
└── README.md
```

---

## Infrastructure (Terraform)

All Azure infrastructure is provisioned via Terraform located in `infra/terraform/`.

### Azure Resources Provisioned

| Resource | Terraform File | Key Details |
|----------|---------------|-------------|
| Resource Group | `main.tf` | Naming: `rg-{app}-{env}-{region}` |
| AKS Cluster | `aks.tf` | Private cluster, Calico network policy, Workload Identity, auto-scaling, auto-upgrade (patch) |
| PostgreSQL | `postgres.tf` | Flexible Server v18, `B_Standard_B1ms`, pgcrypto + uuid-ossp extensions |
| Cosmos DB | `cosmos.tf` | MongoDB API, Free tier |
| Redis | `redis.tf` | `Balanced_B0` SKU |
| ACR | `acr.tf` | Basic SKU |
| Key Vault | `keyvault.tf` | Stores all app secrets, CSI driver integration |
| VNets & Subnets | `network.tf`, `subnets.tf` | App VNet `10.10.0.0/16`, Jump VNet `10.9.0.0/16` |
| Jump Server | `jumpserver.tf` | Bastion VM for private cluster access |
| Monitoring | `monitor.tf` | Azure Managed Grafana + Prometheus |
| Private Endpoints | `private-endpoint/` | Service-specific private links |

### AKS Cluster Configuration

| Setting | Value |
|---------|-------|
| Private Cluster | `true` |
| RBAC | Enabled |
| Workload Identity | Enabled |
| OIDC Issuer | Enabled |
| Network Plugin | Azure CNI (Overlay mode) |
| Network Policy | Calico |
| Service CIDR | `10.40.0.0/16` |
| DNS Service IP | `10.40.0.10` |
| Node Pool VM Size | `Standard_D2as_v4` |
| OS | Ubuntu |
| OS Disk | 64 GB Managed |
| Auto-scaling | 1–1 nodes (configurable) |
| Max Pods per Node | 250 |
| Auto-upgrade | Patch (Weekly, Sunday) |
| Image Cleaner | Enabled (168h interval) |
| Key Vault CSI | Enabled (2m rotation) |

### Terraform State Backend

State is stored remotely in Azure Blob Storage. Backend config is injected at `terraform init` time via CI/CD secrets:

```bash
terraform init \
  -backend-config="resource_group_name=$BACKEND_RESOURCE_GROUP_NAME" \
  -backend-config="storage_account_name=$BACKEND_STORAGE_ACCOUNT_NAME" \
  -backend-config="container_name=$BACKEND_STORAGE_CONTAINER_NAME" \
  -backend-config="key=$BACKEND_STATE_KEY"
```

### Terraform Providers

| Provider | Version |
|----------|---------|
| `hashicorp/azurerm` | 4.65.0 |
| `hashicorp/azuread` | 3.8.0 |
| `hashicorp/http` | 3.5.0 |
| `hashicorp/tls` | 4.2.1 |

### Running Terraform Locally

```bash
cd infra/terraform

# Initialize with backend
terraform init \
  -backend-config="resource_group_name=<RG_NAME>" \
  -backend-config="storage_account_name=<STORAGE_NAME>" \
  -backend-config="container_name=<CONTAINER>" \
  -backend-config="key=<STATE_KEY>"

# Plan
terraform plan \
  -var="postgres_administrator_password=<PASSWORD>" \
  -var="cosmos_administrator_password=<PASSWORD>" \
  -var="jwt_secret=<SECRET>" \
  -var="rabbitmq_default_pass=<PASSWORD>"

# Apply
terraform apply main.tfplan
```

---

## Kubernetes Deployment

### Deployment Strategies

The project supports **three deployment strategies**:

| Strategy | Path | Use Case |
|----------|------|----------|
| **Raw Manifests** | `k8s/raw-manifests/` | Local/dev — includes in-cluster PostgreSQL, Redis, RabbitMQ |
| **Azure Raw Manifests** | `k8s/azure-raw/` | Azure AKS — uses Azure-managed services, Key Vault CSI |
| **Kustomize** | `k8s/kustomize/` | Environment overlays for dev/prod |

### Namespace

All workloads deploy into the `streamforge` namespace.

### Applying Raw Manifests (Local/Dev)

```bash
# Apply in order (numbered for dependency resolution)
kubectl apply -f k8s/raw-manifests/01-namespace.yml
kubectl apply -f k8s/raw-manifests/02-storage.yml
kubectl apply -f k8s/raw-manifests/03-configmaps.yml
kubectl apply -f k8s/raw-manifests/04-secrets.yml
kubectl apply -f k8s/raw-manifests/05-postgres.yml
kubectl apply -f k8s/raw-manifests/06-redis.yml
kubectl apply -f k8s/raw-manifests/07-rabbitmq.yml
kubectl apply -f k8s/raw-manifests/08-api.yml
kubectl apply -f k8s/raw-manifests/09-transcoder.yml
kubectl apply -f k8s/raw-manifests/10-frontend.yml
kubectl apply -f k8s/raw-manifests/11-network-policy.yml
```

### Applying Azure Manifests (AKS Production)

```bash
# Apply in order
kubectl apply -f k8s/azure-raw/ 
```

> **Note:** Azure manifests use `SecretProviderClass` to sync secrets from Azure Key Vault via the CSI Secrets Store Driver. Ensure the Workload Identity service account (`sf-workload-sa`) is properly federated.

### Kustomize Overlays

```bash
# Dev environment
kubectl apply -k k8s/kustomize/envs/dev/

# Production environment
kubectl apply -k k8s/kustomize/envs/prod/
```

### Resource Limits

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---------|------------|-----------|----------------|--------------|
| API | 250m | 500m | 64Mi | 128Mi |
| Transcoder | — | — | — | — |
| Frontend | — | — | — | — |

### Init Containers

The API deployment uses **init containers** to wait for dependent services before starting:

- `wait-for-postgres` — polls `streamforge-postgres-svc:5432`
- `wait-for-redis` — polls `streamforge-redis-svc:6379`
- `wait-for-rabbitmq` — polls `streamforge-rabbitmq-svc:5672`

---

## CI/CD Pipelines

### GitHub Actions Workflows

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| Backend API CI | `api-ci.yml` | Push/PR to `main` on `backend/api/**` | Build, vet, Docker build & push |
| Frontend CI | `frontend-ci.yml` | Push/PR to `main` on `frontend/**` | Build, Docker build & push |
| Transcoder CI | `transcoder-ci.yml` | Push/PR to `main` on `backend/transcoder/**` | Build, Docker build & push |
| Infra Plan/Apply | `infra-plan.yml` | Push/PR to `infra/**` on `infra/terraform/**` | Terraform fmt → init → validate → tfsec → plan → apply |

### CI Pipeline Flow (Application Services)

```
PR → Build & Vet → (merge to main) → Docker Build → Push to GHCR
```

- **Registry**: `ghcr.io` (GitHub Container Registry)
- **Image naming**: `ghcr.io/<org>/streamforge/<service>`
- **Build action**: Uses reusable composite action at `.github/actions/docker-build-push/`
- **Auth**: OIDC-based (`id-token: write` permission)

### Infra Pipeline Flow

```
PR to infra/* ──▶ fmt ──▶ init ──▶ validate ──▶ tfsec ──▶ plan ──▶ PR Comment
                                                                        │
Merge to infra/* ──────────────────────────────────────────────▶ apply ◀─┘
```

- **Environments**: Branch-based (`infra/development` → dev, `infra/production` → prod)
- **Security**: Terraform plan posted as PR comment, `tfsec` security scanning
- **Auth**: Azure OIDC (`ARM_USE_OIDC: true`, no client secret)
- **Artifacts**: Plan file uploaded as GitHub artifact for audit trail

### Required GitHub Secrets

| Secret | Used By | Description |
|--------|---------|-------------|
| `ARM_CLIENT_ID` | Infra | Azure Service Principal Client ID |
| `ARM_SUBSCRIPTION_ID` | Infra | Azure Subscription ID |
| `ARM_TENANT_ID` | Infra | Azure Tenant ID |
| `BACKEND_RESOURCE_GROUP_NAME` | Infra | Terraform state backend RG |
| `BACKEND_STORAGE_ACCOUNT_NAME` | Infra | Terraform state storage account |
| `BACKEND_STORAGE_CONTAINER_NAME` | Infra | Terraform state container |
| `BACKEND_STATE_KEY` | Infra | Terraform state file key |
| `POSTGRES_PASSWORD` | Infra | PostgreSQL admin password |
| `COSMOS_PASSWORD` | Infra | Cosmos DB admin password |
| `JWT_SECRET` | Infra | Application JWT signing key |
| `RABBITMQ_PASSWORD` | Infra | RabbitMQ password |
| `GITHUB_TOKEN` | All CI | Auto-provided, used for GHCR push |

---

## Containerization

### Docker Images

| Service | Dockerfile | Base (Build) | Base (Runtime) | Exposed Port | User |
|---------|-----------|-------------|----------------|-------------|------|
| **Frontend** | `frontend/Dockerfile` | `node:24.11.0-alpine` | `node:24.11.0-alpine` | 3000 | `nextjs` (UID 1001) |
| **API** | `backend/api/Dockerfile` | `golang:1.24.0` | `alpine:3.23` | 8080 | `appuser` (UID 10001) |
| **Transcoder** | `backend/transcoder/Dockerfile` | `golang:1.24.0` | `alpine:3.23` + FFmpeg 8.0.1 | 9090 | `appuser` (UID 10001) |

### Build Features

- **Multi-stage builds** — separate build and runtime stages for minimal image size
- **Build cache mounts** — `--mount=type=cache` for Go modules and npm packages
- **Non-root users** — all containers run as non-root
- **CGO disabled** — static Go binaries (`CGO_ENABLED=0`)
- **Next.js standalone** — output mode for minimal production bundle

### Building Images Locally

```bash
# Frontend
docker build -t streamforge-frontend:local ./frontend

# API
docker build -t streamforge-api:local ./backend/api

# Transcoder
docker build -t streamforge-transcoder:local ./backend/transcoder
```

---

## Environment Variables Reference

### Frontend

| Variable | Description | Example |
|----------|-------------|---------|
| `API_URL` | Backend API endpoint | `http://api:8080` |
| `APP_ENV` | `production` or `development` — controls log verbosity | `development` |

### API Server

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL hostname | `postgres` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database user | `postgres` |
| `POSTGRES_PASSWORD` | Database password | — |
| `POSTGRES_DB` | Database name | `streamforge` |
| `POSTGRES_SSL_MODE` | SSL mode (`disable` / `require`) | `disable` |
| `REDIS_URL` | Redis connection URL | `redis://redis:6379` |
| `REDIS_PASSWORD` | Redis password (empty if none) | — |
| `RABBITMQ_URL` | RabbitMQ AMQP URL | `amqp://guest:guest@rabbitmq:5672/` |
| `AZURE_STORAGE_ACCOUNT_NAME` | Azure Storage account name | — |
| `AZURE_STORAGE_ACCOUNT_KEY` | Azure Storage access key | — |
| `AZURE_STORAGE_ENDPOINT` | Blob endpoint URL | `https://<name>.blob.core.windows.net` |
| `AZURE_BLOB_RAW_CONTAINER` | Container for raw uploads | `raw-videos` |
| `AZURE_BLOB_HLS_CONTAINER` | Container for HLS output | `hls-videos` |
| `COSMOS_CONNECTION_STRING` | MongoDB-compatible connection string | `mongodb://...` |
| `COSMOS_DATABASE` | Cosmos DB database name | `streamforge` |
| `JWT_SECRET` | JWT signing secret | — |
| `JWT_EXPIRY_HOURS` | JWT token TTL (hours) | `24` |
| `REFRESH_TOKEN_EXPIRY_DAYS` | Refresh token TTL (days) | `7` |
| `CORS_ALLOWED_ORIGIN` | Allowed CORS origin | `http://localhost:3000` |

### Transcoder Worker

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL hostname | `postgres` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database user | `postgres` |
| `POSTGRES_PASSWORD` | Database password | — |
| `POSTGRES_DB` | Database name | `streamforge` |
| `POSTGRES_SSL_MODE` | SSL mode | `disable` |
| `RABBITMQ_URL` | RabbitMQ AMQP URL | `amqp://guest:guest@rabbitmq:5672/` |
| `RABBITMQ_QUEUE` | Queue name to consume from | `transcoder.jobs` |
| `RABBITMQ_DEAD_LETTER_QUEUE` | DLQ for failed jobs | `transcoder.jobs.dlq` |
| `AZURE_STORAGE_ACCOUNT_NAME` | Azure Storage account name | — |
| `AZURE_STORAGE_ACCOUNT_KEY` | Azure Storage access key | — |
| `AZURE_STORAGE_ENDPOINT` | Blob endpoint URL | — |
| `AZURE_BLOB_RAW_CONTAINER` | Container for raw uploads | `raw-videos` |
| `AZURE_BLOB_HLS_CONTAINER` | Container for HLS output | `hls-videos` |
| `FFMPEG_PATH` | Path to FFmpeg binary | `/usr/bin/ffmpeg` |
| `TEMP_DIR` | Temp directory for transcoding | `/tmp` |

---

## Local Development (Docker Compose)

The `docker-compose.yml` spins up the **entire stack** locally, including all infrastructure dependencies.

### Services

| Service | Image | Port Mapping | Healthcheck |
|---------|-------|-------------|-------------|
| `frontend` | Built from `./frontend` | `3000:3000` | `wget http://localhost:3000/` |
| `api` | Built from `./backend/api` | `8080:8080` | `wget http://localhost:8080/healthz` |
| `transcoder` | Built from `./backend/transcoder` | `9090:9090` | — |
| `postgres` | `postgres:15-alpine` | `5432:5432` | `pg_isready -U postgres` |
| `redis` | `redis:alpine` | `6379:6379` | `redis-cli ping` |
| `rabbitmq` | `rabbitmq:4.2.4-alpine` | `5672:5672` | `rabbitmq-diagnostics ping` |
| `mongodb` | `mongodb/mongodb-community-server:8.0-ubi8` | `27017:27017` | `mongosh --quiet --eval` |

### Quick Start

```bash
# Set required environment variables
export AZURE_STORAGE_ACCOUNT_NAME=<your_account>
export AZURE_STORAGE_ACCOUNT_KEY=<your_key>
export JWT_SECRET=<your_secret>

# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Tear down
docker compose down -v
```

### Dependency Order

Docker Compose manages startup dependencies using health checks:

```
postgres ─┐
redis ────┤
rabbitmq ─┼──▶ api ──▶ frontend
mongodb ──┘
              rabbitmq ──▶ transcoder
              postgres ──┘
```

### Persistent Volumes

| Volume | Mount Point | Service |
|--------|------------|---------|
| `postgres_data` | `/var/lib/postgresql/data` | postgres |
| `redis_data` | `/data` | redis |
| `mongodb_data` | `/data/db` | mongodb |

---

## Running Services Individually

### Frontend — Next.js (Node 24.11.0)

```bash
cd ./frontend
npm install
npm run dev          # Development server (hot reload)
npm run build        # Production build → ./frontend/build/standalone
```

**Production run after build:**

```bash
cp -r ./build/static ./build/standalone/build/static
node ./build/standalone/server.js
```

### Backend API — Go 1.24.0

```bash
cd ./backend/api
go mod download
go run ./cmd/server/main.go          # Development
CGO_ENABLED=0 go build -o /bin/server ./cmd/server/main.go  # Production build
```

### Transcoder Worker — Go 1.24.0

```bash
cd ./backend/transcoder
go mod download
go run ./cmd/worker/main.go          # Development
CGO_ENABLED=0 go build -o /bin/worker ./cmd/worker/main.go  # Production build
```

> **Note:** The transcoder requires FFmpeg to be installed on the host. Set `FFMPEG_PATH` to the FFmpeg binary location.

---

## Secrets Management

### Local / Raw Manifests

Secrets are base64-encoded in `k8s/raw-manifests/04-secrets.yml` (for local dev only).

### Azure AKS (Production)

Secrets are managed via **Azure Key Vault** with the **Secrets Store CSI Driver**:

1. Secrets stored in Azure Key Vault (`kv-streamforgedevci`)
2. `SecretProviderClass` (`k8s/azure-raw/02-secret-provider-class.yml`) maps KV secrets to K8s secrets
3. Workload Identity (`sf-workload-sa`) authenticates pods to Key Vault
4. Secrets auto-rotate every **2 minutes** via CSI driver

**Key Vault Secrets Synced:**

| Key Vault Secret Name | K8s Secret Key | Used By |
|----------------------|----------------|---------|
| `rabbitmq-url` | `RABBITMQ_URL` | API, Transcoder |
| `rabbitmq-default-pass` | `RABBITMQ_DEFAULT_PASS` | RabbitMQ |
| `azure-storage-account-key` | `AZURE_STORAGE_ACCOUNT_KEY` | API, Transcoder |
| `postgres-password` | `POSTGRES_PASSWORD` | API, Transcoder |
| `cosmos-connection-string` | `COSMOS_CONNECTION_STRING` | API |
| `redis-password` | `REDIS_PASSWORD` | API |
| `jwt-secret` | `JWT_SECRET` | API |

---

## Network Policies & Security

Kubernetes NetworkPolicies enforce **least-privilege network access** within the `streamforge` namespace:

| Policy | Target | Allowed Sources | Port |
|--------|--------|-----------------|------|
| `allow-only-api-transcoder-to-postgres` | PostgreSQL (`database` tier) | API, Transcoder (`backend` tier) | 5432 |
| `allow-only-api-transcoder-to-redis` | Redis (`cache` tier) | API, Transcoder (`backend` tier) | 6379 |
| `allow-only-api-transcoder-to-rabbitmq` | RabbitMQ (`queue` tier) | API, Transcoder (`backend` tier) | 5672 |
| `allow-only-frontend-to-api` | API (`backend` tier) | Frontend (`frontend` tier) | 8080 |

> **Note:** All backend services use standard `app.kubernetes.io/*` labels for pod selection. Network policies enforce that databases and caches are **only reachable** from authorized backend services.

---

## Health Checks & Probes

### Docker Compose Health Checks

| Service | Check | Interval | Timeout | Retries | Start Period |
|---------|-------|----------|---------|---------|-------------|
| Frontend | `wget http://localhost:3000/` | 10s | 5s | 5 | 15s |
| API | `wget http://localhost:8080/healthz` | 10s | 5s | — | 120s |
| PostgreSQL | `pg_isready -U postgres` | 10s | 5s | 5 | 20s |
| Redis | `redis-cli ping \| grep PONG` | 1s | 3s | 5 | — |
| RabbitMQ | `rabbitmq-diagnostics ping` | 10s | 5s | 5 | — |
| MongoDB | `mongosh --quiet --eval quit(...)` | 10s | 10s | 5 | 30s |

### Kubernetes Probes (API)

| Probe | Type | Endpoint | Initial Delay | Period | Timeout |
|-------|------|----------|--------------|--------|---------|
| Liveness | `exec` | `wget http://localhost:8080/healthz` | 20s | 10s | 5s |
| Readiness | `exec` | `wget http://localhost:8080/healthz` | 20s | 10s | 5s |

---

## Observability

| Component | Tool | Purpose |
|-----------|------|---------|
| Metrics | Prometheus | Scrapes application and cluster metrics |
| Dashboards | Azure Managed Grafana | Visualization and alerting |
| Infrastructure | Azure Monitor | Resource-level monitoring |

---

## Required External Dependencies

| Dependency | Minimum Version | Required Extensions / Notes |
|------------|----------------|-----------------------------|
| PostgreSQL | 17+ | `pgcrypto`, `uuid-ossp` extensions |
| Redis | — | Standard Redis instance |
| RabbitMQ | 4.x | Queues: `transcoder.jobs`, `transcoder.jobs.dlq` |
| MongoDB / Cosmos DB | — | MongoDB wire-protocol compatible |
| Azure Storage Account | — | Two containers: `raw-videos`, `hls-videos` |
| FFmpeg | 8.0.1+ | Required on transcoder host/container |
| Go | 1.24.0 | For building backend services |
| Node.js | 24.11.0 | For building frontend |
| Terraform | — | With providers listed in `versions.tf` |
| kubectl | 1.35+ | For cluster management |
# StreamForge

A cloud-native video streaming platform built on **Azure**, following microservices architecture with full Infrastructure-as-Code, CI/CD automation, and Kubernetes-first deployment.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
- [Repository Structure](#repository-structure)
- [Infrastructure (Terraform)](#infrastructure-terraform)
- [Kubernetes Deployment](#kubernetes-deployment)
- [CI/CD Pipelines](#cicd-pipelines)
- [Containerization](#containerization)
- [Environment Variables Reference](#environment-variables-reference)
- [Local Development (Docker Compose)](#local-development-docker-compose)
- [Running Services Individually](#running-services-individually)
- [Secrets Management](#secrets-management)
- [Network Policies & Security](#network-policies--security)
- [Health Checks & Probes](#health-checks--probes)
- [Observability](#observability)
- [Required External Dependencies](#required-external-dependencies)

---

## Architecture Overview

StreamForge follows a **microservices architecture** deployed on **Azure Kubernetes Service (AKS)**. The platform ingests raw video uploads, transcodes them to HLS format via an asynchronous worker pipeline, and serves them through a server-side-rendered Next.js frontend.

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│   Frontend   │────▶│   API Server  │────▶│   PostgreSQL    │
│  (Next.js)   │     │   (Go 1.24)   │     │   (v17+)        │
│  Port: 3000  │     │  Port: 8080   │     │  Port: 5432     │
└─────────────┘     └──────┬───────┘     └────────────────┘
                           │
                    ┌──────┴───────┐
                    │              │
              ┌─────▼─────┐  ┌────▼─────────┐
              │  RabbitMQ  │  │  Redis Cache  │
              │ Port: 5672 │  │  Port: 6379   │
              └─────┬──────┘  └──────────────┘
                    │
              ┌─────▼──────────┐
              │  Transcoder     │     ┌──────────────────┐
              │  Worker (Go)    │────▶│  Azure Blob       │
              │  Port: 9090     │     │  Storage (HLS)    │
              └────────────────┘     └──────────────────┘
                                           │
                                     ┌─────▼──────────┐
                                     │  Cosmos DB       │
                                     │  (MongoDB API)   │
                                     └────────────────┘
```

### Service Communication Flow

1. **Frontend** → API Server (REST over HTTP)
2. **API Server** → RabbitMQ (publishes transcode jobs)
3. **Transcoder Worker** ← RabbitMQ (consumes jobs from `transcoder.jobs` queue)
4. **Transcoder** → Azure Blob Storage (uploads HLS segments)
5. **API Server** → PostgreSQL (users, content, subscriptions)
6. **API Server** → Cosmos DB (watch history, preferences)
7. **API Server** → Redis (session cache, rate limiting)

---

## Technology Stack

| Layer | Technology | Version | Azure Service |
|-------|-----------|---------|---------------|
| **Frontend** | Next.js + React | Node 24.11.0 | AKS Pod |
| **API Gateway** | Go | 1.24.0 | AKS Pod |
| **Transcoder Worker** | Go + FFmpeg | 1.24.0 / 8.0.1 | AKS Pod (async) |
| **Message Queue** | RabbitMQ | 4.2.4 | AKS StatefulSet |
| **Relational DB** | PostgreSQL | 17+ | Azure DB for PostgreSQL Flexible Server |
| **Document DB** | MongoDB API | — | Azure Cosmos DB |
| **Cache** | Redis | Alpine | Azure Cache for Redis |
| **Object Storage** | Blob Storage | — | Azure Storage Account |
| **Secrets** | Key Vault + CSI Driver | — | Azure Key Vault |
| **Container Registry** | ACR / GHCR | — | Azure Container Registry |
| **IaC** | Terraform | azurerm 4.65.0 | Azure Resource Manager |
| **CI/CD** | GitHub Actions | — | GitHub-hosted runners |
| **Networking** | Calico CNI (Overlay) | — | AKS Network Profile |
| **Observability** | Prometheus + Grafana | — | Azure Managed Grafana |

---

## Repository Structure

```
streamforge/
├── .github/
│   ├── actions/
│   │   └── docker-build-push/       # Reusable composite action for Docker builds
│   └── workflows/
│       ├── api-ci.yml                # Backend API build + push
│       ├── frontend-ci.yml           # Frontend build + push
│       ├── transcoder-ci.yml         # Transcoder build + push
│       └── infra-plan.yml            # Terraform plan/apply pipeline
├── backend/
│   ├── api/                          # Go API service
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   ├── migrations/
│   │   └── Dockerfile
│   └── transcoder/                   # Go transcoder worker
│       ├── cmd/worker/main.go
│       ├── internal/
│       └── Dockerfile
├── frontend/                         # Next.js SSR application
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── Dockerfile
├── infra/
│   └── terraform/                    # All Azure infrastructure
│       ├── main.tf                   # Resource group
│       ├── aks.tf                    # AKS cluster + identities
│       ├── postgres.tf               # PostgreSQL Flexible Server
│       ├── cosmos.tf                 # Cosmos DB (MongoDB)
│       ├── redis.tf                  # Azure Cache for Redis
│       ├── acr.tf                    # Azure Container Registry
│       ├── keyvault.tf               # Key Vault + secrets
│       ├── network.tf                # VNets
│       ├── subnets.tf                # Subnet definitions
│       ├── monitor.tf                # Azure Monitor / Grafana
│       ├── jumpserver.tf             # Bastion jump server
│       ├── variables.tf
│       ├── terraform.tfvars
│       ├── outputs.tf
│       └── versions.tf
├── k8s/
│   ├── raw-manifests/                # Plain YAML manifests (local/dev)
│   │   ├── 01-namespace.yml
│   │   ├── 02-storage.yml
│   │   ├── 03-configmaps.yml
│   │   ├── 04-secrets.yml
│   │   ├── 05-postgres.yml
│   │   ├── 06-redis.yml
│   │   ├── 07-rabbitmq.yml
│   │   ├── 08-api.yml
│   │   ├── 09-transcoder.yml
│   │   ├── 10-frontend.yml
│   │   └── 11-network-policy.yml
│   ├── azure-raw/                    # Azure-specific manifests (Key Vault CSI, workload identity)
│   │   ├── 01-streamforge-ns.yml
│   │   ├── 02-secret-provider-class.yml
│   │   ├── 02-sf-workload-sa.yml
│   │   ├── 03-sc-pvc-rabbitmq.yml
│   │   ├── 04-sf-configmaps.yml
│   │   ├── 05-sf-rabbitmq-*.yml
│   │   ├── 06-sf-api-*.yml
│   │   ├── 07-sf-transcoder-*.yml
│   │   ├── 08-sf-frontend-*.yml
│   │   └── 09-network-policy.yml
│   ├── kustomize/                    # Kustomize overlays
│   │   ├── base/                     # Base resources
│   │   └── envs/
│   │       ├── dev/
│   │       └── prod/
│   └── helm/                         # Helm charts (WIP)
├── docs/
├── docker-compose.yml                # Full local dev stack
└── README.md
```

---

## Infrastructure (Terraform)

All Azure infrastructure is provisioned via Terraform located in `infra/terraform/`.

### Azure Resources Provisioned

| Resource | Terraform File | Key Details |
|----------|---------------|-------------|
| Resource Group | `main.tf` | Naming: `rg-{app}-{env}-{region}` |
| AKS Cluster | `aks.tf` | Private cluster, Calico network policy, Workload Identity, auto-scaling, auto-upgrade (patch) |
| PostgreSQL | `postgres.tf` | Flexible Server v18, `B_Standard_B1ms`, pgcrypto + uuid-ossp extensions |
| Cosmos DB | `cosmos.tf` | MongoDB API, Free tier |
| Redis | `redis.tf` | `Balanced_B0` SKU |
| ACR | `acr.tf` | Basic SKU |
| Key Vault | `keyvault.tf` | Stores all app secrets, CSI driver integration |
| VNets & Subnets | `network.tf`, `subnets.tf` | App VNet `10.10.0.0/16`, Jump VNet `10.9.0.0/16` |
| Jump Server | `jumpserver.tf` | Bastion VM for private cluster access |
| Monitoring | `monitor.tf` | Azure Managed Grafana + Prometheus |
| Private Endpoints | `private-endpoint/` | Service-specific private links |

### AKS Cluster Configuration

| Setting | Value |
|---------|-------|
| Private Cluster | `true` |
| RBAC | Enabled |
| Workload Identity | Enabled |
| OIDC Issuer | Enabled |
| Network Plugin | Azure CNI (Overlay mode) |
| Network Policy | Calico |
| Service CIDR | `10.40.0.0/16` |
| DNS Service IP | `10.40.0.10` |
| Node Pool VM Size | `Standard_D2as_v4` |
| OS | Ubuntu |
| OS Disk | 64 GB Managed |
| Auto-scaling | 1–1 nodes (configurable) |
| Max Pods per Node | 250 |
| Auto-upgrade | Patch (Weekly, Sunday) |
| Image Cleaner | Enabled (168h interval) |
| Key Vault CSI | Enabled (2m rotation) |

### Terraform State Backend

State is stored remotely in Azure Blob Storage. Backend config is injected at `terraform init` time via CI/CD secrets:

```bash
terraform init \
  -backend-config="resource_group_name=$BACKEND_RESOURCE_GROUP_NAME" \
  -backend-config="storage_account_name=$BACKEND_STORAGE_ACCOUNT_NAME" \
  -backend-config="container_name=$BACKEND_STORAGE_CONTAINER_NAME" \
  -backend-config="key=$BACKEND_STATE_KEY"
```

### Terraform Providers

| Provider | Version |
|----------|---------|
| `hashicorp/azurerm` | 4.65.0 |
| `hashicorp/azuread` | 3.8.0 |
| `hashicorp/http` | 3.5.0 |
| `hashicorp/tls` | 4.2.1 |

### Running Terraform Locally

```bash
cd infra/terraform

# Initialize with backend
terraform init \
  -backend-config="resource_group_name=<RG_NAME>" \
  -backend-config="storage_account_name=<STORAGE_NAME>" \
  -backend-config="container_name=<CONTAINER>" \
  -backend-config="key=<STATE_KEY>"

# Plan
terraform plan \
  -var="postgres_administrator_password=<PASSWORD>" \
  -var="cosmos_administrator_password=<PASSWORD>" \
  -var="jwt_secret=<SECRET>" \
  -var="rabbitmq_default_pass=<PASSWORD>"

# Apply
terraform apply main.tfplan
```

---

## Kubernetes Deployment

### Deployment Strategies

The project supports **three deployment strategies**:

| Strategy | Path | Use Case |
|----------|------|----------|
| **Raw Manifests** | `k8s/raw-manifests/` | Local/dev — includes in-cluster PostgreSQL, Redis, RabbitMQ |
| **Azure Raw Manifests** | `k8s/azure-raw/` | Azure AKS — uses Azure-managed services, Key Vault CSI |
| **Kustomize** | `k8s/kustomize/` | Environment overlays for dev/prod |

### Namespace

All workloads deploy into the `streamforge` namespace.

### Applying Raw Manifests (Local/Dev)

```bash
# Apply in order (numbered for dependency resolution)
kubectl apply -f k8s/raw-manifests/01-namespace.yml
kubectl apply -f k8s/raw-manifests/02-storage.yml
kubectl apply -f k8s/raw-manifests/03-configmaps.yml
kubectl apply -f k8s/raw-manifests/04-secrets.yml
kubectl apply -f k8s/raw-manifests/05-postgres.yml
kubectl apply -f k8s/raw-manifests/06-redis.yml
kubectl apply -f k8s/raw-manifests/07-rabbitmq.yml
kubectl apply -f k8s/raw-manifests/08-api.yml
kubectl apply -f k8s/raw-manifests/09-transcoder.yml
kubectl apply -f k8s/raw-manifests/10-frontend.yml
kubectl apply -f k8s/raw-manifests/11-network-policy.yml
```

### Applying Azure Manifests (AKS Production)

```bash
# Apply in order
kubectl apply -f k8s/azure-raw/ 
```

> **Note:** Azure manifests use `SecretProviderClass` to sync secrets from Azure Key Vault via the CSI Secrets Store Driver. Ensure the Workload Identity service account (`sf-workload-sa`) is properly federated.

### Kustomize Overlays

```bash
# Dev environment
kubectl apply -k k8s/kustomize/envs/dev/

# Production environment
kubectl apply -k k8s/kustomize/envs/prod/
```

### Resource Limits

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---------|------------|-----------|----------------|--------------|
| API | 250m | 500m | 64Mi | 128Mi |
| Transcoder | — | — | — | — |
| Frontend | — | — | — | — |

### Init Containers

The API deployment uses **init containers** to wait for dependent services before starting:

- `wait-for-postgres` — polls `streamforge-postgres-svc:5432`
- `wait-for-redis` — polls `streamforge-redis-svc:6379`
- `wait-for-rabbitmq` — polls `streamforge-rabbitmq-svc:5672`

---

## CI/CD Pipelines

### GitHub Actions Workflows

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| Backend API CI | `api-ci.yml` | Push/PR to `main` on `backend/api/**` | Build, vet, Docker build & push |
| Frontend CI | `frontend-ci.yml` | Push/PR to `main` on `frontend/**` | Build, Docker build & push |
| Transcoder CI | `transcoder-ci.yml` | Push/PR to `main` on `backend/transcoder/**` | Build, Docker build & push |
| Infra Plan/Apply | `infra-plan.yml` | Push/PR to `infra/**` on `infra/terraform/**` | Terraform fmt → init → validate → tfsec → plan → apply |

### CI Pipeline Flow (Application Services)

```
PR → Build & Vet → (merge to main) → Docker Build → Push to GHCR
```

- **Registry**: `ghcr.io` (GitHub Container Registry)
- **Image naming**: `ghcr.io/<org>/streamforge/<service>`
- **Build action**: Uses reusable composite action at `.github/actions/docker-build-push/`
- **Auth**: OIDC-based (`id-token: write` permission)

### Infra Pipeline Flow

```
PR to infra/* ──▶ fmt ──▶ init ──▶ validate ──▶ tfsec ──▶ plan ──▶ PR Comment
                                                                        │
Merge to infra/* ──────────────────────────────────────────────▶ apply ◀─┘
```

- **Environments**: Branch-based (`infra/development` → dev, `infra/production` → prod)
- **Security**: Terraform plan posted as PR comment, `tfsec` security scanning
- **Auth**: Azure OIDC (`ARM_USE_OIDC: true`, no client secret)
- **Artifacts**: Plan file uploaded as GitHub artifact for audit trail

### Required GitHub Secrets

| Secret | Used By | Description |
|--------|---------|-------------|
| `ARM_CLIENT_ID` | Infra | Azure Service Principal Client ID |
| `ARM_SUBSCRIPTION_ID` | Infra | Azure Subscription ID |
| `ARM_TENANT_ID` | Infra | Azure Tenant ID |
| `BACKEND_RESOURCE_GROUP_NAME` | Infra | Terraform state backend RG |
| `BACKEND_STORAGE_ACCOUNT_NAME` | Infra | Terraform state storage account |
| `BACKEND_STORAGE_CONTAINER_NAME` | Infra | Terraform state container |
| `BACKEND_STATE_KEY` | Infra | Terraform state file key |
| `POSTGRES_PASSWORD` | Infra | PostgreSQL admin password |
| `COSMOS_PASSWORD` | Infra | Cosmos DB admin password |
| `JWT_SECRET` | Infra | Application JWT signing key |
| `RABBITMQ_PASSWORD` | Infra | RabbitMQ password |
| `GITHUB_TOKEN` | All CI | Auto-provided, used for GHCR push |

---

## Containerization

### Docker Images

| Service | Dockerfile | Base (Build) | Base (Runtime) | Exposed Port | User |
|---------|-----------|-------------|----------------|-------------|------|
| **Frontend** | `frontend/Dockerfile` | `node:24.11.0-alpine` | `node:24.11.0-alpine` | 3000 | `nextjs` (UID 1001) |
| **API** | `backend/api/Dockerfile` | `golang:1.24.0` | `alpine:3.23` | 8080 | `appuser` (UID 10001) |
| **Transcoder** | `backend/transcoder/Dockerfile` | `golang:1.24.0` | `alpine:3.23` + FFmpeg 8.0.1 | 9090 | `appuser` (UID 10001) |

### Build Features

- **Multi-stage builds** — separate build and runtime stages for minimal image size
- **Build cache mounts** — `--mount=type=cache` for Go modules and npm packages
- **Non-root users** — all containers run as non-root
- **CGO disabled** — static Go binaries (`CGO_ENABLED=0`)
- **Next.js standalone** — output mode for minimal production bundle

### Building Images Locally

```bash
# Frontend
docker build -t streamforge-frontend:local ./frontend

# API
docker build -t streamforge-api:local ./backend/api

# Transcoder
docker build -t streamforge-transcoder:local ./backend/transcoder
```

---

## Environment Variables Reference

### Frontend

| Variable | Description | Example |
|----------|-------------|---------|
| `API_URL` | Backend API endpoint | `http://api:8080` |
| `APP_ENV` | `production` or `development` — controls log verbosity | `development` |

### API Server

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL hostname | `postgres` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database user | `postgres` |
| `POSTGRES_PASSWORD` | Database password | — |
| `POSTGRES_DB` | Database name | `streamforge` |
| `POSTGRES_SSL_MODE` | SSL mode (`disable` / `require`) | `disable` |
| `REDIS_URL` | Redis connection URL | `redis://redis:6379` |
| `REDIS_PASSWORD` | Redis password (empty if none) | — |
| `RABBITMQ_URL` | RabbitMQ AMQP URL | `amqp://guest:guest@rabbitmq:5672/` |
| `AZURE_STORAGE_ACCOUNT_NAME` | Azure Storage account name | — |
| `AZURE_STORAGE_ACCOUNT_KEY` | Azure Storage access key | — |
| `AZURE_STORAGE_ENDPOINT` | Blob endpoint URL | `https://<name>.blob.core.windows.net` |
| `AZURE_BLOB_RAW_CONTAINER` | Container for raw uploads | `raw-videos` |
| `AZURE_BLOB_HLS_CONTAINER` | Container for HLS output | `hls-videos` |
| `COSMOS_CONNECTION_STRING` | MongoDB-compatible connection string | `mongodb://...` |
| `COSMOS_DATABASE` | Cosmos DB database name | `streamforge` |
| `JWT_SECRET` | JWT signing secret | — |
| `JWT_EXPIRY_HOURS` | JWT token TTL (hours) | `24` |
| `REFRESH_TOKEN_EXPIRY_DAYS` | Refresh token TTL (days) | `7` |
| `CORS_ALLOWED_ORIGIN` | Allowed CORS origin | `http://localhost:3000` |

### Transcoder Worker

| Variable | Description | Example |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL hostname | `postgres` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database user | `postgres` |
| `POSTGRES_PASSWORD` | Database password | — |
| `POSTGRES_DB` | Database name | `streamforge` |
| `POSTGRES_SSL_MODE` | SSL mode | `disable` |
| `RABBITMQ_URL` | RabbitMQ AMQP URL | `amqp://guest:guest@rabbitmq:5672/` |
| `RABBITMQ_QUEUE` | Queue name to consume from | `transcoder.jobs` |
| `RABBITMQ_DEAD_LETTER_QUEUE` | DLQ for failed jobs | `transcoder.jobs.dlq` |
| `AZURE_STORAGE_ACCOUNT_NAME` | Azure Storage account name | — |
| `AZURE_STORAGE_ACCOUNT_KEY` | Azure Storage access key | — |
| `AZURE_STORAGE_ENDPOINT` | Blob endpoint URL | — |
| `AZURE_BLOB_RAW_CONTAINER` | Container for raw uploads | `raw-videos` |
| `AZURE_BLOB_HLS_CONTAINER` | Container for HLS output | `hls-videos` |
| `FFMPEG_PATH` | Path to FFmpeg binary | `/usr/bin/ffmpeg` |
| `TEMP_DIR` | Temp directory for transcoding | `/tmp` |

---

## Local Development (Docker Compose)

The `docker-compose.yml` spins up the **entire stack** locally, including all infrastructure dependencies.

### Services

| Service | Image | Port Mapping | Healthcheck |
|---------|-------|-------------|-------------|
| `frontend` | Built from `./frontend` | `3000:3000` | `wget http://localhost:3000/` |
| `api` | Built from `./backend/api` | `8080:8080` | `wget http://localhost:8080/healthz` |
| `transcoder` | Built from `./backend/transcoder` | `9090:9090` | — |
| `postgres` | `postgres:15-alpine` | `5432:5432` | `pg_isready -U postgres` |
| `redis` | `redis:alpine` | `6379:6379` | `redis-cli ping` |
| `rabbitmq` | `rabbitmq:4.2.4-alpine` | `5672:5672` | `rabbitmq-diagnostics ping` |
| `mongodb` | `mongodb/mongodb-community-server:8.0-ubi8` | `27017:27017` | `mongosh --quiet --eval` |

### Quick Start

```bash
# Set required environment variables
export AZURE_STORAGE_ACCOUNT_NAME=<your_account>
export AZURE_STORAGE_ACCOUNT_KEY=<your_key>
export JWT_SECRET=<your_secret>

# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Tear down
docker compose down -v
```

### Dependency Order

Docker Compose manages startup dependencies using health checks:

```
postgres ─┐
redis ────┤
rabbitmq ─┼──▶ api ──▶ frontend
mongodb ──┘
              rabbitmq ──▶ transcoder
              postgres ──┘
```

### Persistent Volumes

| Volume | Mount Point | Service |
|--------|------------|---------|
| `postgres_data` | `/var/lib/postgresql/data` | postgres |
| `redis_data` | `/data` | redis |
| `mongodb_data` | `/data/db` | mongodb |

---

## Running Services Individually

### Frontend — Next.js (Node 24.11.0)

```bash
cd ./frontend
npm install
npm run dev          # Development server (hot reload)
npm run build        # Production build → ./frontend/build/standalone
```

**Production run after build:**

```bash
cp -r ./build/static ./build/standalone/build/static
node ./build/standalone/server.js
```

### Backend API — Go 1.24.0

```bash
cd ./backend/api
go mod download
go run ./cmd/server/main.go          # Development
CGO_ENABLED=0 go build -o /bin/server ./cmd/server/main.go  # Production build
```

### Transcoder Worker — Go 1.24.0

```bash
cd ./backend/transcoder
go mod download
go run ./cmd/worker/main.go          # Development
CGO_ENABLED=0 go build -o /bin/worker ./cmd/worker/main.go  # Production build
```

> **Note:** The transcoder requires FFmpeg to be installed on the host. Set `FFMPEG_PATH` to the FFmpeg binary location.

---

## Secrets Management

### Local / Raw Manifests

Secrets are base64-encoded in `k8s/raw-manifests/04-secrets.yml` (for local dev only).

### Azure AKS (Production)

Secrets are managed via **Azure Key Vault** with the **Secrets Store CSI Driver**:

1. Secrets stored in Azure Key Vault (`kv-streamforgedevci`)
2. `SecretProviderClass` (`k8s/azure-raw/02-secret-provider-class.yml`) maps KV secrets to K8s secrets
3. Workload Identity (`sf-workload-sa`) authenticates pods to Key Vault
4. Secrets auto-rotate every **2 minutes** via CSI driver

**Key Vault Secrets Synced:**

| Key Vault Secret Name | K8s Secret Key | Used By |
|----------------------|----------------|---------|
| `rabbitmq-url` | `RABBITMQ_URL` | API, Transcoder |
| `rabbitmq-default-pass` | `RABBITMQ_DEFAULT_PASS` | RabbitMQ |
| `azure-storage-account-key` | `AZURE_STORAGE_ACCOUNT_KEY` | API, Transcoder |
| `postgres-password` | `POSTGRES_PASSWORD` | API, Transcoder |
| `cosmos-connection-string` | `COSMOS_CONNECTION_STRING` | API |
| `redis-password` | `REDIS_PASSWORD` | API |
| `jwt-secret` | `JWT_SECRET` | API |

---

## Network Policies & Security

Kubernetes NetworkPolicies enforce **least-privilege network access** within the `streamforge` namespace:

| Policy | Target | Allowed Sources | Port |
|--------|--------|-----------------|------|
| `allow-only-api-transcoder-to-postgres` | PostgreSQL (`database` tier) | API, Transcoder (`backend` tier) | 5432 |
| `allow-only-api-transcoder-to-redis` | Redis (`cache` tier) | API, Transcoder (`backend` tier) | 6379 |
| `allow-only-api-transcoder-to-rabbitmq` | RabbitMQ (`queue` tier) | API, Transcoder (`backend` tier) | 5672 |
| `allow-only-frontend-to-api` | API (`backend` tier) | Frontend (`frontend` tier) | 8080 |

> **Note:** All backend services use standard `app.kubernetes.io/*` labels for pod selection. Network policies enforce that databases and caches are **only reachable** from authorized backend services.

---

## Health Checks & Probes

### Docker Compose Health Checks

| Service | Check | Interval | Timeout | Retries | Start Period |
|---------|-------|----------|---------|---------|-------------|
| Frontend | `wget http://localhost:3000/` | 10s | 5s | 5 | 15s |
| API | `wget http://localhost:8080/healthz` | 10s | 5s | — | 120s |
| PostgreSQL | `pg_isready -U postgres` | 10s | 5s | 5 | 20s |
| Redis | `redis-cli ping \| grep PONG` | 1s | 3s | 5 | — |
| RabbitMQ | `rabbitmq-diagnostics ping` | 10s | 5s | 5 | — |
| MongoDB | `mongosh --quiet --eval quit(...)` | 10s | 10s | 5 | 30s |

### Kubernetes Probes (API)

| Probe | Type | Endpoint | Initial Delay | Period | Timeout |
|-------|------|----------|--------------|--------|---------|
| Liveness | `exec` | `wget http://localhost:8080/healthz` | 20s | 10s | 5s |
| Readiness | `exec` | `wget http://localhost:8080/healthz` | 20s | 10s | 5s |

---

## Observability

| Component | Tool | Purpose |
|-----------|------|---------|
| Metrics | Prometheus | Scrapes application and cluster metrics |
| Dashboards | Azure Managed Grafana | Visualization and alerting |
| Infrastructure | Azure Monitor | Resource-level monitoring |

---

## Required External Dependencies

| Dependency | Minimum Version | Required Extensions / Notes |
|------------|----------------|-----------------------------|
| PostgreSQL | 17+ | `pgcrypto`, `uuid-ossp` extensions |
| Redis | — | Standard Redis instance |
| RabbitMQ | 4.x | Queues: `transcoder.jobs`, `transcoder.jobs.dlq` |
| MongoDB / Cosmos DB | — | MongoDB wire-protocol compatible |
| Azure Storage Account | — | Two containers: `raw-videos`, `hls-videos` |
| FFmpeg | 8.0.1+ | Required on transcoder host/container |
| Go | 1.24.0 | For building backend services |
| Node.js | 24.11.0 | For building frontend |
| Terraform | — | With providers listed in `versions.tf` |
| kubectl | 1.35+ | For cluster management |
