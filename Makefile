.PHONY: build run demo test vet fmt tidy clean

build:
	go build -o bin/agent ./cmd/agent

run:
	go run ./cmd/agent

demo:
	go run ./cmd/agent demo

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

tidy:
	go mod tidy

clean:
	rm -rf bin build