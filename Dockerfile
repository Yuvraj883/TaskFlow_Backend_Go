FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/main ./cmd \
  && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  && cp /go/bin/migrate /out/migrate

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/main /app/main
COPY --from=builder /out/migrate /app/migrate
COPY migrations /app/migrations
COPY scripts/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/main /app/migrate /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]