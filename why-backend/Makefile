.PHONY: build run test docker-build docker-run

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

docker-build:
	docker build -t why-backend:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env why-backend:latest
