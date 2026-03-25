BINARY=speek
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build build-all test install release run-example clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

build-all:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/speek_linux_amd64 .
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/speek_linux_arm64 .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/speek_darwin_amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/speek_darwin_arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/speek_windows_amd64.exe .

test:
	go test ./... -v -cover

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed speek to /usr/local/bin/speek"

release:
	goreleaser release --clean

run-example:
	go run . run examples/hello.spk

clean:
	rm -f $(BINARY)
	rm -rf dist/
