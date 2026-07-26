BINARY := coursesmith

.PHONY: build test test-e2e lint run clean sandbox renderer

build:
	go build -o bin/$(BINARY) ./cmd/coursesmith

test:
	go test ./...

# Full pipeline on lesson 01: needs docker + the sandbox image, a running
# Kokoro server, renderer/node_modules, and tools/align/.venv.
test-e2e:
	COURSESMITH_E2E=1 go test ./internal/pipeline/ -run TestEndToEndLesson01 -v -timeout 45m

lint:
	golangci-lint run ./...

run: build
	./bin/$(BINARY) $(ARGS)

sandbox:
	docker build -t coursesmith-sandbox sandbox/

renderer:
	cd renderer && npm install

clean:
	rm -rf bin
