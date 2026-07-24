# Shared Go development image.
# Mount the repository root at /app for live code reloading.
# The compose file sets the CMD to go run ./cmd/<service>.

FROM golang:1.25-alpine

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN GOTOOLCHAIN=auto go install github.com/air-verse/air@latest

EXPOSE 8080 8081 9090 9091 9092

CMD ["sh"]
