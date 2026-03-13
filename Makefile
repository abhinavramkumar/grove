.PHONY: build install clean test cross

build:
	go build -o bin/grove ./cmd/grove

install:
	go install ./cmd/grove

cross:
	GOOS=darwin GOARCH=arm64 go build -o bin/grove-darwin-arm64 ./cmd/grove
	GOOS=darwin GOARCH=amd64 go build -o bin/grove-darwin-amd64 ./cmd/grove
	GOOS=linux GOARCH=amd64 go build -o bin/grove-linux-amd64 ./cmd/grove

clean:
	rm -rf bin/

test:
	go test ./...
