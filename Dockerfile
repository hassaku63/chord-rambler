# syntax=docker/dockerfile:1
# Two images from one file, picked with --target (ADR-0010):
#   api  the Go server on a minimal runtime, as a non-root user
#   web  nginx serving the built screen and passing /api to the api container
# Build them through `make image` or `make up`: the Go and Node versions come
# from .tool-versions as build arguments, so this file holds no copy of them.

# The placeholders fail the build with "not found" when the arguments are missing.
ARG GO_VERSION=set-by-make
ARG NODE_VERSION=set-by-make

# --- frontend: build the screen -----------------------------------------------
FROM node:${NODE_VERSION}-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# --- backend: build the API server ---------------------------------------------
FROM golang:${GO_VERSION}-alpine AS backend
ARG VERSION=dev
WORKDIR /src/backend
COPY backend/go.mod backend/go.su[m] ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY backend/ ./
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/hassaku63/chord-rambler/backend/internal/httpapi.Version=${VERSION}" \
      -o /out/server ./cmd/server

# --- web: nginx and the built screen -------------------------------------------
# At start the image renders /etc/nginx/templates into conf.d from PORT,
# API_UPSTREAM (a URL such as http://api:8081) and RESOLVER (the DNS that
# resolves it). The image runs nginx as a non-root user.
FROM nginxinc/nginx-unprivileged:1.31-alpine@sha256:b5047b9b76845eac124358373c6852377c52e4b05834a47647e955dd91cb8f1e AS web
COPY --from=frontend /src/frontend/dist /usr/share/nginx/html
COPY deploy/nginx/templates/ /etc/nginx/templates/
ENV PORT=8080 RESOLVER=127.0.0.11
EXPOSE 8080

# --- api: the Go server (the default target) ----------------------------------
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS api
WORKDIR /app
COPY --from=backend /out/server /app/server
ENV PORT=8080
EXPOSE 8080
# The image's nonroot user, by number: a platform that checks for a non-root
# user can only verify a numeric one.
USER 65532:65532
ENTRYPOINT ["/app/server"]
