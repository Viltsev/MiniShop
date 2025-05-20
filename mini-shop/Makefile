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

migrate-balance:
	migrate create -ext sql -dir migrations add_balance_to_users

migrate-order:
	migrate create -ext sql -dir ./order-service/migrations orders

migrate-outbox:
	migrate create -ext sql -dir ./order-service/migrations outbox

test:
	go test ./... 