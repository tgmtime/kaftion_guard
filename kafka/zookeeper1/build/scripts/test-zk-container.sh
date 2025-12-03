#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasının yolu
ENV_FILE=$(realpath "../environments/zookeeper1.env")

# .env dosyasını yükleme
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
else
    echo "Error: .env file not found at $ENV_FILE"
    exit 1
fi

# CONTAINER_NAME kontrolü
if [ -z "$CONTAINER_NAME" ]; then
    echo "Error: CONTAINER_NAME is not set in .env file."
    exit 1
fi

# Varsayılan deneme sayısı ve bekleme süresi
MAX_RETRIES=${MAX_RETRIES:-10}
RETRY_DELAY=${RETRY_DELAY:-3}

# Değişkenlerin boş olup olmadığını kontrol et
if [ -z "$MAX_RETRIES" ]; then
    echo "Error: MAX_RETRIES is not set or is empty in .env file."
    exit 1
fi

if [ -z "$RETRY_DELAY" ]; then
    echo "Error: RETRY_DELAY is not set or is empty in .env file."
    exit 1
fi


# Gerekli değişkenlerin kontrolü
REQUIRED_VARS=(    
    "ZK_GROUP_URL" "ZK_ENDPOINT_URL"
    "ZK_HEALTH_EXPECTED_ERROR"
    "CONTAINER_NAME" "MAX_RETRIES"
    "RETRY_DELAY" "HEALTH_CHECK_BROKERS"
)

for VAR in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!VAR}" ]; then
        echo "❌ Required variable '$VAR' is not set or empty in $ENV_FILE."
        exit 1
    fi
done



echo "Checking container health for $CONTAINER_NAME..."

# Sağlık kontrol döngüsü
for ((i = 1; i <= MAX_RETRIES; i++)); do
    HEALTH_STATUS=$(docker inspect -f '{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "notfound")

    case "$HEALTH_STATUS" in
        "healthy")
            echo "✅ Container '$CONTAINER_NAME' is healthy. Running health check script..."
            bash zk-health-check.sh
            EXIT_CODE=$?

            if [ $EXIT_CODE -ne 0 ]; then
                echo "❌ Error: Health check script failed with exit code $EXIT_CODE."
                exit $EXIT_CODE
            fi

            echo "✅ Health check passed successfully."
            exit 0
            ;;
        "starting")
            echo "⏳ Container '$CONTAINER_NAME' is starting... Retrying in $RETRY_DELAY seconds ($i/$MAX_RETRIES)"
            sleep "$RETRY_DELAY"
            ;;
        "unhealthy")
            echo "❌ Error: Container '$CONTAINER_NAME' is unhealthy. Please check the logs."
            docker logs "$CONTAINER_NAME" | tail -n 20
            exit 1
            ;;
        "notfound")
            echo "❌ Error: Failed to inspect container '$CONTAINER_NAME'. Is it running?"
            exit 1
            ;;
        *)
            echo "❌ Error: Unexpected container health status: $HEALTH_STATUS"
            exit 1
            ;;
    esac
done

echo "❌ Error: Container '$CONTAINER_NAME' did not become healthy within $((MAX_RETRIES * RETRY_DELAY)) seconds."
exit 1