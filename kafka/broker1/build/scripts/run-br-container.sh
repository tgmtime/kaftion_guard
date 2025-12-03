#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasını yükle (Bir üst klasördeki environments klasöründen)
ENV_FILE=$(realpath "../environments/broker1.env")

# .env dosyasını yükleme
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
else
    echo "Error: .env file not found at $ENV_FILE"
    exit 1
fi

# Gerekli değişkenlerin kontrolü
REQUIRED_VARS=(
    "CONTAINER_NAME" "CONTAINER_NETWORK_NAME"
    "BR_CONFIG_EXTERNAL_PATH" "BR_CONFIG_INTERNAL_PATH"
    "BR_DATA_EXTERNAL_PATH" "BR_DATA_INTERNAL_PATH"
    "BR_LOG_EXTERNAL_PATH" "BR_LOG_INTERNAL_PATH"
    "CLIENT_HOST_PORT" "CLIENT_CONTAINER_PORT"
    "HEALTH_CMD" "HEALTH_INTERVAL" "HEALTH_TIMEOUT"
    "HEALTH_RETRIES" "RESTART_POLICY" "RESTART_MAX_RETRIES"
    "CONTAINER_HOST_NAME" "CONTAINER_IMAGE"
)

for VAR in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!VAR}" ]; then
        echo "Error: Required variable '$VAR' is not set or empty in $ENV_FILE."
        exit 1
    fi
done


# Docker container başlatma
echo "Starting container $CONTAINER_NAME..."
if docker run -d \
    --network "$CONTAINER_NETWORK_NAME" \
    --name "$CONTAINER_NAME" \
    --hostname "$CONTAINER_HOST_NAME" \
    -v "$BR_CONFIG_EXTERNAL_PATH:$BR_CONFIG_INTERNAL_PATH" \
    -v "$BR_DATA_EXTERNAL_PATH:$BR_DATA_INTERNAL_PATH" \
    -v "$BR_LOG_EXTERNAL_PATH:$BR_LOG_INTERNAL_PATH" \
    --env-file "$ENV_FILE" \
    -p "$CLIENT_HOST_PORT:$CLIENT_CONTAINER_PORT" \
    --health-cmd "if ${HEALTH_CMD} >/dev/null 2>&1; then 
                    echo \"✅ ${CONTAINER_NAME} health check passed\"; 
                  else 
                    echo \"❌ ${CONTAINER_NAME} health check failed\" >&2; 
                    exit 1; 
                  fi" \
    --health-interval "$HEALTH_INTERVAL" \
    --health-timeout "$HEALTH_TIMEOUT" \
    --health-retries "$HEALTH_RETRIES" \
    --restart "$RESTART_POLICY:$RESTART_MAX_RETRIES" \
    "$CONTAINER_IMAGE"; then
    echo "Container $CONTAINER_NAME was successfully created and is running."
else
    echo "Error: Failed to create and start container $CONTAINER_NAME."
    exit 1
fi

