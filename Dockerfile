# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS portal
RUN corepack enable
WORKDIR /src/portal
COPY portal/package.json portal/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY portal/ ./
RUN pnpm run check && pnpm run build

FROM node:22-alpine AS emails
WORKDIR /src/api
COPY api/package.json api/package-lock.json ./
RUN npm ci
COPY api/render-mjml.sh ./
COPY api/internal/mailing/templates/mjml/ ./internal/mailing/templates/mjml/
RUN ./render-mjml.sh

FROM golang:1.26.4-alpine AS api
WORKDIR /src/api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ ./
COPY --from=emails /src/api/internal/mailing/templates/rendered/ ./internal/mailing/templates/rendered/
COPY --from=portal /src/portal/build/ ./internal/web/static/
RUN mkdir -p /out/data && \
	CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -buildid=" -o /out/letitcall ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=api --chown=65532:65532 /out/letitcall /letitcall
COPY --from=api --chown=65532:65532 /out/data /data

ENV HTTP__PORT=80 \
	STORAGE__LEVELDB__PATH=/data

VOLUME ["/data"]
EXPOSE 80
ENTRYPOINT ["/letitcall"]
