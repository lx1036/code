PROJECT=k8s-api-swagger

dev:
	wget https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json
	docker-compose -f ./docker-compose.yml -p $(PROJECT) down && docker-compose -f ./docker-compose.yml -p $(PROJECT) up -d --remove-orphans
