


# 这里 "apiregistration:v1" apiregistration 表示目录，v1 也是目录

# 突然安装 client,lister,informer 失败，是因为 k8s.io/code-generator v0.19.7 版本问题
groups:
	bash "../../../apiserver/generate-groups.sh" deepcopy,client,lister,informer \
	k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/client k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis \
	"sparkoperator.k9s.io:v1" \
	--output-base "." \
	--go-header-file "../../../apiserver/hack/boilerplate/boilerplate.go.txt" --v=3
