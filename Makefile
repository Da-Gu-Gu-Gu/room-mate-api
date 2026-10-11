.PHONY:  run dev up down tidy


run:
	go run ./cmd/server 

dev: 
	go run ./cmd/server  

up:
	docker-compose -f docker/compose.yml up -d 

down:
	docker-compose -f docker/compose.yml down

tidy:
	go mod tidy