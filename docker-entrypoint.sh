#!/usr/bin/env sh
set -e

echo "[entrypoint] Running migrations..."
node ./scripts/migrate.js || {
  echo "[entrypoint] Migration step failed" >&2
  exit 1
}

echo "[entrypoint] Starting app..."
exec node dist/main.js


