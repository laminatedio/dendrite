.PHONY: build
build:
	go build -o dendrite -ldflags=" -X 'github.com/astaclinic/astafx/info.BuildDate=${BUILD_DATE}' -X 'github.com/astaclinic/astafx/info.ProgramName=${NAME}' " .

.PHONY: clean
clean:
	rm -f dendrite

.PHONY: dev
dev:
	go run main.go

PHONY: test
test:
	go test ./...
