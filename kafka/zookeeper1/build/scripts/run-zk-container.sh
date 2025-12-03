#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasını yükle (Bir üst klasördeki environments klasöründen)
ENV_FILE=$(realpath "../environments/zookeeper1.env")
# ENV_FILE=$(cd ../environments && pwd)/zookeeper.env

# .env dosyasını yükleme
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
else
    echo "Error: .env file not found at $ENV_FILE"
    exit 1
fi

# Gerekli değişkenlerin kontrolü
REQUIRED_VARS=("CONTAINER_NAME" "CONTAINER_NETWORK_NAME" "ZK_CONFIG_EXTERNAL_PATH" "ZK_CONFIG_INTERNAL_PATH"
               "ZK_DATA_EXTERNAL_PATH" "ZK_DATA_INTERNAL_PATH" "ZK_LOG_EXTERNAL_PATH" "ZK_LOG_INTERNAL_PATH"
               "ZK_DYNAMIC_CONFIG_EXTERNAL_PATH" "ZK_DYNAMIC_CONFIG_INTERNAL_PATH" "CLIENT_HOST_PORT"
               "CLIENT_CONTAINER_PORT" "ADMIN_SERVER_HOST_PORT" "ADMIN_SERVER_CONTAINER_PORT" "HEALTH_SUCCESS_MSG"
               "HEALTH_ERROR_MSG" "HEALTH_INTERVAL" "HEALTH_TIMEOUT" "CONTAINER_HOST_NAME"
               "ZK_GROUP_URL" "ZK_ENDPOINT_URL" "ZK_HEALTH_EXPECTED_ERROR" "HEALTH_RETRIES"
               "RESTART_POLICY" "RESTART_MAX_RETRIES" "CONTAINER_IMAGE")


for VAR in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!VAR}" ]; then
        echo "Error: Required variable '$VAR' is not set or empty in $ENV_FILE"
        exit 1
    fi
done

# -p "$ZK_FOLLOWER_SERVER_HOST_PORT:$ZK_FOLLOWER_SERVER_CONTAINER_PORT" \
# -p "$ZK_LEADER_SERVER_HOST_PORT:$ZK_LEADER_SERVER_CONTAINER_PORT" \
# --health-cmd "curl -s \"$CONTAINER_HOST_NAME:$ADMIN_SERVER_CONTAINER_PORT$ZK_GROUP_URL$ZK_ENDPOINT_URL\" | grep -q '$ZK_HEALTH_EXPECTED_ERROR' && echo \"$HEALTH_SUCCESS_MSG\" || (echo \"$HEALTH_ERROR_MSG\" && exit 1)" \
echo "Starting container $CONTAINER_NAME..."
# Docker container başlatma
echo "Starting container $CONTAINER_NAME..."
if docker run -d \
    --network "$CONTAINER_NETWORK_NAME" \
    --name "$CONTAINER_NAME" \
    --hostname "$CONTAINER_HOST_NAME" \
    -v "$ZK_CONFIG_EXTERNAL_PATH:$ZK_CONFIG_INTERNAL_PATH" \
    -v "$ZK_DYNAMIC_CONFIG_EXTERNAL_PATH:$ZK_DYNAMIC_CONFIG_INTERNAL_PATH" \
    -v "$ZK_DATA_EXTERNAL_PATH:$ZK_DATA_INTERNAL_PATH" \
    -v "$ZK_LOG_EXTERNAL_PATH:$ZK_LOG_INTERNAL_PATH" \
    --env-file "$ENV_FILE" \
    -p "$CLIENT_HOST_PORT:$CLIENT_CONTAINER_PORT" \
    -p "$ADMIN_SERVER_HOST_PORT:$ADMIN_SERVER_CONTAINER_PORT" \
    --health-cmd "if curl -s \"$CONTAINER_HOST_NAME:$ADMIN_SERVER_CONTAINER_PORT$ZK_GROUP_URL$ZK_ENDPOINT_URL\" | grep -q '$ZK_HEALTH_EXPECTED_ERROR'; then 
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

