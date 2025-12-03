#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasını yükle
ENV_FILE=$(realpath "../environments/broker1.env")
if [ -f "$ENV_FILE" ]; then    
    source "$ENV_FILE"    
else
    echo "❌ .env file not found at $ENV_FILE"
    exit 1
fi

# Gerekli değişkenlerin kontrolü
REQUIRED_VARS=(
    "CONTAINER_NAME" "CONTAINER_HOST_NAME"
    "ZK_GROUP_URL" "ZK_ENDPOINT_URL" "HEALTH_CHECK_ZOOKEEPERS"    
    "CONTAINER_HOST_NAME" "ZK_CONTAINER_NAMES"
    "CLIENT_HOST_PORT" "CONTAINER_HOST_NAME"
    "HEALTHCHECK_TOPIC" "RETRY_TOPIC" "DLQ_TOPIC"
    "PRODUCER_HEALTHCHECK_MSG" "CONSUMER_GROUP" "HEALTH_SUCCESS_MSG"
)

for VAR in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!VAR}" ]; then
        echo "❌ Required variable '$VAR' is not set or empty in $ENV_FILE."
        exit 1
    fi
done

KAFKA_BROKER="$CONTAINER_HOST_NAME:$CLIENT_HOST_PORT"
TOPICS=("$HEALTHCHECK_TOPIC" "$RETRY_TOPIC" "$DLQ_TOPIC")

# Kafka zookeeper bağlantı kontrolü fonksiyonu
check_zookeeper_connection_from_kafka() {
    local health_check_zookeepers=$1
    local broker_container_name=$2
    local zk_group_url=$3
    local zk_endpoint_url=$4

    # kafka_zookeepers değişkenini bir diziye dönüştürün
    local kafka_zookeepers
    IFS=' ' read -r -a kafka_zookeepers <<< "$health_check_zookeepers"

    zk_path="$zk_group_url$zk_endpoint_url"

    for zookeeper in "${kafka_zookeepers[@]}"; do
        zookeeper_host=$(echo "$zookeeper" | cut -d: -f1)  # Kafka zookeeper hostname
        zookeeper_port=$(echo "$zookeeper" | cut -d: -f2)  # Kafka zookeeper port

        echo "ℹ️ Verifying connection to Kafka zookeeper at '$zookeeper_host' on port '$zookeeper_port' from broker container '$broker_container_name'..."
        
        # Bağlantı kontrolü
        if ! docker exec -i "$broker_container_name" bash -c "curl -s --fail $zookeeper_host:$zookeeper_port$zk_path"; then
            echo "❌ Failed to connect to Kafka zookeeper at $zookeeper_host:$zookeeper_port."
        else
            echo "✅ Successfully connected to Kafka zookeeper at $zookeeper_host:$zookeeper_port."
        fi
    done
}


# Kafka bağlantı kontrolü fonksiyonu
check_kafka_connection_from_zookeeper() {
    local zk_container_names=$1
    local broker_host=$2
    local broker_port=$3

    # ZK_CONTAINER_NAMES değişkenini bir diziye dönüştürün
    local zk_containers_array
    IFS=',' read -r -a zk_containers_array <<< "$zk_container_names"

    for zk_container in "${zk_containers_array[@]}"; do
        echo "ℹ️ Verifying connection from $zk_container to Kafka broker at $broker_host:$broker_port..."
        if ! docker exec -it "$zk_container" /bin/bash -c "nc -zv $broker_host $broker_port"; then
            echo "❌ Connection failed for $zk_container to $broker_host:$broker_port."
            exit 1
        fi
        echo "✅ Connection successful for $zk_container to $broker_host:$broker_port."
    done
}

# Kafka bağlantı kontrolü fonksiyonu
check_kafka_connection() {
    local broker=$1
    if ! docker run --rm --network "$CONTAINER_NETWORK_NAME" bitnami/kafka kafka-topics.sh \
        --bootstrap-server "$broker" --list >/dev/null 2>&1; then
        echo "❌ Cannot connect to Kafka broker at $broker."
        exit 1
    fi
}

# Kafka topic işlemleri fonksiyonları
delete_topic_if_exists() {
    local broker=$1
    local topic=$2
    if docker run --rm --network "$CONTAINER_NETWORK_NAME" bitnami/kafka kafka-topics.sh \
        --bootstrap-server "$broker" --list | grep -w "$topic" &>/dev/null; then
        echo "ℹ️ Topic '$topic' exists. Deleting..."
        docker run --rm --network "$CONTAINER_NETWORK_NAME" bitnami/kafka kafka-topics.sh \
            --bootstrap-server "$broker" --delete --topic "$topic" --if-exists
        echo "✅ Topic '$topic' deleted."
    fi
}

