

**[configmap-controller](https://github.com/fabric8io/configmapcontroller)**:
This controller watches for changes to ConfigMap objects and performs rolling upgrades on their associated deployments for apps 
which are not capable of watching the ConfigMap and updating dynamically.

This is particularly useful if the ConfigMap is used to define environment variables - 
or your app cannot easily and reliably watch the ConfigMap and update itself on the fly.
