#!/bin/bash
set -e
set -o errexit  # Hata olduğunda çık
set -o pipefail # Pipe hatalarını yakala
set -o nounset  # Tanımsız değişken kullanımını engelle

# Çoklu dosya doğrulama scripti
# Usage: ./verify-multiple-signatures.sh "public_key1.asc,file1.txt,signed_file1.asc;public_key2.asc,file2.txt,signed_file2.asc;..."

# Input format: "public_key_path1,input_file1,signature_file1;public_key_path2,input_file2,signature_file2;..."
INPUT_ARRAY=$1

# INPUT_ARRAY'nın boş olup olmadığını kontrol et
if [ -z "$INPUT_ARRAY" ]; then
  echo "Error: No input provided. Please provide the required input in the format."
  exit 1
fi

# Geçici bir keyring oluşturma
TEMP_KEYRING=$(mktemp)

# Array'i ayrıştır ve her bir dosya için doğrulama yap
IFS=';' read -ra ITEMS <<< "$INPUT_ARRAY"
for item in "${ITEMS[@]}"; do
  IFS=',' read -r PUBLIC_KEY_PATH INPUT_FILE SIGNATURE_FILE <<< "$item"

  echo "Importing the public key from $PUBLIC_KEY_PATH..."
  gpg --no-default-keyring --keyring "$TEMP_KEYRING" --batch --yes --import "$PUBLIC_KEY_PATH"

  echo "Verifying the signature of $INPUT_FILE..."
  gpg --no-default-keyring --keyring "$TEMP_KEYRING" --batch --verify "$SIGNATURE_FILE" "$INPUT_FILE"

  if [ $? -eq 0 ]; then
    echo "The signature of $INPUT_FILE is valid and verified successfully."
  else
    echo "The signature verification failed for $INPUT_FILE."
    rm -f "$TEMP_KEYRING"
    exit 1
  fi

  # Keyring'i temizle
  gpg --no-default-keyring --keyring "$TEMP_KEYRING" --batch --yes --delete-keys
done

# Geçici keyring temizleme
rm -f "$TEMP_KEYRING"
echo "All files verified successfully."
