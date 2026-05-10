#!/usr/bin/env bash
set -euo pipefail  # -u catches unbound vars, -o pipefail catches pipe failures

# ── Validate required env vars are set before doing anything ──────────────────
required_vars=(
    RESOURCE_GROUP_NAME
    AZURE_STORAGE_ACCOUNT_NAME
    AZ_KEYVAULT_NAME
    REDIS_NAME
    POSTGRES_PASSWORD
    POSTGRES_HOST
    POSTGRES_USER
    POSTGRES_DB
    COSMOS_USERNAME
    COSMOS_PASSWORD
    COSMOS_CLUSTER_NAME
    RABBITMQ_DEFAULT_PASS
    RABBITMQ_DEFAULT_USER
    JWT_SECRET
)

for var in "${required_vars[@]}"; do
    if [[ -z "${!var:-}" ]]; then
        echo "ERROR: Required environment variable '$var' is not set or empty"
        exit 1
    fi
done

# ── Get Runner IP ─────────────────────────────────────────────────────────────
RUNNER_IP=$(curl -s --fail --max-time 10 api4.ipify.org)
if [[ -z "$RUNNER_IP" ]]; then
    echo "ERROR: Failed to retrieve runner IP"
    exit 1
fi
echo "Runner IP: $RUNNER_IP"


# ── Cleanup trap — always runs on exit ───────────────────────────────────────
cleanup() {
    echo "Closing KeyVault public access..."
    update_keyvault "$AZ_KEYVAULT_NAME" "$RESOURCE_GROUP_NAME" "$RUNNER_IP" Disabled \
        || echo "WARNING: Failed to close KeyVault access — manual cleanup required for $AZ_KEYVAULT_NAME"
}

INTERRUPTED=false
handle_interrupt() {
    INTERRUPTED=true
    echo "Interrupted — stopping uploads"
    exit 1
}

update_keyvault() {
    local AZ_KEYVAULT_NAME=$1
    local RESOURCE_GROUP_NAME=$2
    local RUNNER_IP=$3
    local PUBLIC_ACCESS=$4

    az keyvault update \
        --name "$AZ_KEYVAULT_NAME" \
        --public-network-access "$PUBLIC_ACCESS" \
        --output none

    if [[ "$PUBLIC_ACCESS" == "Enabled" ]]; then
        az keyvault network-rule add \
            --name "$AZ_KEYVAULT_NAME" \
            --ip-address "$RUNNER_IP" \
            --resource-group "$RESOURCE_GROUP_NAME" \
            --output none
    else
        az keyvault network-rule remove \
            --name "$AZ_KEYVAULT_NAME" \
            --ip-address "$RUNNER_IP" \
            --resource-group "$RESOURCE_GROUP_NAME" \
            --output none
    fi
}

# ── Open KeyVault ─────────────────────────────────────────────────────────────
echo "Opening KeyVault: $AZ_KEYVAULT_NAME"
update_keyvault "$AZ_KEYVAULT_NAME" "$RESOURCE_GROUP_NAME" "$RUNNER_IP" Enabled

trap cleanup EXIT
trap handle_interrupt INT TERM 

# ── Get Storage Account Access Key ──────────────────────────────────────────────

AZURE_STORAGE_ACCOUNT_KEY=$(
  az storage account keys list \
    -g "${RESOURCE_GROUP_NAME}" \
    -n "${AZURE_STORAGE_ACCOUNT_NAME}" \
    --query '[0].value' \
    -o tsv
)

if [[ -z "$AZURE_STORAGE_ACCOUNT_KEY" ]]; then
  echo "Failed to retrieve storage account key" >&2
  exit 1
fi

# ── Get Redis HostName and Primary AccessKey ────────────────────────────────────────────────────────────────
REDIS_HOST=$(
    az redisenterprise show \
       --cluster-name "${REDIS_NAME}" \
       -g "${RESOURCE_GROUP_NAME}" \
       --query hostName \
       --output tsv
)

REDIS_PASSWORD=$(
    az redisenterprise database list-keys \
       --cluster-name "${REDIS_NAME}" \
       -g "${RESOURCE_GROUP_NAME}" \
       --query primaryKey \
       --output tsv
)

if [[ -z "$REDIS_HOST" || -z "$REDIS_PASSWORD" ]]; then
  echo "Failed to retrieve redis hostname and access keys" >&2
  exit 1
fi

# ── Secret map ────────────────────────────────────────────────────────────────
declare -A secrets=(
    ["azure-storage-account-name"]="${AZURE_STORAGE_ACCOUNT_NAME}"
    ["azure-storage-account-key"]="${AZURE_STORAGE_ACCOUNT_KEY}"
    ["azure-storage-blob-endpoint"]="https://${AZURE_STORAGE_ACCOUNT_NAME}.blob.core.windows.net/"
    ["rabbitmq-default-pass"]="${RABBITMQ_DEFAULT_PASS}"
    ["rabbitmq-default-user"]="${RABBITMQ_DEFAULT_USER}"
    ["rabbitmq-url"]="amqp://${RABBITMQ_DEFAULT_USER}:${RABBITMQ_DEFAULT_PASS}@sf-rabbitmq-svc:5672/"
    ["postgres-password"]="${POSTGRES_PASSWORD}"
    ["postgres-host"]="${POSTGRES_HOST}"
    ["postgres-user"]="${POSTGRES_USER}"
    ["postgres-database"]="${POSTGRES_DB}"
    ["cosmos-connection-string"]="mongodb+srv://${COSMOS_USERNAME}:${COSMOS_PASSWORD}@${COSMOS_CLUSTER_NAME}.global.mongocluster.cosmos.azure.com/?tls=true&authMechanism=SCRAM-SHA-256&retrywrites=false&maxIdleTimeMS=120000"
    ["redis-hostname"]="${REDIS_HOST}"
    ["redis-password"]="${REDIS_PASSWORD}"
    ["redis-url"]="rediss://${REDIS_HOST}:10000"
    ["jwt-secret"]="${JWT_SECRET}"
)

# ── Upload secrets ────────────────────────────────────────────────────────────
failed_keys=()

for key in "${!secrets[@]}"; do
    if [[ "$INTERRUPTED" == "true" ]]; then
        break
    fi

    if [[ -z "${secrets[$key]}" ]]; then
        echo "WARNING: Skipping '$key' — value is empty"
        failed_keys+=("$key")
        continue
    fi

    echo "Uploading: $key"
    if ! az keyvault secret set \
        --vault-name "$AZ_KEYVAULT_NAME" \
        --name "$key" \
        --value "${secrets[$key]}" \
        --output none; then
        echo "ERROR: Failed to upload '$key'"
        failed_keys+=("$key")
    fi
done

# ── Report failures ───────────────────────────────────────────────────────────
if [[ ${#failed_keys[@]} -gt 0 ]]; then
    echo "ERROR: The following secrets failed to upload: ${failed_keys[*]}"
    exit 1
fi

echo "All secrets uploaded successfully"
# trap fires here automatically, closes KV