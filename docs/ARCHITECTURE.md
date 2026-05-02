# Product Catalog API Architecture

## Overview

The application is a small Product Catalog REST API written in Go. It exposes HTTP endpoints for CRUD operations on products and can run with either an in-memory store for local development/tests or a PostgreSQL store for Docker Compose and persistent data.

At startup, `cmd/api/main.go` reads the environment:

- If `DB_HOST` is set, the API connects to PostgreSQL, ensures the `products` table exists, and registers the PostgreSQL-backed handlers.
- If `DB_HOST` is not set, the API creates a `MemoryStore` and registers the in-memory handlers.

## Request flow

```text
Client
  |
  | HTTP request
  v
Gorilla Mux Router
  |
  | Route match: /health, /products, /products/{id}
  v
HTTP Handler
  |
  | Decode JSON, validate Product, map errors to HTTP responses
  v
Store Layer
  |
  | MemoryStore: Go map protected by mutex
  | PostgresStore: SQL queries through database/sql and lib/pq
  v
Database
```

Example: `POST /products`

1. The client sends a JSON request body such as `{"name":"Widget","price":9.99}`.
2. Gorilla Mux matches `POST /products` and calls `CreateProduct`.
3. The handler decodes the request body into `model.Product`.
4. `Product.Validate()` rejects invalid products, for example an empty name or a negative price.
5. The handler calls the active store implementation.
6. The store assigns or receives the product ID:
   - `MemoryStore.Create` increments an in-process `nextID` counter and stores the product in a map.
   - `PostgresStore.Create` inserts the product into PostgreSQL and reads the generated `SERIAL` ID.
7. The handler returns the created product as JSON with HTTP status `201 Created`.

## Runtime Architecture with Docker Compose

```text
┌──────────────┐        HTTP        ┌────────────────────────┐
│    Client    │ ─────────────────▶ │ Product Catalog API    │
│ curl/browser │                    │ Go service, port 8080  │
└──────────────┘                    └───────────┬────────────┘
                                                │
                                                │ SQL over TCP
                                                v
                                    ┌────────────────────────┐
                                    │ PostgreSQL             │
                                    │ postgres:16-alpine     │
                                    │ port 5432              │
                                    └───────────┬────────────┘
                                                │
                                                │
                                                v
                                    ┌────────────────────────┐
                                    │ Docker volume: pgdata  │
                                    └────────────────────────┘
```

## MemoryStore vs. PostgresStore

The API uses two different storage implementations depending on the environment. Understanding their trade-offs is key to knowing when to use each:

**MemoryStore (`internal/store/memory.go`)**
* **How it works:** Stores products directly in the application's memory using a Go `map` and handles concurrency with a `sync.RWMutex`. It generates its own local `nextID` sequence. 
* **Trade-offs:** It is extremely fast and requires zero external infrastructure to run. However, all data is completely lost the moment the API process restarts or crashes.
* **When to use:** Ideal for fast, isolated local development, quick manual testing, and unit tests using `httptest` where you need a clean state and no dependencies.

**PostgresStore (`internal/store/postgres.go`)**
* **How it works:** Connects to an external PostgreSQL database, delegating data storage to the `products` table, ID generation to `SERIAL PRIMARY KEY`, and concurrency to the database engine.
* **Trade-offs:** It provides true data persistence (data survives restarts via the Docker `pgdata` volume). The downside is that it requires setting up and running external infrastructure (like a Docker container) before the API can function.
* **When to use:** Necessary for integration testing, running the full containerized setup via Docker Compose, and any production-like environment where data must be persistent.
