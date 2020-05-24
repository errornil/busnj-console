build:
	@cd ./cmd/server; \
	go build .

build-static:
	@cd ./cmd/server; \
	CGO_ENABLED=0 GOOS=linux go build -mod=readonly -a -installsuffix cgo -o server .

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
		--network ${NETWORK} \
		--env ALLOW_LOCALHOST=${ALLOW_LOCALHOST} \
		--env DB_HOST=${DB_HOST} \
		--env DB_NAME=${DB_NAME} \
		--env DB_USERNAME=${DB_USERNAME} \
		--env DB_PASSWORD=${DB_PASSWORD} \
		-p 6001:6001 \
		busnj-console:latest;
