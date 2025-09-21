FROM golang:1.25.0 AS build

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o nodebot

FROM alpine:3.22.1

ARG GCOMPAT_VERSION="1.1.0-r4"
RUN apk add --no-cache gcompat=${GCOMPAT_VERSION} && \
      rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build /build/nodebot nodebot
ENTRYPOINT ["/app/nodebot"]
