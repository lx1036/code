
label=$1

if [ ! $label ]; then
    echo "label should be input, e.g. app=nginx-demo-1"
    exit 1
fi

pods=$(kubectl get pods -l ${label} -o jsonpath="{.items[*].metadata.name}")

for pod in ${pods}; do
	# shellcheck disable=SC2046
	echo $(kubectl get pods/${pod} -o jsonpath="{.status.podIP}")
done
