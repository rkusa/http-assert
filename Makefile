all: deps test

test:
	go vet ./...
	go test -cover -short ./...

deps:
	go get -t -v ./...
	go get -u golang.org/x/tools/cmd/vet
	go get -u golang.org/x/tools/cmd/cover

.PHONY: deps test
