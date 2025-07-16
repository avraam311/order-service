TOPIC=orderService

up:
	docker compose up -d

down:
	docker compose down -v

.PHONY: producer

producer:
	docker compose exec kafka kafka-console-producer.sh --bootstrap-server kafka:9092 --topic ${TOPIC}