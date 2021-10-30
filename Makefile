server	:
	go run ./cmd/server/main.go

proxy	:
ifneq ($(and $(N),$(S)),)
	go run ./cmd/proxy/main.go -N=$(N) -S=$(S)
else
ifneq ($(N),)
	go run ./cmd/proxy/main.go -N=$(N)
else
ifneq ($(S),)
	go run ./cmd/proxy/main.go -S=$(S)
else
	go run ./cmd/proxy/main.go
endif
endif
endif

client	:
	go run ./cmd/client/main.go

# Lint and tests
lint	:
	golangci-lint run ./cmd/... ./internal/...

test	:
	CGO_ENABLED=1 go test -race -cover -count=1 -coverprofile=.coverprofile ./internal/...

dockerfiles	:	
	docker build -f ./docker/Dockerfile.proxy -t order-be-proxy .
	docker build -f ./docker/Dockerfile.server -t order-be-server .

dockerserver	:
	docker run --network host -d -p 8081:8081 -it order-be-server

dockerproxy	:
	docker run --network host -p 8080:8080 order-be-proxy
