FROM golang:1.26

ENV GOTOOLCHAIN=local \
    PGDATA=/var/lib/postgresql/benzhi \
    DATABASE_URL=postgres://postgres@127.0.0.1:5432/benzhi?sslmode=disable

RUN apt-get update \
    && apt-get install -y --no-install-recommends postgresql postgresql-client \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build ./...

RUN printf '%s\n' \
    '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'pg_bin="$(dirname "$(find /usr/lib/postgresql -type f -name pg_ctl | sort -V | tail -n 1)")"' \
    'export PATH="${pg_bin}:${PATH}"' \
    'install -d -m 0700 -o postgres -g postgres "${PGDATA}"' \
    'install -d -o postgres -g postgres /var/run/postgresql' \
    'if [[ ! -s "${PGDATA}/PG_VERSION" ]]; then' \
    '  runuser -u postgres -- initdb -D "${PGDATA}" --auth=trust >/dev/null' \
    'fi' \
    'runuser -u postgres -- pg_ctl -D "${PGDATA}" -l /tmp/benzhi-postgres.log -o "-h 127.0.0.1 -p 5432" -w start >/dev/null' \
    'if ! runuser -u postgres -- psql -h 127.0.0.1 -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='"'"'benzhi'"'"'" | grep -q 1; then' \
    '  runuser -u postgres -- createdb -h 127.0.0.1 benzhi' \
    'fi' \
    'exec "$@"' \
    > /usr/local/bin/benzhi-entrypoint \
    && chmod +x /usr/local/bin/benzhi-entrypoint

ENTRYPOINT ["/usr/local/bin/benzhi-entrypoint"]
CMD ["bash"]
