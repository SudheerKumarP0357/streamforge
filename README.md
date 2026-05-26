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

```mermaid

                    ┌───────────┐
                    │   User    │
                    └─────┬─────┘
                          │
                          ▼
                ┌──────────────────┐
                │ Frontend         │
                │ Next.js + React  │
                │ Port: 3000       │
                └────────┬─────────┘
                         │ HTTPS
                         ▼
                ┌──────────────────┐
                │ API Service      │
                │ Go 1.24          │
                │ Port: 8080       │
                └─────┬────┬───────┘
                      │    │
        ┌─────────────┘    └──────────────┐
        ▼                                 ▼
┌────────────────┐              ┌────────────────┐
│ PostgreSQL     │              │ Redis Cache    │
│ Metadata Store │              │ Metadata Cache │
└────────────────┘              └────────────────┘
        │
        ▼
┌────────────────────┐
│ Raw Video Storage  │
│ Azure Blob Storage │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ RabbitMQ           │
│ Transcoding Queue  │
└─────────┬──────────┘
          ▼
┌────────────────────┐
│ Transcoder Worker  │
│ FFmpeg             │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ HLS Blob Storage   │
│ Streaming Assets   │
└────────────────────┘
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