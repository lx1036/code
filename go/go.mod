module k8s-lx1036

go 1.16

require (
	bazil.org/fuse v0.0.0-20200524192727-fb710f7dfd05
	bou.ke/monkey v1.0.2
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/BurntSushi/toml v0.3.1
	github.com/Shopify/sarama v1.19.0
	github.com/astaxie/beego v1.12.1
	github.com/aws/aws-sdk-go v1.38.49
	github.com/bep/debounce v1.2.0
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/caddyserver/caddy v1.0.4
	github.com/codingsince1985/checksum v1.1.0 // indirect
	github.com/container-storage-interface/spec v1.3.0
	github.com/containerd/cgroups v0.0.0-20200531161412-0dbf7f05ba59
	github.com/containerd/containerd v1.4.4
	github.com/containernetworking/cni v0.8.0
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v1.4.2-0.20200309214505-aa6a9891b09c
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0 // indirect
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
	github.com/gogo/protobuf v1.3.2
	github.com/gohouse/gorose/v2 v2.1.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/btree v1.0.1
	github.com/google/cadvisor v0.37.5
	github.com/google/go-querystring v1.0.0
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hanwen/go-fuse/v2 v2.1.1-0.20210611132105-24a1dfe6b4f8
	github.com/hashicorp/go-msgpack v0.5.5
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/raft v1.3.1
	github.com/jacobsa/fuse v0.0.0-20210606185441-fac69e018fad // indirect
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd
	github.com/jacobsa/syncutil v0.0.0-20180201203307-228ac8e5a6c3
	github.com/jacobsa/timeutil v0.0.0-20170205232429-577e5acbbcf6 // indirect
	github.com/jedib0t/go-pretty v4.3.0+incompatible // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/julienschmidt/httprouter v1.3.0
	github.com/jwhited/corebgp v0.2.0
	github.com/kavu/go_reuseport v1.5.0 // indirect
	github.com/klauspost/cpuid v1.2.1
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/kubernetes-csi/external-snapshotter/client/v3 v3.0.0
	github.com/kubernetes-sigs/custom-metrics-apiserver v0.0.0-20210311094424-0ca2b1909cdc
	github.com/kylelemons/godebug v1.1.0
	github.com/labstack/gommon v0.3.0
	github.com/libp2p/go-reuseport v0.0.1
	github.com/lni/dragonboat/v3 v3.3.1
	github.com/lni/goutils v1.3.0
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/mholt/certmagic v0.8.3
	github.com/miekg/dns v1.1.26
	github.com/mitchellh/mapstructure v1.4.1
	github.com/moby/ipvs v1.0.1
	github.com/moby/sys/mountinfo v0.4.1 // indirect
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32
	github.com/olivere/elastic/v7 v7.0.9 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/opencontainers/runc v1.0.0-rc93
	github.com/opencontainers/runtime-spec v1.0.3-0.20200929063507-e6143ca7d51d
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627
	github.com/pkg/errors v0.9.1
	github.com/projectcalico/libcalico-go v1.7.2-0.20201119205058-b367043ede58
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.26.0
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/robfig/cron v1.1.0
	github.com/romanyx/polluter v1.2.2
	github.com/rs/cors v1.7.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v0.0.0-20181127023241-353a9fca669c // indirect
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c
	github.com/tidwall/evio v1.0.2
	github.com/tidwall/wal v0.1.4
	github.com/tiglabs/raft v0.0.0-20200304095606-b25a44ad8b33
	github.com/tomasen/fcgi_client v0.0.0-20180423082037-2bb3d819fd19
	github.com/urfave/cli v1.22.4
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20200520041808-52d707b772fe
	go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.etcd.io/etcd/raft/v3 v3.5.0 // indirect
	go.uber.org/zap v1.17.0
	golang.org/x/build v0.0.0-20190927031335-2835ba2e683f
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210520170846-37e1c6afe023
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/sohlich/elogrus.v7 v7.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.22.1
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.22.1
	k8s.io/apiserver v0.22.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cloud-provider v0.22.1
	k8s.io/code-generator v0.22.1
	k8s.io/component-base v0.22.1
	k8s.io/component-helpers v0.22.1
	k8s.io/cri-api v0.0.0
	k8s.io/gengo v0.0.0-20210203185629-de9496dff47b
	k8s.io/klog/v2 v2.9.0
	k8s.io/kube-aggregator v0.17.3
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e
	k8s.io/kube-scheduler v0.0.0
	k8s.io/kubelet v0.0.0
	k8s.io/kubernetes v1.19.7
	k8s.io/metrics v0.22.1
	k8s.io/sample-controller v0.22.1
	k8s.io/utils v0.0.0-20210707171843-4b05e18ac7d9
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.6.2 // indirect
	sigs.k8s.io/metrics-server v0.4.4
	volcano.sh/apis v0.0.0-20210603070204-70005b2d502a
)

replace (
	github.com/hanwen/go-fuse/v2 v2.1.1-0.20210611132105-24a1dfe6b4f8 => github.com/juicedata/go-fuse/v2 v2.1.1-0.20210629082323-0ec79f5f0a45
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.2.2
	k8s.io/api => k8s.io/api v0.22.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.1
	k8s.io/apiserver => k8s.io/apiserver v0.22.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.1
	k8s.io/client-go => k8s.io/client-go v0.22.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.1
	k8s.io/code-generator => k8s.io/code-generator v0.22.1
	k8s.io/component-base => k8s.io/component-base v0.22.1
	k8s.io/cri-api => k8s.io/cri-api v0.22.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.1
	k8s.io/kubectl => k8s.io/kubectl v0.22.1
	k8s.io/kubelet => k8s.io/kubelet v0.22.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.1
	k8s.io/metrics => k8s.io/metrics v0.22.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.1
)
