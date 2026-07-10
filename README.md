# Let It Call

A SvelteKit and Go scheduling application. The production image compiles the Svelte portal, embeds it in the Go binary, and serves the UI and JSON API from one HTTP process.

## Run locally

Start the API on the portal's development-proxy port:

```sh
cd api
go tool godotenv -f .env.local go run ./cmd/server
```

In a second terminal, start the portal:

```sh
cd portal
npm install
npm run dev -- --host 127.0.0.1 --port 41783 --strictPort
```

Opening the repository in VS Code automatically starts both development tasks on ports `41783` (portal) and `41784` (API). The automatic API task loads the committed `api/.env.local` settings through the pinned `godotenv` Go tool.

There is no signup route. When the users table is empty, the API creates its first user from `FIRSTUSER__CREDENTIALS__EMAIL` and `FIRSTUSER__CREDENTIALS__PASSWORD`. Later users are created from Dashboard → Users.

## Docker

```sh
docker build -t letitcall .
docker run --rm -p 8080:80 \
  -v letitcall-data:/data \
  -e FIRSTUSER__CREDENTIALS__EMAIL=owner@example.com \
  -e FIRSTUSER__CREDENTIALS__PASSWORD='replace-with-at-least-12-characters' \
  letitcall
```

The server listens on `HTTP__PORT` (default `80`). LevelDB data is stored at `STORAGE__LEVELDB__PATH` (default `./data`, `/data` in Docker). Each logical table has its own LevelDB database.

## Configuration

All backend settings use structured uppercase environment variables:

- `HTTP__PORT` (default `80`)
- `HTTP__BASE__PATH` (optional; must start with `/`, for example `/letitcall`)
- `STORAGE__LEVELDB__PATH`
- `FIRSTUSER__CREDENTIALS__EMAIL`, `FIRSTUSER__CREDENTIALS__PASSWORD`
- `LOGIN__SESSION__TTL`
- `LOGIN__PASSWORD__MAX_ATTEMPTS`, `LOGIN__PASSWORD__LOCKOUT`
- `LOGIN__OAUTH__GOOGLE__CLIENT_ID`, `LOGIN__OAUTH__GOOGLE__CLIENT_SECRET`

To enable Google OAuth, set both Google settings and add this authorized redirect URI to the Google client:

```text
https://your-host.example{HTTP__BASE__PATH}/api/auth/google/callback
```

For example, a base path of `/letitcall` produces `https://your-host.example/letitcall/api/auth/google/callback`. Omit the base path portion when the application is served at `/`. A TLS-terminating reverse proxy must preserve the request host and set `X-Forwarded-Proto: https` so the server forms the same redirect URI.

The requested scopes include identity and permission to manage Google Calendar events. OAuth state uses PKCE. Google tokens are encrypted with a random key generated on first use and kept in `google-token.key` under `STORAGE__LEVELDB__PATH`; persist the data directory across restarts.

## Test and publish

```sh
cd api && go test ./...
cd ../portal && npm run check && npm run build
```

`publish.sh` creates and pushes multi-platform version and `latest` tags in one Buildx invocation:

```sh
./publish.sh 0.1.0 ghcr.io/example/letitcall
```
