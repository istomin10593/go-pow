up:
	sudo docker compose up --build

down:
	sudo docker compose down

server-run:
	cd server && make run

client-run:
	cd client && make run

test:
	cd server && make test; \
	cd ..; \
	cd client && make test

lint:
	@(golangci-lint run)

PHONE: lint