


# 这里 "apiregistration:v1" apiregistration 表示目录，v1 也是目录
groups:
	bash "../../../../../../../apiserver/generate-groups.sh" deepcopy,client,lister,informer \
	k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis \
	"autoscaling.k9s.io:v1" \
	--output-base "." \
	--go-header-file "../../../../../../../apiserver/hack/boilerplate/boilerplate.go.txt" --v=5
