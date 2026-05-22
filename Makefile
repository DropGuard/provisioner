VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -s -w -X main.Version=$(VERSION)

.PHONY: build clean test fmt vet lint run-provisioner run-user

build:
	go build -ldflags "$(LDFLAGS)" -o bin/provisioner.exe ./cmd/provisioner
	go build -ldflags "$(LDFLAGS)" -o bin/user.exe ./cmd/user

clean:
	rm -rf bin || true

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: fmt vet

run-provisioner:
	go build -ldflags "$(LDFLAGS)" -o bin/provisioner.exe ./cmd/provisioner
	./bin/provisioner.exe

run-user:
	go build -ldflags "$(LDFLAGS)" -o bin/user.exe ./cmd/user
	./bin/user.exe
