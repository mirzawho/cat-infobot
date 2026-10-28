# syntax=docker/dockerfile:1

# Pinned toolchain: keep in sync with the go directive in go.mod.
FROM golang:1.27.3-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .

# Static binary: the runtime image needs no libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/infobot .

FROM alpine:3.22

# ca-certificates is needed for HTTPS to the Bot API; tzdata is not used.
RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 infobot \
    && adduser -S -D -H -u 10001 -G infobot infobot \
    && mkdir -p /data \
    && chown infobot:infobot /data

COPY --from=build /out/infobot /usr/local/bin/infobot

USER 10001:10001
ENV DATA_DIR=/data

# -healthcheck exits 0 when the polling process holds the instance lock.
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD ["/usr/local/bin/infobot", "-healthcheck"]

ENTRYPOINT ["/usr/local/bin/infobot"]
