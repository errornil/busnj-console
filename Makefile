# Builds server
build:
	@cd ./cmd/server; \
	go build .

# Build for running binary insite scratch container (runned by DroneCI)
build-drone:
	@cd ./cmd/server; \
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Run binary
run:
	@cd ./cmd/server; \
	go run .

# Runs the go vet command, will be a dependency for any test.
vet:
	@go vet ./cmd/...

# Sets up and runs the test suite within drone.
test: # vet
	@go test ./...; \
	cd ./cmd/server; \
	go test ./...
