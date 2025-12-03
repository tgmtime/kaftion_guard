#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# .env dosyasının bulunduğu yolu belirt
ENV_FILE=$(realpath "../environments/zookeeper1.env")

# .env dosyasını yükleme
if [ -f "$ENV_FILE" ]; then
    source "$ENV_FILE"
else
    echo "Error: .env file not found at $ENV_FILE"
    exit 1
fi

# CONTAINER_IMAGE değişkenini kontrol et
if [ -z "$CONTAINER_IMAGE" ]; then
    echo "Error: CONTAINER_IMAGE is not set in .env file."
    exit 1
fi

# Docker imajını kontrol et ve gerekirse çek
echo "Checking if image $CONTAINER_IMAGE exists..."
if ! docker image inspect "$CONTAINER_IMAGE" > /dev/null 2>&1; then
    echo "Image $CONTAINER_IMAGE does not exist. Attempting to pull..."
    if docker pull "$CONTAINER_IMAGE"; then
        echo "Image $CONTAINER_IMAGE pulled successfully."
    else
        echo "Error: Failed to pull image $CONTAINER_IMAGE."
        exit 1
    fi
else
    echo "Image $CONTAINER_IMAGE already exists. Skipping pull."
fi
