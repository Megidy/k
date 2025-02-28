start:
	@docker compose up 
build:
	@docker compose up --build
down:
	@docker compose down
clear-volumes:
	@docker-compose down --volumes --remove-orphans
gen:
	@templ generate