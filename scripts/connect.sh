#!/usr/bin/env bash
# Port-forward to the Cartograph database and optionally launch psql.
set -euo pipefail

NAMESPACE="${NAMESPACE:-cartograph}"
CLUSTER_NAME="${CLUSTER_NAME:-cartograph-db}"

usage() {
    echo "Usage: $0 [--namespace NS] [--psql]"
    echo ""
    echo "Options:"
    echo "  --namespace NS   Kubernetes namespace (default: cartograph)"
    echo "  --psql           Launch psql after port-forward is ready"
    exit 1
}

LAUNCH_PSQL=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace) NAMESPACE="$2"; shift 2 ;;
        --psql) LAUNCH_PSQL=true; shift ;;
        --help) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

echo "Namespace: ${NAMESPACE}"

PRIMARY_POD=$(oc get pods -n "${NAMESPACE}" \
    -l "postgres-operator.crunchydata.com/cluster=${CLUSTER_NAME},postgres-operator.crunchydata.com/role=master" \
    -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [[ -z "${PRIMARY_POD}" ]]; then
    echo "ERROR: Could not find primary pod for cluster ${CLUSTER_NAME} in namespace ${NAMESPACE}"
    exit 1
fi

echo "Primary pod: ${PRIMARY_POD}"

SECRET_NAME="${CLUSTER_NAME}-pguser-${CLUSTER_NAME}"
DB_USER=$(oc get secret -n "${NAMESPACE}" "${SECRET_NAME}" -o jsonpath='{.data.user}' | base64 -d)
DB_NAME=$(oc get secret -n "${NAMESPACE}" "${SECRET_NAME}" -o jsonpath='{.data.dbname}' | base64 -d)
DB_PASS=$(oc get secret -n "${NAMESPACE}" "${SECRET_NAME}" -o jsonpath='{.data.password}' | base64 -d)

LOCAL_PORT=15432

echo "Starting port-forward to ${PRIMARY_POD}:5432 on localhost:${LOCAL_PORT}..."
oc port-forward -n "${NAMESPACE}" "pod/${PRIMARY_POD}" "${LOCAL_PORT}:5432" &
PF_PID=$!
trap "kill ${PF_PID} 2>/dev/null" EXIT

sleep 2

echo ""
echo "Connection string:"
echo "  CARTOGRAPH_DB_URL=postgres://${DB_USER}:${DB_PASS}@localhost:${LOCAL_PORT}/${DB_NAME}"
echo ""
echo "psql command:"
echo "  PGPASSWORD='${DB_PASS}' psql -h localhost -p ${LOCAL_PORT} -U ${DB_USER} -d ${DB_NAME}"
echo ""

if [[ "${LAUNCH_PSQL}" == "true" ]]; then
    PGPASSWORD="${DB_PASS}" psql -h localhost -p "${LOCAL_PORT}" -U "${DB_USER}" -d "${DB_NAME}"
else
    echo "Port-forward is running. Press Ctrl+C to stop."
    wait ${PF_PID}
fi
