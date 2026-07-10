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
pnpm install
pnpm run dev -- --host 127.0.0.1 --port 41783 --strictPort
```

Open portal at `http://127.0.0.1:41783`. 

Default local login credentials is admin:admin.

Opening the repository in VS Code automatically starts both development tasks on ports `41783` (portal) and `41784` (API).

There is no signup route. When the users table is empty, the API creates its first user from `FIRSTUSER__CREDENTIALS__EMAIL` and `FIRSTUSER__CREDENTIALS__PASSWORD` from the `api/.env.local` file. Later users are created from Dashboard → Users.

## Docker

```sh
docker build -t letitcall .
docker run --rm -p 8080:80 \
  -v letitcall-data:/data \
  -e FIRSTUSER__CREDENTIALS__EMAIL=owner@example.com \
  -e FIRSTUSER__CREDENTIALS__PASSWORD='replace-with-at-least-12-characters' \
  letitcall
```

The server listens on `HTTP__PORT` (default `80`). LevelDB data is stored at `STORAGE__LEVELDB__PATH` (`/data` in Docker). Each logical table has its own LevelDB database.

## Configuration

All backend settings use structured uppercase environment variables:

| Name | What it does | Default value |
| --- | --- | --- |
| `HTTP__PORT` | Sets the HTTP server port. | `80` |
| `HTTP__BASE__PATH` | Sets the URL path prefix where the application is served. A configured value must start with `/`, for example `/letitcall`. | Empty (served at `/`) |
| `STORAGE__LEVELDB__PATH` | Sets the directory containing the LevelDB databases. | `./data` (`/data` in Docker) |
| `FIRSTUSER__CREDENTIALS__EMAIL` | Sets the email of the user created when the users table is empty. Must be set together with the first-user password. | Not set |
| `FIRSTUSER__CREDENTIALS__PASSWORD` | Sets the password of the user created when the users table is empty. Must be set together with the first-user email. | Not set |
| `LOGIN__SESSION__TTL` | Sets how long an authenticated session remains valid. | `24h` |
| `LOGIN__PASSWORD__MAX_ATTEMPTS` | Sets the number of failed password attempts allowed before login is temporarily locked. | `5` |
| `LOGIN__PASSWORD__LOCKOUT` | Sets how long a password login remains locked after reaching the failed-attempt limit. | `15m` |
| `LOGIN__OAUTH__GOOGLE__CLIENT_ID` | Sets the Google OAuth client ID and enables Google login when the client secret is also set. | Not set |
| `LOGIN__OAUTH__GOOGLE__CLIENT_SECRET` | Sets the Google OAuth client secret. Must be set together with the client ID. | Not set |

To enable Google OAuth, set both Google settings and add this authorized redirect URI to the Google client:

```text
https://your-host.example{HTTP__BASE__PATH}/api/auth/google/callback
```

For example, a HTTP__BASE__PATH  of `/letitcall` produces `https://your-host.example/letitcall/api/auth/google/callback`. Omit the base path portion when the application is served at `/`. A TLS-terminating reverse proxy must preserve the request host and set `X-Forwarded-Proto: https` so the server forms the same redirect URI.

The requested scopes include identity and permission to manage Google Calendar events. OAuth state uses PKCE. Google tokens are encrypted with a random key generated on first use and kept in `google-token.key` under `STORAGE__LEVELDB__PATH`; persist the data directory across restarts.

## Test and publish

```sh
cd api && go test ./...
cd ../portal && pnpm run check && pnpm run build
```

`publish.sh` creates and pushes multi-platform version and `latest` tags in one Buildx invocation:

```sh
./publish.sh
```

Set `VERSION` and `PACKAGE_NAME` inside the script before publishing to Docker Hub.
