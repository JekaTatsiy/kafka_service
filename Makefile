
build:
	GOOS=linux go build -o main main.go 

up: build
	docker-compose build
	docker-compose up -d

down:
	docker-compose down

runtest:
	@docker-compose -f test/Docker-compose.yml build
	@docker-compose -f test/Docker-compose.yml up -d 
	@echo --- TEST START ---
	@go test github.com/JekaTatsiy/kafka_service/service -v || true
	@echo --- TEST END ---
	@docker-compose -f test/Docker-compose.yml down