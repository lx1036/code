


# 这里 "apiregistration:v1" apiregistration 表示目录，v1 也是目录
groups:
	bash "../generate-groups.sh" deepcopy,client,lister,informer \
	k8s-lx1036/k8s/apiserver/aggregator-server/pkg/client k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis \
	"apiregistration:v1" \
	--output-base "." \
	--go-header-file "../hack/boilerplate/boilerplate.go.txt"


deepcopy-gen:
    # go get k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen -h "../hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration" \
	-i "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1" \
	-O zz_generated.deepcopy -o . -v 3


defaults-gen:
    # go get k8s.io/code-generator/cmd/defaulter-gen
	defaulter-gen -h "../hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1" \
	--output-file-base zz_generated.defaults -o . -v 3


conversion-gen:
    # go get k8s.io/code-generator/cmd/conversion-gen
	conversion-gen -h "../hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration" \
	-i "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1" \
	--output-file-base zz_generated.conversion -o . -v 3

