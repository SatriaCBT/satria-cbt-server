# Satria CBT Server

REST API server for Computer-Based Test (CBT) system — microservice architecture built with Go Fiber and GORM.

## Tech Stack

- **Framework:** [Go Fiber](https://gofiber.io/) v2
- **ORM:** [GORM](https://gorm.io/)
- **Database:** PostgreSQL
- **Auth:** JWT (golang-jwt)
- **Realtime:** WebSocket (gofiber/contrib/websocket)
- **Export:** Excelize

## Architecture

```
┌─────────┐     ┌──────────┐     ┌──────────┐
│ Client  │────▶│ Gateway  │────▶│  Auth    │
│         │     │ :3000    │     │  :3001   │
│         │     │          │────▶│  Exam    │
│         │     │          │     │  :3002   │
└─────────┘     └──────────┘     └──────────┘
                       │                │
                       └──────┬─────────┘
                              │
                         ┌────▼────┐
                         │  Postgres │
                         └─────────┘
```

- **Gateway** (port 3000) — validates JWT, injects `X-User-ID` / `X-User-Role` headers, proxies to services
- **Auth** (port 3001) — admin/teacher/student login, register, profile
- **Exam** (port 3002) — subjects, questions, exams, attempts, auto-grading, dashboard, Excel export, WebSocket

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL

### Running (dev)

```bash
git clone https://github.com/SatriaCBT/satria-cbt-server
cd satria-cbt-server

cp .env.example .env
# edit .env with your DB config

# Start each service in separate terminals:
cd services/gateway && go run .
cd services/auth   && go run .
cd services/exam   && go run .
```

### With Docker

```bash
docker compose up --build
```

### API Docs

Gateway routes are accessible at `http://localhost:3000/api/*` after startup.

## Project Structure

```
├── pkg/                 # Shared library
│   ├── auth/jwt.go      # JWT middleware
│   ├── database/db.go   # Singleton DB connection
│   └── response/        # Response DTO
├── services/
│   ├── gateway/         # API gateway (port 3000)
│   │   ├── main.go      # JWT validation + reverse proxy
│   │   └── go.mod
│   ├── auth/            # Auth service (port 3001)
│   │   ├── models/      # Admin, Teacher, Student
│   │   ├── handlers/    # Login, Register, CRUD
│   │   ├── main.go
│   │   └── go.mod
│   └── exam/            # Exam service (port 3002)
│       ├── models/      # Subject, Question, Exam, Attempt, Answer, Class
│       ├── handlers/    # CRUD, grading, WS, export
│       ├── main.go
│       └── go.mod
├── Dockerfile.gateway
├── Dockerfile.auth
├── Dockerfile.exam
├── docker-compose.yml
└── .env.example
```
