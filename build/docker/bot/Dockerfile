# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.2
ARG ALPINE_VERSION=3.22

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /src

# Copy dependency manifests first to leverage layer caching.
COPY go.mod go.sum ./
COPY vendor/ vendor/

# Copy the rest of the source.
COPY . .

# Build the bot binary. CGO is disabled so the result is a fully static binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/ge-bot ./cmd/bot

# ---- Runtime stage ----
FROM alpine:${ALPINE_VERSION} AS final

COPY --from=builder /out/ge-bot /ge-bot

# The bot token must be supplied at runtime via the TELEGRAM_BOT_TOKEN env var.
ENV TELEGRAM_BOT_TOKEN=""

ENTRYPOINT ["/ge-bot"]
