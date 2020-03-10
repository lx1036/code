#!/bin/bash

# Simulation of Integration tests
# makes requests from some pod to jobposting

NAMESPACE=$1

function usepod {
  pod=$1
  echo "Making requests to $pod..."
}

function wait_for_n_running_pods {
	local NPODS=$1
	echo "Waiting for $NPODS running pods in $NAMESPACE"

	local sleep_time=1
	local iter=0
	local found=$(kubectl -n $NAMESPACE get pod | grep Running -c || true)
	until [[ "$found" -eq "$NPODS" ]]; do
		if [[ $((iter++)) -gt $((5*60/$sleep_time)) ]]; then
			echo ""
			echo "Timeout while waiting for $NPODS running pods"
			exit 1
		else
			kubectl -n $NAMESPACE get pod -o wide
			echo -n " [${found}/${NPODS}]"
			sleep $sleep_time
		fi
		found=$(kubectl -n $NAMESPACE get pod | grep Running -c || true)
	done

	kubectl -n $NAMESPACE get pod -o wide
}
# wait_for_n_running_pods 7
# sleep 5

public_pod=`kubectl get pods -l app=jobposting -n $NAMESPACE -o name | sed s/.*\\\///`
private_pod=`kubectl get pods -l app=recruiter -n $NAMESPACE -o name | sed s/.*\\\///`



run () {
  prefix="kubectl -n $NAMESPACE exec $pod sh -- -c"
  cmd="$prefix '$1'"
  eval $cmd
}

get () {
  echo -ne "$1 - "
  run "curl -o /dev/null -w %{http_code} -s http://localhost:9080$1"
  echo
}

post () {
  echo $1
  run "curl -s -o /dev/null -w %{http_code} -X POST http://localhost:9080$1"
}

usepod $public_pod
get "/"
get "/jobs?id=1"
get "/apply?name=joe&jobId=222"

usepod $private_pod
get "/"
get "/applicants?id=1"