.PHONY: all build test demo clean generate

all: build

generate:
	sqlc generate
	oapi-codegen --config=internal/api/cfg.yaml project/openapi.yaml

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/fake-idp ./cmd/fake-idp

test:
	go test -v ./...

demo: build
	./scripts/demo.sh

clean:
	rm -rf bin demo_identity.db
