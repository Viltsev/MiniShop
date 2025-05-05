up:
	docker-compose up -d --build

stop:
	docker-compose stop

down:
	docker-compose down

restart:
	docker-compose down && docker-compose up -d --build

logs:
	docker-compose logs -f

migrate-up:
	./scripts/migrate.sh

migrate-down:
	./scripts/migrate.sh down

test:
	go test ./... 