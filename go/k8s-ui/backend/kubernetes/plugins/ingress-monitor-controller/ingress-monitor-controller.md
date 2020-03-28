


**[ingress-monitor-controller](https://stakater.com/projects/imc.html)**: 
A kubernetes Controller to watch your ingresses and create liveness alerts for your endpoints.

**[stakater/IngressMonitorController](https://github.com/stakater/IngressMonitorController)**: 
A Kubernetes/Openshift controller to watch ingresses/routes and create liveness alerts for your apps/microservices in Uptime checkers.

# Problem
We want to monitor ingresses in a kubernetes cluster and routes in openshift cluster via any uptime checker 
but the problem is having to manually check for new ingresses or routes / removed ingresses or routes and add them to the checker or remove them.

# Solution
This controller will continuously watch ingresses/routes in specific or all namespaces, and automatically add / remove monitors in any of the uptime checkers. 
With the help of this solution, you can keep a check on your services and see whether they're up and running and live, without worrying about manually registering them on the Uptime checker.
