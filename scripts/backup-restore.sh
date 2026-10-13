#!/usr/bin/env bash
# Trigger or restore Crunchy PGO backups.
set -euo pipefail

NAMESPACE="${NAMESPACE:-cartograph}"
CLUSTER_NAME="${CLUSTER_NAME:-cartograph-db}"

usage() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  trigger          Trigger an immediate full backup"
    echo "  status           Show backup status"
    echo "  restore          Restore from backup (creates new cluster, requires manual steps)"
    echo ""
    echo "Options:"
    echo "  --namespace NS   Kubernetes namespace (default: cartograph)"
    echo "  --target-time T  For restore: PITR target (e.g. '2026-10-13 14:30:00+00')"
    exit 1
}

COMMAND="${1:-}"
shift || true
TARGET_TIME=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace) NAMESPACE="$2"; shift 2 ;;
        --target-time) TARGET_TIME="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

case "${COMMAND}" in
    trigger)
        echo "Triggering immediate backup..."
        ANNOTATION="postgres-operator.crunchydata.com/pgbackrest-backup=$(date +%Y%m%d%H%M%S)"
        oc annotate -n "${NAMESPACE}" postgrescluster "${CLUSTER_NAME}" --overwrite "${ANNOTATION}"
        echo "Backup triggered. Check status with: $0 status"
        ;;

    status)
        echo "pgBackRest backup status:"
        oc exec -n "${NAMESPACE}" \
            "$(oc get pods -n "${NAMESPACE}" -l "postgres-operator.crunchydata.com/cluster=${CLUSTER_NAME},postgres-operator.crunchydata.com/role=master" -o jsonpath='{.items[0].metadata.name}')" \
            -- pgbackrest info
        ;;

    restore)
        echo "=== RESTORE PROCEDURE ==="
        echo ""
        echo "Crunchy PGO restore creates a NEW PostgresCluster from backup."
        echo "This is a manual process with the following steps:"
        echo ""
        echo "1. Create a restore CR (${CLUSTER_NAME}-restore) that bootstraps from backup:"
        echo ""
        if [[ -n "${TARGET_TIME}" ]]; then
            echo "   Target time: ${TARGET_TIME}"
            cat <<YAML
apiVersion: postgres-operator.crunchydata.com/v1beta1
kind: PostgresCluster
metadata:
  name: ${CLUSTER_NAME}-restore
  namespace: ${NAMESPACE}
spec:
  dataSource:
    postgresCluster:
      clusterName: ${CLUSTER_NAME}
      repoName: repo1
      options:
        - --type=time
        - --target="${TARGET_TIME}"
  # ... (copy remaining spec from postgres-cluster.yaml)
YAML
        else
            cat <<YAML
apiVersion: postgres-operator.crunchydata.com/v1beta1
kind: PostgresCluster
metadata:
  name: ${CLUSTER_NAME}-restore
  namespace: ${NAMESPACE}
spec:
  dataSource:
    postgresCluster:
      clusterName: ${CLUSTER_NAME}
      repoName: repo1
  # ... (copy remaining spec from postgres-cluster.yaml)
YAML
        fi
        echo ""
        echo "2. Apply the restore CR: oc apply -f restore-cr.yaml"
        echo "3. Wait for the restored cluster to be ready"
        echo "4. Verify data in the restored cluster"
        echo "5. If satisfied, delete the old cluster and rename the restore"
        echo "   (or update the cartograph Deployment to point to the new Secret)"
        ;;

    *)
        usage
        ;;
esac
