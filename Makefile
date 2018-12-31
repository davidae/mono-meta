.PHONY: dep
dep:
		@go get ./...

.PHONY: install
install:
		@go install

.PHONY: build
build:
		@go build 

.PHONY: test
test: 
		@go test -v ./...
