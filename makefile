.PHONY: build install clean test run

BINARY=power-recovery
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/power-recovery

install: build
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
	sudo mkdir -p /var/lib/power-failure-recovery
	sudo cp configs/default.json /etc/power-failure-recovery/config.json
	sudo cp systemd/power-failure-recovery.service /etc/systemd/system/
	sudo systemctl daemon-reload

test:
	go test -v ./...

run: build
	sudo $(BUILD_DIR)/$(BINARY) --debug

clean:
	rm -rf $(BUILD_DIR)
	sudo rm -f /usr/local/bin/$(BINARY)

restore: build
	$(BUILD_DIR)/$(BINARY) --restore