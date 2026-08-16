FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /issue-pm ./cmd/server

FROM alpine:3.20
RUN adduser -D -g '' app
WORKDIR /app
COPY --from=build /issue-pm /app/issue-pm
COPY migrations /app/migrations
COPY web /app/web
USER app
EXPOSE 8080
ENTRYPOINT ["/app/issue-pm"]

