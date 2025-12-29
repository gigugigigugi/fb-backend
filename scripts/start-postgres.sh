#!/usr/bin/env bash
set -euo pipefail

PGDATA="${HOME}/pgdata"
PGPORT="${PGPORT:-5432}"
PGHOST="127.0.0.1"
SOCKET_DIR="${PGDATA}"

APP_DB="${APP_DB:-football_db}"
APP_USER="${APP_USER:-user}"
APP_PASS="${APP_PASS:-password}"

mkdir -p "$PGDATA"

if [ ! -f "$PGDATA/PG_VERSION" ]; then
  initdb -D "$PGDATA" --username=postgres >/dev/null
fi

# Start (idempotent)
if ! pg_ctl -D "$PGDATA" status >/dev/null 2>&1; then
  pg_ctl -D "$PGDATA" -o "-h ${PGHOST} -p ${PGPORT} -k ${SOCKET_DIR}" -l "$PGDATA/logfile" start
fi

# Ensure user/db
psql -h "${PGHOST}" -p "${PGPORT}" -U postgres -d postgres <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${APP_USER}') THEN
    CREATE ROLE "${APP_USER}" LOGIN PASSWORD '${APP_PASS}';
  END IF;
END
\$\$;

SELECT 'CREATE DATABASE ${APP_DB} OWNER "${APP_USER}"'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${APP_DB}')
\\gexec

GRANT ALL PRIVILEGES ON DATABASE ${APP_DB} TO "${APP_USER}";
SQL

echo "Postgres ready."
echo "DSN: postgres://${APP_USER}:******@localhost:${PGPORT}/${APP_DB}?sslmode=disable"
