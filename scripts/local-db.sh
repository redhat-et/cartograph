#!/usr/bin/env bash
# scripts/local-db.sh — Start/stop a local Postgres+AGE instance for development.
#
# The Crunchy image has no standard entrypoint, so we run initdb/pg_ctl manually.
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-cartograph-dev}"
IMAGE="${IMAGE:-quay.io/redhat-et/crunchy-postgres-age:16-1.6.0}"
PG_PORT="${PG_PORT:-5432}"
PG_PASSWORD="${PG_PASSWORD:-devpass}"
PG_DATABASE="${PG_DATABASE:-cartograph}"

usage() {
    echo "Usage: $0 {start|stop|status|url}"
    echo ""
    echo "Commands:"
    echo "  start   Start the local Postgres+AGE container"
    echo "  stop    Stop and remove the container"
    echo "  status  Show container status"
    echo "  url     Print the connection URL"
    exit 1
}

start_db() {
    if podman container exists "${CONTAINER_NAME}" 2>/dev/null; then
        echo "Container ${CONTAINER_NAME} already exists."
        echo "Run '$0 stop' first, or '$0 status' to check."
        exit 1
    fi

    echo "Starting ${CONTAINER_NAME}..."
    podman run --rm -d \
        --name "${CONTAINER_NAME}" \
        --entrypoint "" \
        -p "${PG_PORT}:5432" \
        -v "${CONTAINER_NAME}-data:/pgdata:Z" \
        "${IMAGE}" \
        bash -c "
            PGDATA=/pgdata/data
            if [ ! -f \"\$PGDATA/PG_VERSION\" ]; then
                mkdir -p \"\$PGDATA\" && chmod 700 \"\$PGDATA\"
                /usr/pgsql-16/bin/initdb -D \"\$PGDATA\"
                echo \"shared_preload_libraries = 'age'\" >> \"\$PGDATA/postgresql.conf\"
                echo \"listen_addresses = '*'\" >> \"\$PGDATA/postgresql.conf\"
                # Trust local socket, password for TCP (dev only)
                sed -i 's/^host.*127.0.0.1.*/host all all 0.0.0.0\\/0 md5/' \"\$PGDATA/pg_hba.conf\"
                echo \"host all all 0.0.0.0/0 md5\" >> \"\$PGDATA/pg_hba.conf\"
                /usr/pgsql-16/bin/pg_ctl -D \"\$PGDATA\" start
                /usr/pgsql-16/bin/psql -h /tmp -U postgres -c \"ALTER USER postgres PASSWORD '${PG_PASSWORD}';\"
                /usr/pgsql-16/bin/pg_ctl -D \"\$PGDATA\" stop
            fi
            /usr/pgsql-16/bin/postgres -D \"\$PGDATA\"
        "

    echo "Waiting for Postgres to be ready..."
    for i in $(seq 1 30); do
        if podman exec "${CONTAINER_NAME}" /usr/pgsql-16/bin/pg_isready -h /tmp &>/dev/null; then
            break
        fi
        sleep 1
    done

    # Create dev database and AGE extension
    podman exec "${CONTAINER_NAME}" /usr/pgsql-16/bin/psql -h /tmp -U postgres \
        -c "SELECT 1 FROM pg_database WHERE datname = '${PG_DATABASE}'" | grep -q 1 || \
        podman exec "${CONTAINER_NAME}" /usr/pgsql-16/bin/psql -h /tmp -U postgres \
            -c "CREATE DATABASE ${PG_DATABASE};"

    podman exec "${CONTAINER_NAME}" /usr/pgsql-16/bin/psql -h /tmp -U postgres -d "${PG_DATABASE}" \
        -c "CREATE EXTENSION IF NOT EXISTS age;"

    podman exec "${CONTAINER_NAME}" /usr/pgsql-16/bin/psql -h /tmp -U postgres -d "${PG_DATABASE}" \
        -c "ALTER DATABASE ${PG_DATABASE} SET search_path TO \"\\\$user\", public, ag_catalog;"

    echo ""
    echo "Postgres+AGE is ready."
    print_url
}

stop_db() {
    if podman container exists "${CONTAINER_NAME}" 2>/dev/null; then
        echo "Stopping ${CONTAINER_NAME}..."
        podman stop "${CONTAINER_NAME}"
        echo "Stopped."
    else
        echo "Container ${CONTAINER_NAME} is not running."
    fi
}

show_status() {
    if podman container exists "${CONTAINER_NAME}" 2>/dev/null; then
        podman ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    else
        echo "Container ${CONTAINER_NAME} is not running."
    fi
}

print_url() {
    echo "CARTOGRAPH_DB_URL=postgres://postgres:${PG_PASSWORD}@localhost:${PG_PORT}/${PG_DATABASE}"
}

case "${1:-}" in
    start)  start_db ;;
    stop)   stop_db ;;
    status) show_status ;;
    url)    print_url ;;
    *)      usage ;;
esac
