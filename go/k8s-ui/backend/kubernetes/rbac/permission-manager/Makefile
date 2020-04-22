
dev:
	kubectl apply -f ./deploy/namespace.yml
	kubectl apply -f ./deploy/crd
	kubectl apply -f ./deploy/deployment.yml

cleanup:
	kubectl delete -f ./deploy/crd
	kubectl delete -f ./deploy/deployment.yml
	kubectl delete -f ./deploy/namespace.yml

proxy:
	kubectl port-forward svc/permission-manager-service 4000 --namespace permission-manager
	open http://localhost:4000
