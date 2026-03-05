SHELL := bash

.DEFAULT_GOAL := run

ENV_FILE ?= .env
REQUIRED_ENV_VARS := TELEGRAM_BOT_TOKEN ALLOWED_TELEGRAM_USER_IDS ALLOWED_TELEGRAM_CHAT_IDS

.PHONY: run build-linux-amd64 test tidy check-env

define LOAD_ENV
if [[ -f "$(ENV_FILE)" ]]; then \
	while IFS= read -r line || [[ -n "$$line" ]]; do \
		[[ "$$line" =~ ^[[:space:]]*# ]] && continue; \
		[[ -z "$$line" ]] && continue; \
		key="$${line%%=*}"; \
		value="$${line#*=}"; \
		key="$${key%%[[:space:]]*}"; \
		[[ -z "$$key" ]] && continue; \
		if [[ -z "$${!key+x}" ]]; then \
			export "$$key=$$value"; \
		fi; \
	done < "$(ENV_FILE)"; \
fi
endef

check-env:
	@$(LOAD_ENV); \
	missing=0; \
	for var in $(REQUIRED_ENV_VARS); do \
		if [[ -z "$${!var:-}" ]]; then \
			echo "missing required env var: $$var" >&2; \
			missing=1; \
		fi; \
	done; \
	if [[ $$missing -ne 0 ]]; then \
		echo "Set required vars in shell or $(ENV_FILE)." >&2; \
		exit 1; \
	fi

run: check-env
	@$(LOAD_ENV); \
	go run ./cmd/main.go

build-linux-amd64:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/bot-linux-amd64 ./cmd/main.go

test:
	go test ./...

tidy:
	go mod tidy
