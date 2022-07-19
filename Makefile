build:
	docker build -t ticket-service:1.0 .
run:
	./scripts/ticket-service-up.sh
start:
	docker-compose start
stop:
	docker-compose stop
purge:
	docker-compose down -v