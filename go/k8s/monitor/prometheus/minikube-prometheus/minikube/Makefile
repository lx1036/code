
install:
	kubectl apply -f ./namespace.yml
	kubectl apply -f ./operator/
	kubectl apply -f ./adapter/
	kubectl apply -f ./alertmanager/
	kubectl apply -f ./node-exporter/
	kubectl apply -f ./kube-state-metrics/
	kubectl apply -f ./grafana/dashboardDefinitions.yml
	kubectl apply -f ./grafana/
	kubectl apply -f ./prometheus/
	kubectl apply -f ./service-monitor/
	kubectl apply -f ./example/

validate:
	kubectl apply -f ./example/

cleanup:
	kubectl delete -f ./operator/
	kubectl delete -f ./adapter/
	kubectl delete -f ./alertmanager/
	kubectl delete -f ./node-exporter/
	kubectl delete -f ./kube-state-metrics/
	kubectl delete -f ./grafana/
	kubectl delete -f ./prometheus/
	kubectl delete -f ./service-monitor/
	kubectl delete -f ./example/
	kubectl delete -f ./namespace.yml
