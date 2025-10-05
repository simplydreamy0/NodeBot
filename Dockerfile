FROM golang:1.25.0 AS build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o nodebot

FROM debian:trixie-20250929-slim AS migrate
ARG CURL_VERSION="8.*"
ARG GOMIGRATE_VERSION="v4.19.0"
WORKDIR /build
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl=${CURL_VERSION}
RUN ARCH="$(dpkg --print-architecture)" && \
  curl -sL "https://github.com/golang-migrate/migrate/releases/download/${GOMIGRATE_VERSION}/migrate.linux-${ARCH}.tar.gz" -o migrate.tar.gz
RUN tar -xzf migrate.tar.gz


FROM alpine:3.22.1
ARG GCOMPAT_VERSION="1.1.0-r4"
RUN apk add --no-cache gcompat=${GCOMPAT_VERSION} && \
      rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build /build/nodebot nodebot
WORKDIR /migrations
COPY --from=build /build/internal/db/migrations migrations
COPY --from=migrate /build/migrate /usr/local/bin/migrate

ENTRYPOINT ["/app/nodebot"]
