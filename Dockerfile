FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /repo-tokens ./cmd/repo-tokens

# Pre-warm tiktoken cache so the action never needs network access.
ENV TIKTOKEN_CACHE_DIR=/cache
RUN mkdir -p /cache && /repo-tokens --json . >/dev/null 2>&1 || true

FROM alpine:3.20
RUN apk add --no-cache git
COPY --from=builder /repo-tokens /repo-tokens
COPY --from=builder /cache /cache
ENV TIKTOKEN_CACHE_DIR=/cache
ENTRYPOINT ["/repo-tokens"]
