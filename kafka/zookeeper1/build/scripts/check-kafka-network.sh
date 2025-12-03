#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasını yükle (Bir üst klasördeki environments klasöründen)
ENV_FILE=$(realpath "../environments/zookeeper1.env")

# .env dosyasını yükleme
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
else
    echo "Error: .env file not found at $ENV_FILE"
    exit 1
fi

# CONTAINER_NETWORK_NAME değişkenini kontrol et
if [ -z "$CONTAINER_NETWORK_NAME" ]; then
    echo "Error: CONTAINER_NETWORK_NAME is not set in .env file."
    exit 1
fi

if ! docker network inspect "$CONTAINER_NETWORK_NAME" > /dev/null 2>&1; then
    echo "Error: Network $CONTAINER_NETWORK_NAME does not exist."
    exit 1
else
    echo "Network $CONTAINER_NETWORK_NAME exists. Continuing..."
fi
