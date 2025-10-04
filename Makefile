# Builds

build-bin:
	go build -o nodebot .

build-container:
	podman build -t local/nodebot:test .

# Lints

lint-go:
	podman run -it --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.5.0 golangci-lint run

lint-dockerfile:
	hadolint Dockerfile

lint-commit:
	commitlint --last --verbose

lint: lint-dockerfile lint-go lint-commit

# Local env

start-env: | setup migrate

setup:
	podman compose up -d

migrate:
	podman run -v $(PWD)/internal/db/migrations:/migrations --network host migrate/migrate:v4.19.0 -path=/migrations/ -database postgres://postgres:password@localhost:5432/nodebot up
