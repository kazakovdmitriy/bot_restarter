BINARY := bot-restarter
REMOTE_HOST ?= your-server
REMOTE_USER ?= root
REMOTE_DIR ?= /opt/bot-restarter

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: test-cover
test-cover:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: lint
lint:
	go vet ./...

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY) .

.PHONY: deploy
deploy: build
	scp $(BINARY) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	scp config.json $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)/
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "systemctl restart bot-restarter"

.PHONY: install-service
install-service:
	scp bot-restarter.service $(REMOTE_USER)@$(REMOTE_HOST):/etc/systemd/system/
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "systemctl daemon-reload && systemctl enable bot-restarter"

.PHONY: status
status:
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "systemctl status bot-restarter"

.PHONY: logs
logs:
	ssh $(REMOTE_USER)@$(REMOTE_HOST) "journalctl -u bot-restarter -f"
