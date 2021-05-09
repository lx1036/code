


deepcopy-gen:
    # go get k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen -h "hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/pkg/apis/example" \
	-i "k8s-lx1036/k8s/apiserver/pkg/apis/example/v1" \
	--output-file-base zz_generated.deepcopy -o . -v 3



defaults-gen:
    # go get k8s.io/code-generator/cmd/defaulter-gen
	defaulter-gen -h "hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/pkg/apis/example/v1" \
	--output-file-base zz_generated.defaults -o . -v 3



conversion-gen:
    # go get k8s.io/code-generator/cmd/conversion-gen
	conversion-gen -h "hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/pkg/apis/example/v1" \
	--output-file-base zz_generated.conversion -o . -v 3

