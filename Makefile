build-bin:
	go build -o nodebot .

build-container:
	podman build -t local/nodebot:test .

lint-go:
	podman run -it --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.5.0 golangci-lint run

lint-dockerfile:
	hadolint Dockerfile

lint-commit:
	commitlint --last --verbose

lint: lint-dockerfile lint-go lint-commit
