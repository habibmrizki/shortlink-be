# 🔗 Shortlink App — Backend Service

A high-performance, production-ready URL Shortener REST API built with **Go (Gin)**, **PostgreSQL**, and **Redis**. Designed with clean architecture principles, robust security practices, and containerized deployment.

---

## 📑 Table of Contents

- [Architecture & Tech Stack](#-architecture--tech-stack)
- [Key Features](#-key-features)
- [Database Schema](#-database-schema)
- [Environment Variables](#-environment-variables)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Method 1: Docker Compose (Recommended)](#method-1-docker-compose-recommended)
  - [Method 2: Manual Local Setup](#method-2-manual-local-setup)
- [Makefile Commands](#-makefile-commands)
- [API Documentation](#-api-documentation)
  - [Authentication Endpoints](#1-authentication)
  - [Link Management Endpoints](#2-link-management)
  - [Redirect Endpoint](#3-redirection)
  - [System Health](#4-system)
- [Architecture & Engineering Decisions](#-architecture--engineering-decisions)
- [CI/CD & Container Registry](#-cicd--container-registry)

---

## 🏛️ Architecture & Tech Stack

```
shortlink-be/
├── cmd/
│   └── main.go                  # Application entry point & dependency wiring
├── internal/
│   ├── handlers/                # HTTP request controllers (Gin)
│   ├── middlewares/             # JWT auth & CORS middlewares
│   ├── models/                  # Request/response DTOs & DB entities
│   ├── repositories/            # Database queries using pgxpool & SQL
│   ├── routes/                  # Route registrations & grouping
│   └── services/                # Business logic, transactions, Redis caching
├── pkg/
│   ├── database/                # PostgreSQL pgxpool & Redis client setup
│   └── utils/                   # Password hashing (Argon2id), JWT, slug generator
├── schema.sql                   # Database DDL schema & indexes
├── Dockerfile                   # Multi-stage lightweight Docker image
├── docker-compose.yaml          # Multi-container orchestration (App, PG, Redis, FE)
└── Makefile                     # Developer workflow automation
```

- **Language / Framework**: Go 1.27.2 with [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Primary Database**: PostgreSQL 17 (via [`pgx/v5`](https://github.com/jackc/pgx) with connection pooling)
- **Caching & Blacklisting**: Redis 8 (via [`go-redis/v9`](https://github.com/redis/go-redis))
- **Password Security**: Argon2id (`golang.org/x/crypto/argon2`)
- **Authentication**: Stateless Short-Lived JWT (`golang-jwt/jwt/v5`) with server-side token revocation in Redis
- **Containerization**: Docker & Docker Compose

---

## ✨ Key Features

1. **Authentication & Authorization**
   - User registration & login with secure Argon2id password hashing.
   - Stateless JWT authentication using short-lived access tokens (no refresh tokens).
   - Instant token invalidation (logout) via Redis blacklist.
2. **Link Management**
   - Shorten long URLs with cryptographically secure random 6-character slugs (`crypto/rand`).
   - Custom slug support with strict validations:
     - Length: 3–50 characters.
     - Allowed characters: alphanumeric and hyphens (`^[a-zA-Z0-9-]+$`).
     - Reserved system words blocked (`api`, `login`, `register`, `dashboard`).
   - Real-time slug availability check endpoint (`/api/links/check`).
   - Per-user link retrieval with pagination support (`page`, `limit`).
   - Soft delete mechanism: preserves analytics history while instantly disabling public redirects.
3. **High-Performance Redirection**
   - Public endpoint `/{slug}` returning **HTTP 301 Moved Permanently**.
   - Redis caching for active slugs (24h TTL) to achieve sub-millisecond redirect latency.
   - Atomic click counter increments.
   - Soft-deleted or non-existent slugs return **404 Not Found**.

---

## 🗄️ Database Schema

The database utilizes PostgreSQL with indexed foreign keys and partial unique indexes for optimal read performance and soft delete integrity:

```sql
-- PostgreSQL Schema (schema.sql)

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_url TEXT NOT NULL,
    slug VARCHAR(50) NOT NULL,
    click_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for maximum query efficiency
CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id);
-- Partial unique index ensures active slugs are unique while allowing reuse of deleted slugs
CREATE UNIQUE INDEX IF NOT EXISTS idx_links_slug ON links(slug) WHERE deleted_at IS NULL;
```

---

## ⚙️ Environment Variables

Create a `.env` file in the root directory based on `.env.example`:

| Variable         | Description                                                                                  | Default Local           | Docker Compose Default  |
| :--------------- | :------------------------------------------------------------------------------------------- | :---------------------- | :---------------------- |
| `PORT`           | HTTP server listening port                                                                   | `2901`                  | `2901`                  |
| `BASE_URL`       | Optional base URL used to prefix short links; leave empty to derive it from the request host | empty (request host)    | empty (request host)    |
| `DBHOST`         | PostgreSQL host                                                                              | `localhost`             | `shortlink-pg`          |
| `DBPORT`         | PostgreSQL port                                                                              | `5432`                  | `5432`                  |
| `DBUSER`         | PostgreSQL username                                                                          | `postgres`              | `habib`                 |
| `DBPASS`         | PostgreSQL password                                                                          | `postgres`              | `kuningan1`             |
| `DBNAME`         | PostgreSQL database name                                                                     | `shortlink`             | `shortlink`             |
| `REDIS_HOST`     | Redis host                                                                                   | `127.0.0.1`             | `shortlink-rdb`         |
| `REDIS_PORT`     | Redis port                                                                                   | `6379`                  | `6379`                  |
| `REDIS_PASSWORD` | Redis password (optional)                                                                    | _empty_                 | _empty_                 |
| `JWT_SECRET`     | Secret key for signing JWT tokens                                                            | `your-super-secret-key` | `your-super-secret-key` |
| `JWT_ISSUER`     | JWT issuer tag                                                                               | `shortlink`             | `shortlink`             |

---

## 🚀 Getting Started

### Prerequisites

- [Go 1.24+](https://golang.org/dl/) (for local development)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)
- [PostgreSQL](https://www.postgresql.org/) & [Redis](https://redis.io/) (if running without Docker)

---

### Method 1: Docker Compose (Recommended)

The Compose file pulls the frontend and backend images from GHCR and starts them with PostgreSQL and Redis. The frontend image is built in the separate [shortlink-fe repository](https://github.com/habibmrizki/shortlink-fe).

```bash
# Clone the backend repository
git clone https://github.com/habibmrizki/shortlink-be.git
cd shortlink-be

# Create and configure environment values
cp .env.example .env
# Set secure values for DBPASS and JWT_SECRET before exposing the app.

# Pull published images and start the application
docker compose pull
docker compose up -d
docker compose ps

# Follow logs
docker compose logs -f shortlink-be
```

- **Frontend app and short links**: `http://localhost:5174`
- **Backend API**: `http://localhost:2901`
- **Swagger UI**: `http://localhost:2901/swagger/index.html`
- **PostgreSQL**: `localhost:5433`; **Redis**: `localhost:6380`

To open the app from another device on the same network, use `http://<host-lan-ip>:5174`. Short links use the incoming host and port by default. Set `BASE_URL` in `.env` when using a public domain, for example `https://short.example.com`.

If a GHCR package is private, authenticate before pulling:

```bash
docker login ghcr.io -u <github-username>
docker compose pull
```

The PostgreSQL initialization script runs only when the named database volume is first created. Existing volumes keep their data and are not reinitialized by `docker compose up`.

---

### Method 2: Manual Local Setup

1. **Start PostgreSQL and Redis** (Docker Compose can provide these services):

```bash
docker compose up -d shortlink-pg shortlink-rdb
```

2. **Configure the backend environment**:

   ```bash
   cp .env.example .env
   ```

   Set `DBHOST=localhost`, `DBPORT=5433` and `REDIS_HOST=localhost`, `REDIS_PORT=6380` when using the Compose database/cache ports above. Ensure `DBUSER`, `DBPASS`, and `DBNAME` match the PostgreSQL container credentials.

3. **Initialize the schema and run the API**:
   ```bash
   psql -h localhost -p 5433 -U <DBUSER> -d <DBNAME> -f schema.sql
   go mod download
   go run ./cmd/main.go
   ```

The backend listens on `http://localhost:2901`. If you use locally installed PostgreSQL and Redis instead, use their host/port in `.env` and apply `schema.sql` to the configured database.

---

## 🛠️ Makefile Commands

The Makefile currently provides database migration and seed commands; the application itself is started with `go run ./cmd/main.go`.

| Command                             | Description                           |
| :---------------------------------- | :------------------------------------ |
| `make migrate-create NAME=<name>`   | Create a numbered migration file      |
| `make migrate-createUp`             | Apply migrations from `db/migrations` |
| `make migrate-createDown s=<count>` | Roll back migrations                  |
| `make migrate-status`               | Show the current migration version    |
| `make migrate-force v=<version>`    | Set the migration version manually    |
| `make seed`                         | Apply the configured seed SQL file    |

---

## 📡 API Documentation

Base URL: `http://localhost:2901`

### 1. Authentication

#### `POST /api/register`

Register a new user account.

- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "Password123"
  }
  ```
- **Response `201 Created`**:
  ```json
  {
    "success": true,
    "message": "User registered successfully",
    "data": {
      "id": 1,
      "email": "user@example.com",
      "created_at": "2026-10-22T10:00:00Z"
    }
  }
  ```

#### `POST /api/login`

Authenticate user and obtain a short-lived access token.

- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "Password123"
  }
  ```
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Login successful",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "user": {
        "id": 1,
        "email": "user@example.com",
        "created_at": "2026-10-22T10:00:00Z"
      }
    }
  }
  ```

#### `POST /api/logout`

Revoke the current access token by storing it in the Redis blacklist.

- **Headers**: `Authorization: Bearer <token>`
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Logout successful"
  }
  ```

---

### 2. Link Management

_(Requires `Authorization: Bearer <token>` header)_

#### `POST /api/links`

Create a new shortened link.

- **Request Body**:
  ```json
  {
    "original_url": "https://google.com/search?q=fullstack",
    "custom_slug": "my-promo" // Optional: 3-50 chars alphanumeric + hyphens
  }
  ```
- **Response `201 Created`**:
  ```json
  {
    "success": true,
    "message": "Short link created successfully",
    "data": {
      "id": 1,
      "user_id": 1,
      "original_url": "https://google.com/search?q=fullstack",
      "slug": "my-promo",
      "short_url": "http://localhost:2901/my-promo",
      "click_count": 0,
      "created_at": "2026-10-22T10:05:00Z"
    }
  }
  ```

#### `GET /api/links`

Retrieve user's links with optional pagination.

- **Query Params**:
  - `page`: Page number (default: `1`)
  - `limit`: Items per page (default: `10`)
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Links retrieved successfully",
    "data": {
      "items": [
        {
          "id": 1,
          "original_url": "https://google.com/search?q=fullstack",
          "slug": "my-promo",
          "short_url": "http://localhost:2901/my-promo",
          "click_count": 14,
          "created_at": "2026-10-22T10:05:00Z"
        }
      ],
      "pagination": {
        "current_page": 1,
        "limit": 10,
        "total_items": 1,
        "total_pages": 1
      }
    }
  }
  ```

#### `GET /api/links/check`

Check slug availability before form submission.

- **Query Params**: `slug=custom-name`
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Slug is available",
    "data": {
      "available": true,
      "slug": "custom-name"
    }
  }
  ```

#### `DELETE /api/links/:id`

Soft-delete a link owned by the user.

- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Link deleted successfully"
  }
  ```

---

### 3. Redirection

#### `GET /{slug}`

Public redirection endpoint.

- **Status `301 Moved Permanently`**: Redirects user agent to original destination URL.
- **Status `404 Not Found`**: Returned if the slug does not exist or has been soft-deleted.

---

### 4. System

#### `GET /health`

Liveness/readiness probe.

- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Server is healthy",
    "data": null
  }
  ```

---

## 🧠 Architecture & Engineering Decisions

1. **Database Transactions (`tx.Begin`)**:
   - Short link creation uses PostgreSQL transactions to prevent race conditions during concurrent slug allocations and ensure referential integrity.
2. **Partial Unique Indexing**:
   - `idx_links_slug ON links(slug) WHERE deleted_at IS NULL` allows re-registering previously deleted custom slugs while maintaining strict unicity across all active links.
3. **Multi-tier Caching (Redis)**:
   - Frequent redirect targets are cached in Redis with a 24-hour TTL, relieving database read pressure and enabling sub-millisecond redirect speeds.
   - Deleted links immediately purge their corresponding Redis cache key to guarantee instant propagation.
4. **Argon2id vs bcrypt**:
   - Implements Argon2id for password hashing, offering superior protection against GPU-based cracking attacks while remaining memory and CPU configurable.

---

## 🚢 CI/CD & Container Registry

- **Workflow**: Automated GitHub Actions pipeline located in `.github/workflows/build-image-and-push-to-ghcr.yml`.
- **Trigger**: Every push to the `master`/`main` branch.
- **Image Target**: Published to GitHub Container Registry (`ghcr.io/<username>/shortlink-be:latest`).
