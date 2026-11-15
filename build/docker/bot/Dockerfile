# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27.3
ARG ALPINE_VERSION=3.24

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

# Run as a non-root user. /data holds the user store (USER_STORE_PATH, default users.json);
# a new named volume mounted there copies this ownership.
ARG BOT_UID=10001
ARG BOT_GID=10001

RUN addgroup -S -g "${BOT_GID}" bot && adduser -S -u "${BOT_UID}" -G bot -H -h /data bot && \
    mkdir -p /data && chown "${BOT_UID}:${BOT_GID}" /data

COPY --from=builder /out/ge-bot /ge-bot

WORKDIR /data
USER ${BOT_UID}:${BOT_GID}
VOLUME ["/data"]

# The bot token must be supplied at runtime via the TELEGRAM_BOT_TOKEN env var.
ENTRYPOINT ["/ge-bot"]
