server	:
	go run ./cmd/server/main.go

proxy	:
	go run ./cmd/proxy .

client	:
	go run ./cmd/client/main.go