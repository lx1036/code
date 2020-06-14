
PROJECT=harbor


harbor:
	docker-compose -f ./docker-compose.yml -p $(PROJECT) down && docker-compose -f ./docker-compose.yml -p $(PROJECT) up --remove-orphans


