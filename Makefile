CI?=0

update:
	go get -u -d && go mod tidy

test:
	go test -race ./...

lint: install-tools
	golangci-lint run --timeout=3m

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.0

go-mod:
	go mod tidy
	go mod verify

verify: go-mod test lint