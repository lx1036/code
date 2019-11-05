module k8s-lx1036

go 1.13

require (
	github.com/astaxie/beego v1.12.0
	github.com/codingsince1985/checksum v1.1.0
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/getsentry/sentry-go v0.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kavu/go_reuseport v1.4.0 // indirect
	github.com/klauspost/cpuid v1.2.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/libp2p/go-reuseport v0.0.1
	github.com/mholt/certmagic v0.7.3
	github.com/prometheus/client_golang v0.9.3
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/evio v1.0.2
	github.com/ugorji/go v1.1.7 // indirect
	github.com/urfave/cli v1.22.1
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.0.0-20191003035645-10e821c09743
	k8s.io/apimachinery v0.0.0-20191025225532-af6325b3a843
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
)

replace (
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 => github.com/ugorji/go v1.1.4
	k8s.io/api => k8s.io/api v0.0.0-20191025225708-5524a3672fbb
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191025225532-af6325b3a843
)
