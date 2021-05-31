


deepcopy-gen:
    # go get k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen -h "../../../../../../apiserver/hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/v1alpha1" \
	--output-file-base zz_generated.deepcopy -o . -v 3



groups:
	bash "../../../../../../apiserver/generate-groups.sh" client,lister,informer \
	k8s-lx1036/k8s/scheduler/pkg/scheduler/client k8s-lx1036/k8s/scheduler/pkg/scheduler/apis \
	"scheduling:v1alpha1" \
	--output-base "." \
	--go-header-file "../../../../../apiserver/hack/boilerplate/boilerplate.go.txt"

