SHELL := /bin/bash
.ONESHELL:  # Tüm komutları tek bir shell içinde çalıştırmak için
.SHELLFLAGS = -e  # Hata durumunda süreci durdur

# .env dosyasının bulunduğu yol (bir üst klasörde environments klasörü)
ENV_FILE := ../environments/makefile.env

# .env dosyasını içeri aktar
include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE)) # .env içindeki tüm değişkenleri export eder

# Renkli çıktı tanımları
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
RESET  := \033[0m

# Başlangıç ve bitiş mesajları için yardımcı fonksiyon (özel mesaj desteği eklendi)
define print_message
	@echo -e "$(YELLOW)🚀 Running: $(1)$(RESET)"
	@echo -e "$(GREEN)➡️  $(2)$(RESET)"
	@$(1); \
	if [ $$? -ne 0 ]; then \
		echo -e "$(RED)❌ Error occurred in: $(1)$(RESET)"; \
		exit 1; \
	fi
	@echo -e "$(GREEN)✅ Tamamlandı: $(1)$(RESET)"
endef



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


.PHONY: zookeeper-1 broker-1 check-all system-setup build help

# İlk Makefile'i çalıştırma hedefi (özel mesaj ile)
zookeeper-1:
	$(call print_message, $(MAKE) -C $(ZOOKEEPER1) $(MAKEFILE_COMMAND) -f $(MAKEFILE_NAME), $(ZOOKEEPER1_INFO_MSG))

# İkinci Makefile'i çalıştırma hedefi (farklı özel mesaj ile)
broker-1:
	$(call print_message, $(MAKE) -C $(BROKER1) $(MAKEFILE_COMMAND) -f $(MAKEFILE_NAME), $(BROKER1_INFO_MSG))

# Tüm Makefile'leri sırasıyla çalıştır
check-all: zookeeper-1 broker-1

# Setup işlemleri: Eğer SETUP_SPECIFIER true ise, gerekli işlemleri çalıştır
system-setup: set-permissions verify-all
	@if [ "$(SETUP_SPECIFIER)" = "true" ]; then \
		$(MAKE) zookeeper-1 broker-1 || exit 1; \
	else \
		echo -e "$(RED)Setup is not specified. Skipping setup.$(RESET)"; \
	fi

# Build işlemi: Setup işlemi çağrılır
build: system-setup

# Yardımcı hedef
help:
	@echo -e "$(GREEN)Kullanım:$(RESET)"
	@echo -e "  make check-network-1  - zookeeper-1 run"
	@echo -e "  make check-network-2  - broker-1 run"
	@echo -e "  make check-all        - zookeeper-1 and broker-1 run"
	@echo -e "  make system-setup     - runs system setup if specified"
	@echo -e "  make build            - builds after setup"
	@echo -e "  make help             - help command"
