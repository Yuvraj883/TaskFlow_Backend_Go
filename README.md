# TaskFlow (Backend-Only Submission)

## 1. Overview

TaskFlow is a task management backend with JWT authentication, projects, and tasks.  
This submission targets the **Backend Engineer** role and includes:

- Go + Gin REST API
- PostgreSQL relational schema
- SQL migrations (up/down) using `golang-migrate`
- Dockerized local setup with automatic migration + seed execution
- Integration tests for auth flow
- Postman collection for API verification in `docs/taskflow-backend.postman_collection.json`

## 2. Architecture Decisions

- **Thin handlers + SQL-first data access:** I chose raw SQL with `pgx` to keep full control over queries and schema behavior.
- **JWT auth middleware:** all non-auth routes require bearer token validation with 24h expiry claims.
- **Ownership and access control:** project/task actions enforce owner/assignee/creator constraints at the query level.
- **Migration-driven schema:** no auto-migrate; schema changes are explicit and reviewable in versioned SQL files.
- **Tradeoffs:** no ORM, no API docs generator, and no frontend in this repo because this is a backend-only submission.

## 3. Running Locally

Assuming Docker is installed:

```bash
git clone https://github.com/your-name/taskflow-backend
cd taskflow-backend
cp .env.example .env
docker compose up --build
```

API base URL:

- `http://localhost:8080`

Health check:

- `GET http://localhost:8080/health`

Run tests:

```bash
go test ./...
```

## 4. Running Migrations

Migrations run automatically when the backend container starts via `scripts/entrypoint.sh`.

Manual commands (if needed):

```bash
docker compose exec backend /app/migrate -path /app/migrations -database "$DATABASE_URL" up
docker compose exec backend /app/migrate -path /app/migrations -database "$DATABASE_URL" down 1
```

## 5. Test Credentials

Seed data is applied through migration `000004_seed_data.up.sql`.

- Email: `test@example.com`
- Password: `password123`

## 6. API Reference

All responses are JSON.  
All non-auth endpoints require:

`Authorization: Bearer <token>`

### Auth

- `POST /auth/register`
- `POST /auth/login`

Example register request:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "secret123"
}
```

Example login response:

```json
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "Jane Doe",
    "email": "jane@example.com"
  }
}
```

### Projects

- `GET /projects`
- `POST /projects`
- `GET /projects/:id`
- `PATCH /projects/:id`
- `DELETE /projects/:id`
- `GET /projects/:id/stats` (bonus)

Example create project request:

```json
{
  "name": "New Project",
  "description": "Optional description"
}
```

### Tasks

- `GET /projects/:id/tasks?status=todo&assignee=<uuid>`
- `POST /projects/:id/tasks`
- `PATCH /tasks/:id`
- `DELETE /tasks/:id`

Example create task request:

```json
{
  "title": "Design homepage",
  "description": "Optional details",
  "priority": "high",
  "assignee_id": "uuid",
  "due_date": "2026-04-20"
}
```

### Standard Error Responses

Validation error (`400`):

```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required"
  }
}
```

Unauthenticated (`401`):

```json
{
  "error": "unauthorized"
}
```

Forbidden (`403`):

```json
{
  "error": "forbidden"
}
```

Not found (`404`):

```json
{
  "error": "not found"
}
```

Postman collection:

- `docs/taskflow-backend.postman_collection.json`

## 7. What I'd Do With More Time

- Add richer integration tests for project/task authorization and edge-case validation.
- Add OpenAPI/Swagger for contract-level documentation and easier API review.
- Improve observability with request IDs, trace correlation, and metrics.
- Add role-based access control and invite-based project collaboration.
