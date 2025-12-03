
docker run -d --name kafkabroker1 \
  --hostname kafkabroker1 \
  --network kafka-network \
  -e KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper1:2181 \
  bitnami/kafka:latest


  docker run -d --name kafkabroker1 \
  --hostname kafkabroker1 \
  --network kafka-network \
  -e KAFKA_CFG_ZOOKEEPER_CONNECT='zookeeper1:2181' \
  -e KAFKA_CFG_LISTENERS='PLAINTEXT://0.0.0.0:9092' \
  -e KAFKA_CFG_ADVERTISED_LISTENERS='PLAINTEXT://kafkabroker1:9092' \
  -e KAFKA_CFG_LISTENER_PORT=9092 \
  -p 9092:9092 \
  bitnami/kafka:latest

 **  Kafka broker bağlanabilirlik kontrolü 
  docker run --rm --network kafka-network bitnami/kafka kafka-topics.sh \
    --bootstrap-server kafka-server:9092 --list

** topic oluşturma 

topic oluştururken aynı topic tekrardan oluşturma hatalarını önlemek için "--if-not-exists" kullanılır. Bu şekilde çalıştırınca var olan topic mevcut ise 
topic oluşturmaz hatada vermez

docker run --rm --network kafka-network bitnami/kafka kafka-topics.sh \
    --bootstrap-server kafka-server:9092 --create --topic test-topic \
    --partitions 1 --replication-factor 1 --if-not-exists

** oluşturulan topic başarılı şekilde çalıştırma

docker run --rm --network kafka-network bitnami/kafka kafka-topics.sh \
    --bootstrap-server kafka-server:9092 --list


** topic mesaj gönderme 

echo "Test Message" | docker run --rm --network kafka-network bitnami/kafka kafka-console-producer.sh \
    --broker-list kafka-broker1:9092 --topic test-topic

echo "key1:Test Message" | docker run --rm --network kafka-network bitnami/kafka kafka-console-producer.sh \
    --broker-list kafka-broker1:9092 --topic healthcheck-topic \
    --property "parse.key=true" \
    --property "key.separator=:"


** topic mesaj dinleme 

docker run --rm --network kafka-network bitnami/kafka kafka-console-consumer.sh \
    --bootstrap-server kafka-broker1:9092 --topic test-topic --from-beginning --timeout-ms 5000


docker run --rm --network kafka-network bitnami/kafka kafka-console-consumer.sh \
    --bootstrap-server kafka-server:9092 --topic test-topic \
    --group my-consumer-group \
    --from-beginning \
    --property "print.key=true" \
    --property "key.separator=:"

docker run --rm --network kafka-network bitnami/kafka kafka-console-consumer.sh \
    --bootstrap-server 172.20.0.3:9092 --topic test-topic \
    --group my-consumer-group \
    --from-beginning \
    --property "print.key=true" \
    --property "key.separator=:"

docker run --rm --network kafka-network -d bitnami/kafka kafka-console-consumer.sh \
    --bootstrap-server kafka-broker1:9092 --topic test-topic \
    --group my-consumer-group --from-beginning \
    --property "print.key=true" --property "key.separator=:"

docker run --rm --network kafka-network -d bitnami/kafka kafka-console-consumer.sh \
    --bootstrap-server kafka-broker1:9092 --topic test-topic \
    --group my-consumer-group --from-beginning \
    --property "session.timeout.ms=60000" \
    --property "print.key=true" --property "key.separator=:"


** consumer group bilgisi alma
docker run --rm --network kafka-network bitnami/kafka kafka-consumer-groups.sh \
    --bootstrap-server kafka-broker1:9092 --describe --group my-consumer-group

** consumer group resetleme mesajları baştan almak için

docker run --rm --network kafka-network bitnami/kafka kafka-consumer-groups.sh \
    --bootstrap-server kafka-broker1:9092 --group my-consumer-group --topic test-topic --reset-offsets --to-earliest --execute


** Topic silme işlemi 

docker run --rm --network kafka-network bitnami/kafka kafka-topics.sh \
    --bootstrap-server kafka-broker1:9092 --delete --topic test-topic






docker run --rm --network kafka-network bitnami/kafka kafka-topics.sh --bootstrap-server kafka-broker1:9092 --list



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
    --health-cmd "$HEALTH_CMD" \
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
