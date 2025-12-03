SHELL := /bin/bash

# .env dosyasının bulunduğu yol (bir üst klasörde environments klasörü)
ENV_FILE := ../environments/makefile.env

# .env dosyasını içeri aktar
include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE)) # .env içindeki tüm değişkenleri export eder

# .sh dosyalarına yürütme izni ver
set-permissions:
	@echo "Setting execute permissions for .sh scripts..."
	@chmod +x $(SCRIPT_DIR)/*.sh || { echo "Error: Failed to set execute permissions for scripts in $(SCRIPT_DIR)."; exit 1; }
	@echo "Execute permissions set successfully for all scripts."

# Çoklu dosya doğrulama işlemi için bir hedef
verify-all:	
	@echo "Defining VERIFY_INPUT..."	
	@echo "Starting signature verification for all files..."
	@$(VERIFY_SIGNATURE_FILE) "$(VERIFY_INPUT)" || { \
		echo "Error: Signature verification failed for the following files:"; \
		echo "$(VERIFY_INPUT)"; \
		exit 1; \
	}
	@echo "All files verified successfully."

# Her bir dosya özelinde hedefler
check-kafka-network:
	$(CHECK_KAFKA_NETWORK_FILE)

pull-br:
	$(PULL_BR_FILE)

stop-and-remove-br-container:
	$(STOP_AND_REMOVE_BR_CONTAINER_FILE)

br-set-write-permissions:
	$(BR_SET_WRITE_PERMISSIONS_FILE)

run-br-container:
	$(RUN_BR_CONTAINER_FILE)

test-br-container:
	$(TEST_BR_CONTAINER_FILE)

# Setup işlemleri: Eğer SETUP_SPECIFIER true ise, gerekli işlemleri çalıştır
system-setup: set-permissions verify-all
	@if [ "$(SETUP_SPECIFIER)" = "true" ]; then \
		$(MAKE) check-kafka-network pull-br stop-and-remove-br-container br-set-write-permissions run-br-container; \
		$(MAKE) test-br-container; \
	else \
		echo "Setup is not specified. Skipping setup."; \
	fi

# Build işlemi: Setup işlemi çağrılır
build: system-setup