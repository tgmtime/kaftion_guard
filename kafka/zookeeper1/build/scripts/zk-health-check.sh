#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasını yükle
ENV_FILE=$(realpath "../environments/zookeeper1.env")
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
else
    echo "❌ .env file not found at $ENV_FILE"
    exit 1
fi

# Gerekli değişkenlerin kontrolü
REQUIRED_VARS=(
    "ZK_GROUP_URL" "ZK_ENDPOINT_URL"
    "ZK_HEALTH_EXPECTED_ERROR"
    "HEALTH_CHECK_BROKERS" "CONTAINER_NAME"
    "ADMIN_SERVER_CONTAINER_PORT" "ZK_CHECK_LOCALHOST"
)

for VAR in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!VAR}" ]; then
        echo "❌ Required variable '$VAR' is not set or empty in $ENV_FILE."
        exit 1
    fi
done

# Zookeeper sağlık kontrolü
ZK_URL="$ZK_CHECK_LOCALHOST:$ADMIN_SERVER_CONTAINER_PORT$ZK_GROUP_URL$ZK_ENDPOINT_URL"
ZK_RESPONSE=$(curl -s "$ZK_URL")
if [ -z "$ZK_RESPONSE" ]; then
    echo "❌ Zookeeper health check failed: No response from $ZK_URL"
    exit 1
fi
if [[ $ZK_RESPONSE != *"$ZK_HEALTH_EXPECTED_ERROR"* ]]; then
    echo "❌ Zookeeper is not healthy!"
    exit 1
fi

# Kafka bağlantı kontrolü fonksiyonu
check_kafka_connection() {
    local zookeeper_container_name=$1
    local health_check_brokers=$2

    # kafka_brokers değişkenini bir diziye dönüştürün
    local kafka_brokers
    IFS=' ' read -r -a kafka_brokers <<< "$health_check_brokers"

    for broker in "${kafka_brokers[@]}"; do
        kafka_host=$(echo "$broker" | cut -d: -f1)  # Kafka broker hostname
        kafka_port=$(echo "$broker" | cut -d: -f2)  # Kafka broker port

        echo "ℹ️ Verifying connection to Kafka container '$kafka_host' on port '$kafka_port' from Zookeeper container '$zookeeper_container_name'..."
        if ! docker exec -i "$zookeeper_container_name" bash -c "nc -zv $kafka_host $kafka_port"; then
            echo "❌ Failed to connect to Kafka broker at $kafka_host:$kafka_port."
        else
            echo "✅ Successfully connected to Kafka broker at $kafka_host:$kafka_port."
        fi
    done
}

# Kafka bağlantı kontrolü
check_kafka_connection "$CONTAINER_NAME" "$HEALTH_CHECK_BROKERS"


echo "✅ Zookeeper is healthy."
exit 0