create_topic() {
    local broker=$1
    local topic=$2
    if docker run --rm --network "$CONTAINER_NETWORK_NAME" bitnami/kafka kafka-topics.sh \
        --bootstrap-server "$broker" --create --topic "$topic" \
        --partitions 1 --replication-factor 1 --if-not-exists; then
        echo "✅ Kafka topic '$topic' successfully created."
    else
        echo "❌ Failed to create topic '$topic'."        
        exit 1
    fi
}

cleanup_topics() {
    local kafka_broker=$1         # İlk argüman Kafka broker adresi
    shift                         # İlk argümanı atlıyoruz
    local topics=("$@")           # Kalan argümanlar topic listesi olarak alınıyor

    for topic in "${topics[@]}"; do
        delete_topic_if_exists "$kafka_broker" "$topic"
    done
}

echo "checHEALTH_CHECK_ZOOKEEPERS  path: $HEALTH_CHECK_ZOOKEEPERS"

# zookeeper bağlantı kontrolü
check_zookeeper_connection_from_kafka "$HEALTH_CHECK_ZOOKEEPERS" "$CONTAINER_NAME" "$ZK_GROUP_URL" "$ZK_ENDPOINT_URL"


check_kafka_connection_from_zookeeper "$ZK_CONTAINER_NAMES" "$CONTAINER_HOST_NAME" "$CLIENT_HOST_PORT"

echo "✅ kafka-zookeeper/kafka-broker connection is healthy."

# Kafka bağlantı ve topic işlemleri
check_kafka_connection "$KAFKA_BROKER"

# Test case başlamadan önce tüm topic'lerin temizlenmesi
cleanup_topics "kafka-broker1:9092" "${TOPICS[@]}"

# Topic oluşturma
create_topic "$KAFKA_BROKER" "$HEALTHCHECK_TOPIC"
create_topic "$KAFKA_BROKER" "$RETRY_TOPIC"
create_topic "$KAFKA_BROKER" "$DLQ_TOPIC"

# Test mesajı oluşturma
TEST_MESSAGE="${PRODUCER_HEALTHCHECK_MSG}-$(uuidgen)"
KEY=$(uuidgen)

# Asenkron tüketici başlatma
consume_test_message() {
    local consumed_message=""
    local max_attempts=5
    local attempt=1
    local timeout_ms=5000
    local key_filter="$KEY"

    echo "ℹ️ Starting to consume test message from topic '$HEALTHCHECK_TOPIC'..."

    while [ $attempt -le $max_attempts ]; do
        echo "⏳ Attempting to consume message (Attempt $attempt/$max_attempts)..."

        # Docker içinde Kafka tüketici başlat
        consumed_message=$(docker run --rm --network "$CONTAINER_NETWORK_NAME" bitnami/kafka kafka-console-consumer.sh \
            --bootstrap-server "$KAFKA_BROKER" \
            --topic "$HEALTHCHECK_TOPIC" \
            --group "$CONSUMER_GROUP" \
            --offset latest \
            --timeout-ms $timeout_ms 2>&1)

        # Gelen mesaj kontrolü
        if [[ $consumed_message == *"TimeoutException"* ]]; then
            echo "❌ Timeout occurred while waiting for message. Retrying..."
        elif [[ $consumed_message == *"$key_filter"* ]]; then
            echo "✅ Test message consumed successfully: $consumed_message"
            return 0
        elif [[ $consumed_message != *"$key_filter"* ]]; then
            echo "⚠️ Another message received: $consumed_message"
            return 1
        fi

        attempt=$((attempt + 1))
        sleep 2
    done

    echo "❌ Test message not consumed after $max_attempts attempts."
    exit 1
}

# Tüketiciyi arka planda başlat
consume_test_message &

# Küçük bir bekleme süresi tüketicinin başlaması için
sleep 3

# Test mesajını Kafka'ya gönder
if echo "$KEY:$TEST_MESSAGE" | docker run --rm --network "$CONTAINER_NETWORK_NAME" bitnami/kafka kafka-console-producer.sh \
    --broker-list "$KAFKA_BROKER" \
    --topic "$HEALTHCHECK_TOPIC" \
    --property "parse.key=true" \
    --property "key.separator=:"; then
    echo "✅ Test message sent to Kafka with key '$KEY'."
else
    echo "❌ Failed to send test message to Kafka."
    exit 1
fi

# Tüketici işlemini bekle
wait


# Sağlık kontrolü sonrası tüm topic'leri sil
cleanup_topics "kafka-broker1:9092" "${TOPICS[@]}"

# Başarılı mesaj
echo "✅ $HEALTH_SUCCESS_MSG"
exit 0
