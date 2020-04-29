CI?=0

update:
	go get -u -m && go mod tidy

test:
	go test -race ./...

lint: install-tools
ifeq ($(CI),1)
	fgt goimports -d .
else
	fgt goimports -w .
endif
	fgt golint ./...
	fgt go vet ./...
	fgt go fmt ./...
	fgt errcheck -ignore Close  ./...

install-tools:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -u github.com/GeertJohan/fgt
	GO111MODULE=off go get -u github.com/kisielk/errcheck

go-mod:
	go mod tidy
	go mod verify

verify: go-mod test lint