BINARY_NAME := memory-log-monitor
MOCK_CLIENT_BINARY := mock-client
FRONTEND_DIR := frontend
BACKEND_DIR := ./cmd/goMonitor
WEB_DIST_DIR := web/dist

.PHONY: all build backend frontend clean run

all: build

build: frontend backend mock-client

backend: $(WEB_DIST_DIR)
	@echo ">> Building Go backend"
	GO111MODULE=on go build -o bin/$(BINARY_NAME) $(BACKEND_DIR)

mock-client:
	@echo ">> Building mock client"
	GO111MODULE=on go build -o bin/$(MOCK_CLIENT_BINARY) ./cmd/mock-client

frontend:
	@echo ">> Building Vue3 frontend"
	# cd $(FRONTEND_DIR) && npm install && npm run build
	cd $(FRONTEND_DIR) && npm run format && npm run format:css && npm run build
	@echo ">> Copying frontend dist to embed directory"
	rm -rf $(WEB_DIST_DIR)
	mkdir -p web
	cp -r $(FRONTEND_DIR)/dist $(WEB_DIST_DIR)

$(WEB_DIST_DIR):
	@echo ">> No frontend build found; building..."
	$(MAKE) frontend

clean:
	rm -rf bin
	rm -rf $(WEB_DIST_DIR)

run: build
	./bin/$(BINARY_NAME) -config cmd/goMonitor/config.yml

