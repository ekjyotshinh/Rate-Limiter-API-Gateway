# Rate-Limiter API Gateway

A small project that implements an API Gateway with a rate-limiting service. The repository contains an HTTP gateway, a gRPC-based rate limiter service, protocol buffers, deployment manifests, and Dockerfiles.

## Repository layout

```
gateway/
├─ cmd/
│  ├─ gateway/         # HTTP gateway main
│  └─ ratelimiter/     # gRPC rate limiter main
├─ internal/
│  ├─ gateway/
│  │  ├─ middleware/
│  │  ├─ router.go
│  │  └─ auth.go
│  └─ ratelimiter/
│     └─ lua/          # lua script for atomic operations
├─ proto/
│  └─ ratelimiter.proto
├─ deploy/
│  ├─ docker-compose.yml
│  └─ k8s/
│     ├─ ratelimiter-deploy.yaml
│     └─ gateway-deploy.yaml
├─ Dockerfile.gateway
└─ Dockerfile.ratelimiter
```

## Features

- HTTP gateway exposing endpoints to clients
- gRPC rate limiter service for enforcing limits
- Lua scripts for atomic operations (e.g., rate counter updates)
- Docker and Kubernetes manifests for quick deployment