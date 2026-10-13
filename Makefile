.PHONY:  run dev migrate-up migrate-down up down tidy 


run:
	go run ./cmd/server 

dev: 
	go run ./cmd/server  

migrate-up:
	migrate -path ./migrations -database "postgresql://user:password@localhost:5432/roommate?sslmode=disable" up

migrate-down:
	migrate -path ./migrations -database "postgresql://user:password@localhost:5432/roommate?sslmode=disable" down

up:
	docker-compose -f docker/compose.yml up -d 

down:
	docker-compose -f docker/compose.yml down

tidy:
	go mod tidy