# syntax=docker/dockerfile:1.4
# Base image matches plugin-cli exactly — s6-overlay v3 included.
FROM ghcr.io/home-assistant/base:3.23-2026.05.0@sha256:3036cd72ba7755263cd103acc77cb0b438462720c2c8c23b7c2b52e52d7f4b50 AS base

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

# =============================================================================
# Build stage — compile the Go binary
# =============================================================================
FROM golang:1.22-alpine AS builder

WORKDIR /usr/src/lva-cli

# Cache module downloads separately from source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILD_VERSION="0.0.1.dev0"
ARG TARGETOS=linux
ARG TARGETARCH

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
        -ldflags="-s -w -X main.version=${BUILD_VERSION}" \
        -o /usr/bin/lva \
        .

# =============================================================================
# Final image — identical approach to plugin-cli:
#   1. Install rlwrap from source (same version as HA)
#   2. Copy lva binary to /usr/bin/lva
#   3. Copy rootfs (just cli.sh)
#   4. CMD = cli.sh, which execs into `lva banner`
# =============================================================================
FROM base

ENV S6_VERBOSITY=0
ENV TERM=xterm-256color
ENV HOME=/root

WORKDIR /usr/src

ARG BUILD_VERSION="0.0.1.dev0"
ARG RLWRAP_VERSION=0.46.1

# Install rlwrap — same build steps as plugin-cli Dockerfile.
RUN apk add --no-cache --virtual .build-deps \
        build-base \
        readline-dev \
        ncurses-dev \
    && curl -L -s "https://github.com/hanslub42/rlwrap/releases/download/${RLWRAP_VERSION}/rlwrap-${RLWRAP_VERSION}.tar.gz" \
       | tar zxvf - -C /usr/src/ \
    && cd rlwrap-${RLWRAP_VERSION} \
    && ./configure \
    && make \
    && make install \
    && apk del .build-deps \
    && rm -rf /usr/src/*

COPY --from=builder /usr/bin/lva /usr/bin/lva

# rootfs only contains /usr/bin/cli.sh
COPY rootfs /
RUN chmod +x /usr/bin/cli.sh

WORKDIR /

LABEL \
    io.lva.type="cli" \
    org.opencontainers.image.title="LVA CLI" \
    org.opencontainers.image.description="LVA OS command-line interface container" \
    org.opencontainers.image.authors="aryanhasgithub" \
    org.opencontainers.image.url="https://github.com/aryanhasgithub/lva-os" \
    org.opencontainers.image.licenses="Apache License 2.0" \
    org.opencontainers.image.version="${BUILD_VERSION}"
