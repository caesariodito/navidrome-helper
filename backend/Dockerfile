### Builder
FROM golang:1.22-bookworm AS build
WORKDIR /src

# Pre-fetch dependencies for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy sources and build a trimmed binary.
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/navidrome-helper .

### Runtime
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=build /out/navidrome-helper /app/navidrome-helper

ENV PORT=8173 \
    DATA_DIR=/data \
    TEMP_DIR=/data/tmp \
    NAVIDROME_MUSIC_PATH=/music

EXPOSE 8173
VOLUME ["/data", "/music"]

ENTRYPOINT ["/app/navidrome-helper"]
