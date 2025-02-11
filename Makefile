start:
	@docker compose up 
build:
	@docker compose up --build
down:
	@docmer compose down
clear-volumes:
	@docker-compose down --volumes --remove-orphans