build:
	@go build .

build-static:
	@CGO_ENABLED=0 GOOS=linux go build -mod=readonly -a -installsuffix cgo -o server .

run:
	@go run .

vet:
	@go vet ./...

test: # vet
	@go test ./...

build-docker:
	@docker build --tag busnj-console:latest .;

run-docker:
	@docker run --name busnj-console \
		--rm \
		--network busnj-network \
		-p 6001:6001 \
		busnj-console:latest;

run-docker-allow-localhost:
	@docker run --name busnj-console \
		--rm \
		--network busnj-network \
		--env ALLOW_LOCALHOST=true \
		-p 6001:6001 \
		busnj-console:latest;
