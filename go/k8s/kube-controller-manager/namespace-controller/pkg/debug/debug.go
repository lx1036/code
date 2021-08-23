package debug

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func LogAPIResourceList(list []*metav1.APIResourceList) {
	for _, resourceList := range list {
		tmp := ""
		for _, apiResource := range resourceList.APIResources {
			tmp = fmt.Sprintf("Name:%s,Namespaced:%s,Kind:%s,Verbs:%s", apiResource.Name,
				strconv.FormatBool(apiResource.Namespaced),
				apiResource.Kind,
				strings.Join(apiResource.Verbs, "/"),
			)
		}

		//{"level":"debug","msg":"GroupVersion:v1,Name:limitranges,Namespaced:true,Kind:LimitRange,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:extensions/v1beta1,Name:ingresses,Namespaced:true,Kind:Ingress,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:apps/v1,Name:deployments,Namespaced:true,Kind:Deployment,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:events.k8s.io/v1beta1,Name:events,Namespaced:true,Kind:Event,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:authorization.k8s.io/v1,Name:localsubjectaccessreviews,Namespaced:true,Kind:LocalSubjectAccessReview,Verbs:create","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:autoscaling/v1,Name:horizontalpodautoscalers,Namespaced:true,Kind:HorizontalPodAutoscaler,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:batch/v1,Name:jobs,Namespaced:true,Kind:Job,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:batch/v1beta1,Name:cronjobs,Namespaced:true,Kind:CronJob,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:networking.k8s.io/v1,Name:networkpolicies,Namespaced:true,Kind:NetworkPolicy,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:networking.k8s.io/v1beta1,Name:ingresses,Namespaced:true,Kind:Ingress,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:policy/v1beta1,Name:poddisruptionbudgets,Namespaced:true,Kind:PodDisruptionBudget,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:rbac.authorization.k8s.io/v1,Name:roles,Namespaced:true,Kind:Role,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:coordination.k8s.io/v1,Name:leases,Namespaced:true,Kind:Lease,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:discovery.k8s.io/v1beta1,Name:endpointslices,Namespaced:true,Kind:EndpointSlice,Verbs:create/delete/deletecollection/get/list/patch/update/watch","time":"2020-12-17T23:30:49+08:00"}
		//{"level":"debug","msg":"GroupVersion:crd.projectcalico.org/v1,Name:networksets,Namespaced:true,Kind:NetworkSet,Verbs:delete/deletecollection/get/list/patch/create/update/watch","time":"2020-12-17T23:30:49+08:00"}
		log.Debugf("GroupVersion:%s,%s", resourceList.GroupVersion, tmp)
	}
}
