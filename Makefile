VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -s -w -X main.Version=$(VERSION)
OUT_DIR ?= bin
SUFFIX ?= 

.PHONY: build clean test fmt vet lint install-user install-apps disable-password-policy

build:
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/provisioner$(SUFFIX).exe ./cmd/provisioner
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/user$(SUFFIX).exe ./cmd/user

disable-password-policy:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/disable-password-policy.ps1


clean:
	rm -rf bin || true

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: fmt vet

install-user: build
	./bin/user.exe

install-apps: build
	./bin/provisioner.exe
