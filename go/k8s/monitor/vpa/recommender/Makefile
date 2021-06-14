


# 这里 "apiregistration:v1" apiregistration 表示目录，v1 也是目录

# 突然安装 client,lister,informer 失败，是因为 k8s.io/code-generator v0.19.7 版本问题
groups:
	bash "../../../apiserver/generate-groups.sh" deepcopy,client,lister,informer \
	k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis \
	"autoscaling.k9s.io:v1" \
	--output-base "." \
	--go-header-file "../../../apiserver/hack/boilerplate/boilerplate.go.txt" --v=3


# 自动化生成 crd
# controller-gen -hhh
controller-gen:
	controller-gen crd:trivialVersions=true,maxDescLen=0,crdVersions=v1 paths="./pkg/apis/autoscaling.k9s.io/v1" output:crd:artifacts:config=./deploy/release

