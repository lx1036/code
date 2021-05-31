


deepcopy-gen:
    # go get k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen -h "../../../../../../apiserver/hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/config" \
	-i "k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/config/v1beta1" \
	--output-file-base zz_generated.deepcopy -o . -v 3

conversion-gen:
    # go get k8s.io/code-generator/cmd/conversion-gen
	conversion-gen -h "../../../../../../apiserver/hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/config/v1beta1" \
	--output-file-base zz_generated.conversion -o . -v 3

