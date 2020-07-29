module k8s-lx1036

go 1.13

require (
	bou.ke/monkey v1.0.2
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/astaxie/beego v1.12.1
	github.com/caddyserver/caddy v1.0.4
	github.com/codingsince1985/checksum v1.1.0
	github.com/containerd/containerd v1.3.2
	github.com/coreos/etcd v3.3.22+incompatible // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/getsentry/sentry-go v0.3.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-logr/logr v0.1.0
	github.com/go-redis/redis/v7 v7.0.0-beta.4
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gohouse/gorose/v2 v2.1.3
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/jinzhu/gorm v1.9.11
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kavu/go_reuseport v1.5.0 // indirect
	github.com/klauspost/cpuid v1.2.0
	github.com/labstack/gommon v0.3.0
	github.com/libp2p/go-reuseport v0.0.1
	github.com/mholt/certmagic v0.8.3
	github.com/miekg/dns v1.1.22
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moby/ipvs v1.0.1
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32
	github.com/olivere/elastic/v7 v7.0.9
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/romanyx/polluter v1.2.2
	github.com/rs/cors v1.7.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/evio v1.0.2
	github.com/tomasen/fcgi_client v0.0.0-20180423082037-2bb3d819fd19
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	go.etcd.io/etcd v3.3.22+incompatible
	go.uber.org/zap v1.14.1
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527
	google.golang.org/grpc v1.27.0
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/sohlich/elogrus.v7 v7.0.0
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v3 v3.0.0-20190905181640-827449938966
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.18.3
	k8s.io/sample-controller v0.18.3
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	k8s.io/api => k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.3
	k8s.io/apiserver => k8s.io/apiserver v0.18.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.3
	k8s.io/client-go => k8s.io/client-go v0.18.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.3
	k8s.io/code-generator => k8s.io/code-generator v0.18.3
	k8s.io/component-base => k8s.io/component-base v0.18.3
	k8s.io/cri-api => k8s.io/cri-api v0.18.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.3
	k8s.io/kubectl => k8s.io/kubectl v0.18.3
	k8s.io/kubelet => k8s.io/kubelet v0.18.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.3
	k8s.io/metrics => k8s.io/metrics v0.18.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.3
)
