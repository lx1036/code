


network:
	docker network create kong-net

dev:
	docker-compose -f ./docker-compose.yml -p kong down && docker-compose -f ./docker-compose.yml -p kong up -d

scale:
	docker-compose scale kong=3
