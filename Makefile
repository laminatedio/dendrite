ifeq ($(PREFIX),)
	PREFIX := /usr/local
endif

.PHONY: build
build:
	go build -o dendrite .

.PHONY: clean
clean:
	rm -f dendrite

.PHONY: dev
dev:
	go run main.go

PHONY: test
test:
	go test ./...
