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

pull-zk:
	$(PULL_ZK_FILE)

stop-and-remove-zk-container:
	$(STOP_AND_REMOVE_ZK_CONTAINER_FILE)

zk-set-write-permissions:
	$(ZK_SET_WRITE_PERMISSIONS_FILE)

run-zk-container:
	$(RUN_ZK_CONTAINER_FILE)

test-zk-container:
	$(TEST_ZK_CONTAINER_FILE)

# Setup işlemleri: Eğer SETUP_SPECIFIER true ise, gerekli işlemleri çalıştır
system-setup: set-permissions verify-all
	@if [ "$(SETUP_SPECIFIER)" = "true" ]; then \
		$(MAKE) check-kafka-network pull-zk stop-and-remove-zk-container zk-set-write-permissions run-zk-container; \
		$(MAKE) test-zk-container; \
	else \
		echo "Setup is not specified. Skipping setup."; \
	fi

# Build işlemi: Setup işlemi çağrılır
build: system-setup