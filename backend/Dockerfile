FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev ca-certificates

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/server \
    ./main.go


FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/server /app/server

RUN mkdir -p /app/data && chown -R app:app /app

USER app

EXPOSE 8080

ENV HTTP_ADDR=:8080 \
    DATABASE_PATH=/app/data/dashboard.db

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/server"]
