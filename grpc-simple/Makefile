.PHONY: server client deps clean

server:
	go run server.go

client:
	go run client.go

client-with-name:
	go run client.go $(name)

deps:
	go mod tidy

clean:
	go clean