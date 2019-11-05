update:
	go get -u -m && go mod tidy

test: install-tools
	go test -race ./...

lint: install-tools
	fgt goimports -w .
	fgt golint ./...
	fgt go vet ./...
	fgt go fmt ./...
	fgt errcheck -ignore Close  ./...

install-tools:
	go mod download
	go get -u golang.org/x/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/GeertJohan/fgt
	go get -u github.com/kisielk/errcheck
	go mod tidy