module k8s-lx1036

go 1.13

require (
	github.com/astaxie/beego v1.12.0
	github.com/codingsince1985/checksum v1.1.0
	github.com/getsentry/sentry-go v0.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/google/uuid v1.1.1
	github.com/julienschmidt/httprouter v1.2.0
	github.com/klauspost/cpuid v1.2.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/mholt/certmagic v0.7.3
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/ugorji/go v1.1.7 // indirect
	github.com/urfave/cli v1.22.1
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/apimachinery v0.0.0-20190927035529-0104e33c351d
)

replace github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 => github.com/ugorji/go v1.1.4
