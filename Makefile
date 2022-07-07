
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
	@/usr/local/bin/go test github.com/JekaTatsiy/kafka_service/service --args -s 0.0.0.0:9092 || true
	@echo --- TEST END ---
#	@docker-compose -f test/Docker-compose.yml down