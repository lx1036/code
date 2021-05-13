





# 这里 "apiextensions:v1" apiextensions 表示目录，v1 也是目录
all:
	bash "../generate-groups.sh" all \
	k8s-lx1036/k8s/apiserver/apiextension-server/pkg/client k8s-lx1036/k8s/apiserver/apiextension-server/pkg/apis \
	"apiextensions:v1" \
	--output-base "." \
	--go-header-file "../hack/boilerplate/boilerplate.go.txt"


deepcopy-gen:
    # go get k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen -h "../hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/apiextension-server/pkg/apis/apiextensions" \
	-O zz_generated.deepcopy -o . -v 3


defaults-gen:
    # go get k8s.io/code-generator/cmd/defaulter-gen
	defaulter-gen -h "../hack/boilerplate/boilerplate.go.txt" \
	-i "k8s-lx1036/k8s/apiserver/apiextension-server/pkg/apis/apiextensions/v1" \
	--output-file-base zz_generated.defaults -o . -v 3

