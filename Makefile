.PHONY: dep
dep:
		@go get -t ./...

.PHONY: install
install:
		@go install

.PHONY: build
build:
		@go build 

.PHONY: test
test: 
		@go test -v ./...

.PHONY: release
release:
		@mkdir -p release
		@GOOS=linux  GOARCH=amd64 go build -o release/mono-meta-linux-amd64
		@GOOS=darwin GOARCH=amd64 go build -o release/mono-meta-darwin-amd64
