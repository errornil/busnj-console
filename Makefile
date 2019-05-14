build:
	@cd ./cmd/server; \
	go build .

build-drone:
	@cd ./cmd/server; \
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

run:
	@cd ./cmd/server; \
	go run .

vet:
	@go vet ./cmd/...

test: # vet
	@go test ./...

build-docker:
	@docker build --tag busnj-console:latest ./cmd/server;

run-docker:
	@docker run --name busnj-console \
		--rm \
		--network busnj-network \
		-p 6001:6001 \
		busnj-console:latest;
