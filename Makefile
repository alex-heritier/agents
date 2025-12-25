.PHONY: build clean run test format

build:
	go build -o agents ./src

clean:
	rm -f agents

run: build
	./agents

test:
	go test -v ./src/... ./test/e2e/...

format:
	go fmt ./src/... ./test/...
