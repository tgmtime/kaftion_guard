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

# CONTAINER_NAME değişkenini kontrol et
if [ -z "$CONTAINER_NAME" ]; then
    echo "Error: CONTAINER_NAME is not set in .env file."
    exit 1
fi

# Konteyneri kontrol et ve işlemleri uygula
echo "Checking for existing container $CONTAINER_NAME..."
if docker container ls -aq -f name="$CONTAINER_NAME" | grep -q .; then
    echo "Container $CONTAINER_NAME found."

    # Eğer konteyner çalışıyorsa durdur
    if docker container inspect -f '{{.State.Running}}' "$CONTAINER_NAME" 2>/dev/null | grep -q true; then
        echo "Container $CONTAINER_NAME is running. Stopping..."
        if ! docker stop "$CONTAINER_NAME"; then
            echo "Error: Failed to stop container $CONTAINER_NAME."
            exit 1
        fi
        echo "Container $CONTAINER_NAME stopped successfully."
    fi

    # Konteyneri sil
    echo "Removing container $CONTAINER_NAME..."
    if ! docker rm "$CONTAINER_NAME"; then
        echo "Error: Failed to remove container $CONTAINER_NAME."
        exit 1
    fi
    echo "Container $CONTAINER_NAME removed successfully."
else
    echo "No container $CONTAINER_NAME found. Skipping removal."
fi
