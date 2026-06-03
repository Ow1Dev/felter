# Shared Go production image.
# Build with:
#   docker build -f infra/prod/Dockerfile.go --build-arg APP_NAME=<cmd_name> -t <image>:<tag> .
# Example:
#   docker build -f infra/prod/Dockerfile.go --build-arg APP_NAME=proxy -t felter/proxy:latest .

FROM golang:1.25-alpine AS builder

ARG APP_NAME
RUN test -n "$APP_NAME" || (echo "APP_NAME build arg is required" && exit 1)

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -buildvcs=false -o /bin/app ./cmd/${APP_NAME}

# ─── Final stage ───
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /app

ENTRYPOINT ["/app"]
