build:
	go build -o bin/main cmd/*.go

run: build
	./bin/main

debug: build
	DEBUG=true ./bin/main

test:
	go test -v ./...

