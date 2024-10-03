# Makefile for my project

# Target to run the tests
test:
	go test -cover ./...

# Run cover Profile (Windows)
coverprofile_w:
	go test ./... -coverprofile cover.out
	go tool cover -func cover.out

# Run cover Profile (Linux)
coverprofile_l:
	go test ./... -coverprofile cover.out
	go tool cover -func cover.out | grep total | awk '{print $$3}'
	go tool cover -func cover.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'

audit:
	go vet ./...
	staticcheck ./...
	golint ./...
	golangci-lint run ./...
	gocyclo -over 15 .
	govulncheck ./...
	gosec ./...


cleanandupdate:
	go clean --modcache
	go get -u ./...
	go mod tidy