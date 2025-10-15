docker-up:
	docker compose up --build --force-recreate -d

docker-down:
	docker compose down --rmi all