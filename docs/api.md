# API

All responses use `{ "code": 0, "message": "ok", "data": ... }`. Send `Authorization: Bearer <token>` on protected endpoints.

## Auth
`POST /api/v1/auth/register` body `{email,name,password,role?}`. `POST /api/v1/auth/login` body `{email,password}` returns `{token,user}`.

## Projects
`POST /api/v1/projects` body `{name,key,description}`. `GET /api/v1/projects?limit=20&offset=0`.

## Issues
`POST /api/v1/projects/:projectId/issues` body `{title,description,status,priority,assignee_id,labels:["bug"],custom_fields:{"estimate":3}}`.
`GET /api/v1/projects/:projectId/issues?status=open&assignee_id=<uuid>&label=bug&limit=20&offset=0&sort=created_at`.
`GET /api/v1/issues/:id`, `PATCH /api/v1/issues/:id`, `DELETE /api/v1/issues/:id`.
`POST /api/v1/issues/:id/comments` body `{body}`; `GET /api/v1/issues/:id/comments`.

## Permission model

A normal user can only see projects they created or projects in `project_members`. Project creators and admins can add members. Issue creation and assignment require project membership; issue, comment, and comment deletion are checked against the containing project. Comment deletion is restricted to the author or an administrator.

## OpenAPI

The complete OpenAPI 3.0 definition is in [openapi.yaml](openapi.yaml) and can be imported directly into Swagger UI, Insomnia, or Postman.
