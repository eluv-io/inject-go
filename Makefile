all: test

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

build: deps
	go build ./...

lint: testdeps
	go get -v golang.org/x/lint/golint
	golint ./.

vet: testdeps
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

test: testdeps pretest
	go test ./...

clean:
	go clean ./...

.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	clean
