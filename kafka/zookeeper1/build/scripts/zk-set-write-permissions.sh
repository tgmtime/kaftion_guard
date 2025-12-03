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

# Gerekli değişkenlerin kontrolü
REQUIRED_VARS=("ZK_DATA_EXTERNAL_PATH" "ZK_LOG_EXTERNAL_PATH" "ZK_CONFIG_EXTERNAL_PATH")
for VAR in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!VAR}" ]; then
        echo "Error: $VAR is not set in .env file."
        exit 1
    fi
done

# Yazma izinlerini ayarla
echo "Setting write permissions for directories..."

# ZK_DATA_EXTERNAL_PATH için izinler
if chmod -R 777 "$ZK_DATA_EXTERNAL_PATH"; then
    echo "Write permissions set for $ZK_DATA_EXTERNAL_PATH."
else
    echo "Error: Failed to set write permissions for $ZK_DATA_EXTERNAL_PATH."
    exit 1
fi

# ZK_LOG_EXTERNAL_PATH için izinler
if chmod -R 777 "$ZK_LOG_EXTERNAL_PATH"; then
    echo "Write permissions set for $ZK_LOG_EXTERNAL_PATH."
else
    echo "Error: Failed to set write permissions for $ZK_LOG_EXTERNAL_PATH."
    exit 1
fi

# ZK_CONFIG_EXTERNAL_PATH için izinler
if chmod 777 "$ZK_CONFIG_EXTERNAL_PATH"; then
    echo "Write permissions set for $ZK_CONFIG_EXTERNAL_PATH."
else
    echo "Error: Failed to set write permissions for $ZK_CONFIG_EXTERNAL_PATH."
    exit 1
fi

echo "All write permissions set successfully."
