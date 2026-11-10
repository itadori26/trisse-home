# syntax=docker/dockerfile:1

# Stage 1: build the Svelte single-page application.
FROM node:26.13.0-alpine AS web
WORKDIR /src/web
RUN mkdir -p ../internal/platform/webui
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

# Stage 2: build a static Go binary with the SPA embedded.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY --from=web /src/internal/platform/webui/dist ./internal/platform/webui/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/trisse-home ./cmd/server \
    && mkdir -p /out/data

# Stage 3: minimal runtime image without shell, running as non-root.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/trisse-home /trisse-home
COPY --from=build --chown=65532:65532 /out/data /data
ENV DATABASE_PATH=/data/trisse-home.db \
    LISTEN_ADDR=:8090
EXPOSE 8090
VOLUME ["/data"]
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/trisse-home", "healthcheck"]
ENTRYPOINT ["/trisse-home"]
CMD ["serve"]
