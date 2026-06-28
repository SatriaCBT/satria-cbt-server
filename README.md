# Satria CBT Server

REST API server for Computer-Based Test (CBT) system built with Go Fiber and GORM.

## Tech Stack

- **Framework:** [Go Fiber](https://gofiber.io/) v2
- **ORM:** [GORM](https://gorm.io/)
- **Database:** PostgreSQL
- **Auth:** JWT (golang-jwt)
- **Docs:** Swagger/OpenAPI

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL

### Running

```bash
# Copy env and configure
cp .env.example .env

# Run
go run .
```

### With Docker

```bash
docker compose up --build
```

### API Docs

Visit `http://localhost:3000/swagger/` after starting the server.

## Project Structure

```
├── configs/        # Database config, default admin seeder
├── controllers/    # Request handlers
├── middleware/      # JWT auth middleware
├── models/         # GORM models
├── res/            # Response DTOs
├── routers/        # Route definitions
├── docs/           # Swagger generated docs
├── main.go         # Entry point
└── Dockerfile
```
