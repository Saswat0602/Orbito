# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.20 AS runner

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/api /app/bin/api

EXPOSE 3000

ENV PORT=3000 \
    APP_ENV=production

ENTRYPOINT ["/app/bin/api"]
