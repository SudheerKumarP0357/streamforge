# StreamForge — video streaming platform built for the modern cloud - Azure

## Architecture Overview

StreamForge follows a microservices architecture deployed on Azure Kubernetes Service. All services are containerized, with infrastructure provisioned via Terraform and deployed through GitHub Actions & Azure DevOps CI/CD pipelines.

|Layer|Component|Azure Service|Purpose|
|-----|-----|-----|-----|
|Frontend|NextJs + React|AKS Pods|SSR web app & video UI|
|API Gateway|Go API Service|AKS Pod + App Gateway|REST API, auth, routing|
|Worker|Go Transcoder|AKS Pod (async)|FFmpeg video processing|
|Messaging|RabbitMQ|AKS StatefulSet|Async job queue|
|Video Storage|HLS segments + raw|Azure Blob Storage|Object storage for video|
|Relational DB|PostgreSQL|Azure DB for PostgreSQL|Users, content, subs|
|Document DB|MongoDB / Cosmos DB|Azure Cosmos DB (Mongo)|Watch history, prefs|
|Cache|Redis|Azure Cache for Redis|Sessions, rate limiting|
|Secrets|Key Vault|Azure Key Vault|Credentials & certs|
|Observability|Prometheus + Grafana|Azure Managed Grafana|Metrics & dashboards|
|CI/CD|GitHub Actions/Azure Pipelines|ACR + AKS Deploy|Build, test, deploy|
|IaC|Terraform|Azure Resource Manager|All infra provisioned|

