# High-Performance Banking API

This project is a robust, scalable backend service for a modern banking application, built entirely in Go. It provides a complete API for secure user management, atomic financial transactions, and real-time balance inquiries.

The system is designed with a focus on high concurrency, full observability, and enterprise-grade security practices. The entire stack, including the application, database, cache, and monitoring tools, is containerized with Docker for a seamless, one-command setup and a consistent development environment.

## Key Features

### Secure Authentication & Authorization
- **JWT-Based Authentication**: Secure session management using JSON Web Tokens.
- **Password Hashing**: Uses the robust bcrypt algorithm to securely store user passwords.
- **Role-Based Access Control (RBAC)**: Differentiates between user and admin roles, with specific endpoints protected by an admin-only middleware.

### Robust Transactional System
- **Atomic Operations & Rollback Mechanism**: Guarantees data integrity for all financial operations (transfer, credit, debit) by wrapping them in ACID-compliant database transactions. In case of any failure during an operation, the entire transaction is automatically rolled back, preventing data loss and ensuring the database remains in a consistent state.
- **Asynchronous Processing**: Utilizes a Worker Pool to process transactions in the background, ensuring the API remains highly responsive and available even under heavy load.
- **State Management**: Implements a clear state transition model for transactions (e.g., pending -> completed).

### High-Performance Architecture
- **Redis Caching**: Implements a "Cache-Aside" pattern with Redis to dramatically improve read performance and reduce database load for frequently accessed data like user profiles and balances.
- **Optimized Database Access**: Uses parametrized queries to prevent SQL injection and follows best practices for repository-layer design.
- **Thread-Safe Operations**: Ensures that in-memory operations, such as balance updates, are thread-safe.

### Full Observability Stack
- **Structured Logging**: Employs slog for structured, machine-readable logs. Every incoming request is tagged with a unique UUID for end-to-end tracing and debugging.
- **Prometheus Metrics**: Exposes key performance indicators (request rate, latency, error counts) on a `/metrics` endpoint for real-time monitoring.
- **Grafana Dashboards**: The stack includes Grafana, ready to be configured with dashboards to visualize application performance.

### Professional-Grade API Design
- **Clean Routing**: Uses the efficient chi router for defining API routes and middleware.
- **Robust Middleware Chain**: Includes custom middleware for logging, error handling, authentication, and CORS header management.
- **Rate Limiting**: Protects the API from brute-force attacks and abuse by limiting the number of requests a client can make in a given time frame.
- **Graceful Shutdown**: Implemented to ensure the server finishes processing in-flight requests before shutting down, preventing data loss.

### Containerized & Deployable
- **Multi-Stage Dockerfile**: Creates a minimal, secure, and optimized production image.
- **Full Docker Compose Stack**: Manages the entire multi-container environment (app, db, redis, prometheus, grafana) for easy local development.

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Backend | Go (Golang) | Core application language |
| API Framework | Chi v5 | Routing and middleware |
| Database | MySQL 8.0 | Relational data storage |
| Caching | Redis | In-memory caching for performance |
| Migrations | golang-migrate | Database schema management |
| Logging | slog, uuid | Structured logging and request tracing |
| Monitoring | Prometheus, Grafana | Metrics collection and visualization |
| Containerization | Docker, Docker Compose | Environment setup and deployment |

## Project Structure

The project follows a standard layered architecture to separate concerns and improve maintainability.

```
.
├── cmd/api/
│   └── main.go              # Application entry point, DI wiring, server startup.
├── db/
│   └── migrations/          # SQL database migration files.
├── devops/
│   └── prometheus/
│       └── prometheus.yml   # Prometheus configuration.
├── internal/
│   ├── config/              # Configuration loading from environment variables.
│   ├── domain/              # Core data models and repository interfaces.
│   ├── logger/              # Structured logger setup.
│   ├── repository/          # Data access layer (interacts with the database and cache).
│   ├── server/              # HTTP server, routing, handlers, and middleware.
│   ├── service/             # Business logic layer.
│   └── worker/              # Asynchronous worker pool for background jobs.
├── .env                     # Local environment variables (ignored by Git).
├── .gitignore               # Files and directories ignored by Git.
├── .dockerignore            # Files and directories ignored by Docker builds.
├── Dockerfile               # Multi-stage Dockerfile for building the application.
├── docker-compose.yml       # Defines and configures all services for local development.
├── go.mod                   # Go module definitions.
└── go.sum                   # Go module checksums.
```

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go (for running the migration tool)
- A curl-compatible terminal.

