# TaskFlow Backend

## Status

✅ Fully functional backend with Docker support

---

## 1. Overview

TaskFlow is a minimal task management system where users can:

* Register and log in securely
* Create and manage projects
* Create, update, and delete tasks
* Assign tasks to themselves or other users

This backend is built using Go and follows clean architecture principles with JWT-based authentication and PostgreSQL for data persistence.

---

## 2. Tech Stack

* **Language:** Go (Golang)
* **Framework:** Gin
* **Database:** PostgreSQL
* **Authentication:** JWT
* **Password Hashing:** bcrypt
* **Migrations:** golang-migrate
* **Containerization:** Docker & Docker Compose

---

## 3. Architecture Decisions

* **Gin Framework:** Chosen for its simplicity, performance, and minimal overhead for building REST APIs.
* **Raw SQL over ORM:** Used raw SQL queries instead of an ORM for better control, transparency, and performance.
* **JWT Authentication:** Stateless authentication using JWT for scalability and simplicity.
* **Project Structure:** Code is organized into:

  * `handlers/` → request handling
  * `middleware/` → authentication logic
  * `db/` → database connection
  * `utils/` → helper utilities (JWT, etc.)
* **Dockerized Setup:** Ensures consistent local development environment with zero manual DB setup.

---

## 4. Running Locally

### Prerequisites

* Docker installed

### Steps

```bash
git clone https://github.com/your-username/taskflow-backend
cd taskflow-backend
docker compose up --build
```

Server will start at:

```
http://localhost:8080
```

---

## 5. Environment Variables

Create a `.env` file (or use `.env.example`):

```env
DATABASE_URL=postgres://postgres:postgres@db:5432/taskflow?sslmode=disable
JWT_SECRET=supersecretkey
PORT=8080
```

---

## 6. Running Migrations

Migrations are managed using **golang-migrate**.

If not auto-run, execute manually:

```bash
docker exec -it taskflow_backend sh
migrate -path migrations -database "$DATABASE_URL" up
```

---

## 7. Test Credentials

You can either:

### Option 1: Register manually

Use `/auth/register`

### Option 2: (If seeded)

```
Email:    test@example.com
Password: password123
```

---

## 8. API Reference

### Auth

* `POST /auth/register`
* `POST /auth/login`

---

### Projects

* `GET /projects` → List projects
* `POST /projects` → Create project
* `GET /projects/:id` → Get project with tasks
* `PATCH /projects/:id` → Update project
* `DELETE /projects/:id` → Delete project

---

### Tasks

* `GET /projects/:id/tasks` → List tasks (supports filters)

  * `?status=todo|in_progress|done`
  * `?assignee=<user_id>`

* `POST /projects/:id/tasks` → Create task

* `PATCH /tasks/:id` → Update task

* `DELETE /tasks/:id` → Delete task

---

## 9. Error Handling

All responses follow consistent JSON format:

### Validation Error (400)

```json
{
  "error": "validation failed",
  "fields": {
    "field_name": "error message"
  }
}
```

### Unauthorized (401)

```json
{
  "error": "unauthorized"
}
```

### Forbidden (403)

```json
{
  "error": "forbidden"
}
```

### Not Found (404)

```json
{
  "error": "not found"
}
```

---

## 10. Key Features

* JWT-based authentication (24-hour expiry)
* Password hashing using bcrypt
* Protected routes using middleware
* Project ownership validation
* Task assignment (self or others)
* Filtering tasks by status and assignee
* Clean and structured API responses
* Dockerized setup for easy execution
* **Optional Optimistic Locking:** `PATCH` and `DELETE` operations natively support concurrency control (`updated_at`).
* **Request Context Propagation:** Database commands securely bind `c.Request.Context()` strictly cascading HTTP lifecycle bounds preventing ghost queries.
* **Native Token Bucket Rate Limiter:** Auth endpoints restrict traffic strictly tracking IPs in memory mapping preventing credential stuffing automatically.
* **CORS Connectivity:** Standard REST bindings strictly accept React environment requests locally avoiding preflight blocking headers.
* **Production Observability:** Implemented complete `log/slog` structured logging, `os/signal` graceful shutdown lifecycles handling active HTTP contexts natively, and generic `/health` endpoint polling.
* **API Pagination:** Safe metric limiters natively mapping `?page=` and `?limit=` securely bounding JSON size sizes dynamically.
* **Project Statistics Endpoint:** Exposing `GET /projects/:id/stats` directly tracking dynamic SQL aggregated datasets quickly parsing metrics securely constraint-free.
* **Integration Testing Coverage:** Validating constraints bounds against native POST/GET/DELETE operations cleanly.

---

## 11. Future Scope of Enhancements

* Implement role-based access control (RBAC)
* Improve validation (e.g., stricter enums, date validation)
* Enhance task update endpoint with field-level validation
* Implement real-time updates using WebSockets or SSE

---

## 12. Notes

* No secrets are committed in the repository
* `.env` is ignored; use `.env.example` for setup
* Docker ensures the app runs with a single command

---

## 13. How to Test Quickly

```bash
# Register
POST /auth/register

# Login → get token

# Use token in header
Authorization: Bearer <token>

# Create project → create tasks → test full flow
```

---

## 14. Author

Yuvraj Singh
