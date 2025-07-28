TOPIC=order-service

up:
	docker-compose up -d --build

down:
	docker-compose down -v

.PHONY: producer

producer:
	docker-compose exec kafka kafka-console-producer.sh --bootstrap-server kafka:9092 --topic ${TOPIC}