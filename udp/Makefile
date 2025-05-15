build: compile test lint

test:
	go test -count=1 -cover ./...

lint:
	golangci-lint run ./...

compile:
	go build udp/...

run:
	go run test_udp/main.go
