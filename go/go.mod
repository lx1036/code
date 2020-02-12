module k8s-lx1036

go 1.13

require (
	bou.ke/monkey v1.0.2
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/astaxie/beego v1.12.1
	github.com/caddyserver/caddy v1.0.4
	github.com/codingsince1985/checksum v1.1.0
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/getsentry/sentry-go v0.3.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-redis/redis/v7 v7.0.0-beta.4
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gohouse/gorose/v2 v2.1.3
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kavu/go_reuseport v1.4.0 // indirect
	github.com/klauspost/cpuid v1.2.0
	github.com/labstack/gommon v0.3.0
	github.com/libp2p/go-reuseport v0.0.1
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mholt/certmagic v0.8.3
	github.com/miekg/dns v1.1.15
	github.com/mitchellh/mapstructure v1.1.2
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32
	github.com/olivere/elastic/v7 v7.0.9
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/romanyx/polluter v1.2.2
	github.com/rs/cors v1.7.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.2
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/evio v1.0.2
	github.com/urfave/cli v1.22.1
	golang.org/x/crypto v0.0.0-20200208060501-ecb85df21340 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/sys v0.0.0-20200116001909-b77594299b42
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/sohlich/elogrus.v7 v7.0.0
	k8s.io/api v0.0.0-20191003035645-10e821c09743
	k8s.io/apimachinery v0.0.0-20191025225532-af6325b3a843
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
)

replace (
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 => github.com/ugorji/go v1.1.4
	k8s.io/api => k8s.io/api v0.0.0-20191025225708-5524a3672fbb
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191025225532-af6325b3a843
	k8s.io/client-go v11.0.0+incompatible => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
)
