server	:
	go run ./cmd/server/main.go

proxy	:
	go run ./cmd/proxy .

client	:
	go run ./cmd/client/main.go

# Lint and tests
lint	:
	golangci-lint run ./cmd/... ./internal/...

test        :
	CGO_ENABLED=1 go test -race -cover -count=1 -coverprofile=.coverprofile ./internal/...