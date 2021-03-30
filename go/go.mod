module k8s-lx1036

go 1.16

require (
	bou.ke/monkey v1.0.2
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/Azure/go-autorest/autorest v0.11.15 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/Shopify/sarama v1.19.0
	github.com/astaxie/beego v1.12.1
	github.com/aws/aws-sdk-go v1.35.24
	github.com/bep/debounce v1.2.0
	github.com/caddyserver/caddy v1.0.4
	github.com/codingsince1985/checksum v1.1.0
	github.com/container-storage-interface/spec v1.3.0
	github.com/containerd/cgroups v0.0.0-20200531161412-0dbf7f05ba59
	github.com/containerd/containerd v1.4.4
	github.com/containernetworking/cni v0.8.0
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/emicklei/go-restful v2.14.3+incompatible
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/getsentry/sentry-go v0.3.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-logr/logr v0.4.0
	github.com/go-redis/redis/v7 v7.0.0-beta.4
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/gohouse/gorose/v2 v2.1.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.3
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/btree v1.0.0
	github.com/google/cadvisor v0.37.3
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jacobsa/fuse v0.0.0-20201216155545-e0296dec955f
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/jinzhu/gorm v1.9.11
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kavu/go_reuseport v1.5.0 // indirect
	github.com/klauspost/cpuid v1.2.0
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/kubernetes-csi/external-snapshotter/client/v3 v3.0.0
	github.com/kubernetes-sigs/custom-metrics-apiserver v0.0.0-20210311094424-0ca2b1909cdc
	github.com/labstack/gommon v0.3.0
	github.com/libp2p/go-reuseport v0.0.1
	github.com/mholt/certmagic v0.8.3
	github.com/miekg/dns v1.1.22
	github.com/mitchellh/mapstructure v1.1.2
	github.com/moby/ipvs v1.0.1
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32
	github.com/olivere/elastic/v7 v7.0.9
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.1
	github.com/opencontainers/runc v1.0.0-rc91.0.20200707015106-819fcc687efb
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627
	github.com/pkg/errors v0.9.1
	github.com/projectcalico/libcalico-go v1.7.2-0.20201119205058-b367043ede58
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/robfig/cron v1.1.0
	github.com/romanyx/polluter v1.2.2
	github.com/rs/cors v1.7.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v0.0.0-20181127023241-353a9fca669c
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c
	github.com/tidwall/evio v1.0.2
	github.com/tiglabs/raft v0.0.0-20200304095606-b25a44ad8b33
	github.com/tomasen/fcgi_client v0.0.0-20180423082037-2bb3d819fd19
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20200520041808-52d707b772fe
	go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	go.uber.org/zap v1.14.1
	golang.org/x/build v0.0.0-20190927031335-2835ba2e683f
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20210324051608-47abb6519492
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/grpc v1.29.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/sohlich/elogrus.v7 v7.0.0
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery v0.20.2
	k8s.io/apiserver v0.20.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.20.5 // indirect
	k8s.io/component-base v0.20.0
	k8s.io/component-helpers v0.20.2
	k8s.io/cri-api v0.0.0
	k8s.io/gengo v0.0.0-20210203185629-de9496dff47b // indirect
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-scheduler v0.0.0
	k8s.io/kubelet v0.0.0
	k8s.io/kubernetes v1.19.7
	k8s.io/metrics v0.20.0
	k8s.io/sample-controller v0.19.7
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	k8s.io/api => k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.7
	k8s.io/apiserver => k8s.io/apiserver v0.19.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.7
	k8s.io/client-go => k8s.io/client-go v0.19.7
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.7
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.7
	k8s.io/code-generator => k8s.io/code-generator v0.19.7
	k8s.io/component-base => k8s.io/component-base v0.19.7
	k8s.io/cri-api => k8s.io/cri-api v0.19.7
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.7
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.7
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.7
	k8s.io/kubectl => k8s.io/kubectl v0.19.7
	k8s.io/kubelet => k8s.io/kubelet v0.19.7
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.7
	k8s.io/metrics => k8s.io/metrics v0.19.7
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.7
)
