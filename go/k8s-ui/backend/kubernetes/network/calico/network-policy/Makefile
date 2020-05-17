


install:
	kubectl apply --kustomize .

cleanup:
	kubectl delete --kustomize .

stars-deny:
	kubectl apply -n stars -f ./default-deny.yaml

client-deny:
	kubectl apply -n client -f ./default-deny.yaml

stars-allow:
	kubectl apply -f ./allow-ui.yaml

client-allow:
	kubectl apply -f ./allow-ui-client.yaml

backend-allow:
	kubectl apply -f ./backend-policy.yaml

frontend-allow:
	kubectl apply -f ./frontend-policy.yaml

cleanup-ns:
	kubectl delete ns client stars management-ui