### 1. Setup

#### a. Clone the repository and navigate into the directory.

#### b. Configure Environment Variables
Create a `.env` file from the example provided.

```bash
# App Environment: development, staging, or production
ENV=development
PORT=8080
JWT_SECRET="SECRET KEY EXAMPLE"

DATABASE_DSN="USERNAME:PASSWORD@tcp(db:3306)/DB_NAME?parseTime=true"

# Variables required by the MySQL container in docker-compose.yml
DB_NAME= DATABASE NAME
DB_PASSWORD= PASSWORD

REDIS_ADDRESS="redis:6379"
REDIS_PASSWORD=""
```

This file contains all necessary configuration, including database credentials and your JWT secret. The defaults are set up to work with Docker Compose.

#### c. Run Database Migrations
This crucial one-time step creates all the necessary tables in your database.

First, install the migration tool:

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Then, with Docker running, apply the migrations. (Ensure your docker-compose up command has been run at least once to create the db container).

```bash
# Replace YOUR_PASSWORD with the DB_PASSWORD from your .env file
migrate -path db/migrations -database "mysql://root:YOUR_PASSWORD@tcp(localhost:3306)/mydatabase" up
```

### 2. Build and Run the Application

Launch the entire application stack with a single command:

```bash
docker-compose up --build -d
```

### 3. Create an Admin User

For security, user registration creates standard user roles. To create an admin, you must promote a user manually after they have registered.

1. Register a new user via the API endpoint who you want to make an admin.
2. Promote the user using one of the following methods:

#### Method 1: Command Line (Recommended)
Execute this command in your terminal, replacing the email with your new user's email.

```bash
# Replace YOUR_PASSWORD with your DB_PASSWORD
docker-compose exec db mysql -u root -pYOUR_PASSWORD mydatabase -e "UPDATE users SET role='admin' WHERE email='admin@example.com';"
```

#### Method 2: MySQL Workbench
1. Connect to your database running at localhost:3306 with the credentials from your .env file.
2. Open the users table.
3. Find the user you want to promote and change their role column from user to admin.
4. Apply the changes.

## API Testing Commands

The following curl commands can be used to test all major functionalities. Remember to replace placeholders like `<YOUR_JWT_TOKEN>` and `<USER_ID>`.

### Authentication

**Register a User:**
```bash
curl -X POST -H "Content-Type: application/json" -d '{"username":"testuser", "email":"user@example.com", "password":"password123"}' http://localhost:8080/api/v1/auth/register
```

**Login to get a JWT Token:**
```bash
curl -X POST -H "Content-Type: application/json" -d '{"email":"user@example.com", "password":"password123"}' http://localhost:8080/api/v1/auth/login
```

### Transactions (Requires Authentication)

**Transfer Funds:**
```bash
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <YOUR_JWT_TOKEN>" -d '{"to_user_id": 2, "amount": 50.00}' http://localhost:8080/api/v1/transactions/transfer
```

**Credit an Account (Admin Only):**
```bash
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <ADMIN_JWT_TOKEN>" -d '{"user_id": 2, "amount": 1000.00}' http://localhost:8080/api/v1/transactions/credit
```

**Get Transaction History:**
```bash
curl -H "Authorization: Bearer <YOUR_JWT_TOKEN>" http://localhost:8080/api/v1/transactions/history
```

### Balances (Requires Authentication)

**Get Current Balance:**
```bash
curl -H "Authorization: Bearer <YOUR_JWT_TOKEN>" http://localhost:8080/api/v1/balances/current
```

## Monitoring

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (Login: admin / admin)
