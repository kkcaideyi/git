# Issue-driven lightweight project management backend

Go 1.22 service using Gin, GORM and PostgreSQL. It provides project isolation, issue lifecycle operations, flexible JSON custom fields, JWT authentication, role claims, filtering/pagination, soft deletion and an audit-log table.

## Architecture and data model

`controller` handles HTTP binding and response codes; `service` contains business rules; `repository` owns GORM queries; `model` defines persistence entities. `middleware` provides JWT, logging and CORS. Issues belong to projects and optionally reference a user assignee. `labels` and `custom_fields` are JSONB so new fields do not require schema changes. All core entities have `deleted_at`; PostgreSQL indexes cover project/status, assignee, labels and audit lookups.

## Run locally

1. Copy `.env.example` to `.env` and set a strong `JWT_SECRET` and a local `DATABASE_URL`.
2. Ensure PostgreSQL is running and apply `migrations/001_init.sql` (or run the service in non-production mode, where GORM auto-migrates).
3. Run `go mod tidy` then `go run ./cmd/server`.
4. Health check: `GET http://localhost:8080/health`.

## Docker

`docker compose up --build` starts PostgreSQL and the API on port 8080. The migration is mounted into PostgreSQL's initialization directory.

## API examples

Register: `curl -X POST localhost:8080/api/v1/auth/register -H 'Content-Type: application/json' -d '{"email":"admin@example.com","name":"Admin","password":"password123","role":"admin"}'`.

Login returns a JWT. Use it to create a project, then create issues. Issue updates accept any subset of `title`, `description`, `status`, `priority`, `assignee_id`, `labels`, and `custom_fields`. List endpoints return `items`, `total`, `limit`, and `offset`.

See [docs/api.md](docs/api.md) for the complete endpoint list and request shapes.

## 前端工作台

服务启动后直接访问 `http://localhost:8080/`。前端是 Go 服务内置的静态页面，包含登录/注册、项目切换、新建项目、新建 Issue、状态/优先级/标签筛选、搜索、分页、Issue 详情和评论。页面文件位于 `web/index.html`、`web/static/styles.css` 和 `web/static/app.js`。
## Quality checks

Run `go test ./...` for service, JWT, and route integration tests. The repository uses project/status/time composite indexes, partial indexes for non-deleted rows, and a GIN index for JSONB labels. Pagination is bounded to 1-100 items and uses deterministic secondary ordering by UUID.

## 本地无 Docker 开发模式

如果本机没有 PostgreSQL 或 Docker，可直接使用项目内 SQLite 开发数据库：`.env` 已配置为 `sqlite://data/issuepm.db`，运行 `go run ./cmd/server` 即可。生产环境和 Docker Compose 仍使用 PostgreSQL；SQLite 仅用于本地开发和演示。
