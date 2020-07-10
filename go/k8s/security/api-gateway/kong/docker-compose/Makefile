


network:
	docker network create kong-network

dev:
	docker-compose -f ./docker-compose.yml -p kong down && docker-compose -f ./docker-compose.yml -p kong up

scale:
	docker-compose scale kong=3
