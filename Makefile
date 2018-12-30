install:
		@go install

build:
		@go build 

unit-test: 
		@go test -v ./...
