PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

all: verifiers test

build:
	@echo "Running $@"
	@go build -ldflags=\
	"-X 'github.com/rawdaGastan/app_tfgrid_deployer/cmd.commit=$(shell git rev-parse HEAD)'\
	 -X 'github.com/rawdaGastan/app_tfgrid_deployer/cmd.version=$(shell git tag --sort=-version:refname | head -n 1)'"\
	 -o bin/app_tfgrid_deployer main.go

deploy: build
	bin/app_tfgrid_deployer deploy

test: 
	@echo "Running Tests"
	go test -v ./...

coverage: clean 
	mkdir coverage
	go test -v -vet=off ./... -coverprofile=coverage/coverage.out
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@${GOPATH}/bin/gopherbadger -png=false -md="README.md"
	rm coverage.out

clean:
	rm ./coverage -rf
	rm ./bin -rf
	
getverifiers:
	@echo "Installing staticcheck" && go get -u honnef.co/go/tools/cmd/staticcheck && go install honnef.co/go/tools/cmd/staticcheck
	@echo "Installing gocyclo" && go get -u github.com/fzipp/gocyclo/cmd/gocyclo && go install github.com/fzipp/gocyclo/cmd/gocyclo
	@echo "Installing deadcode" && go get -u github.com/remyoudompheng/go-misc/deadcode && go install github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell" && go get -u github.com/client9/misspell/cmd/misspell && go install github.com/client9/misspell/cmd/misspell
	@echo "Installing golangci-lint" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
	@echo "Installing gopherbadger" && go get github.com/jpoles1/gopherbadger && go install github.com/jpoles1/gopherbadger
	go mod tidy

verifiers: fmt lint cyclo deadcode spelling staticcheck

checks: verifiers

fmt:
	@echo "Running $@"
	@gofmt -d .

lint:
	@echo "Running $@"
	@${GOPATH}/bin/golangci-lint run

cyclo:
	@echo "Running $@"
	@${GOPATH}/bin/gocyclo -over 100 .

deadcode:
	@echo "Running $@"
	@${GOPATH}/bin/deadcode -test $(shell go list ./...) || true

spelling:
	@echo "Running $@"
	@${GOPATH}/bin/misspell -i monitord -error `find .`

staticcheck:
	@echo "Running $@"
	@${GOPATH}/bin/staticcheck -- ./...
