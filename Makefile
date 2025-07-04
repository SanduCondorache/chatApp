build-server:
	mkdir -p ./bin
	go build -o ./bin/server ./cmd/server

build-client:
	mkdir -p ./bin
	go build -o ./bin/client ./cmd/client

run-server: build-server
	./bin/server

run-client: build-client
	./bin/client
