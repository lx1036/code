package logs

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strings"
)

var (
	logger = logs.NewLogger(10000)

	sentryLogLevel = logs.LevelInfo
)



func init() {
	logLevel := beego.AppConfig.DefaultInt("LogLevel", 4)
	logger.SetLogger(logs.AdapterConsole, fmt.Sprintf(`{"level":%d}`, logLevel))
	logger.EnableFuncCallDepth(true)
	logger.SetLogFuncCallDepth(3)

	sentryEnable := beego.AppConfig.DefaultBool("SentryEnable", false)
	if sentryEnable {
		sentryLogLevel = beego.AppConfig.DefaultInt("SentryLogLevel", 4)
		//dsn := beego.AppConfig.String("SentryDSN")
		var err error
		//sentryClient, err = raven.New(dsn)
		if err != nil {
			logs.Error(err)
		} else {
			//sentryClient.SetRelease("")
			//sentryClient.SetEnvironment("")
		}
	}
}

func Info(f interface{}, v ...interface{}) {
	logger.Info(formatLog(f, v...))
	/*if sentryClient != nil && sentryLogLevel >= logs.LevelInfo {
		sentryLog(raven.INFO, f, v...)
	}*/
	return
}

// vendor/github.com/astaxie/beego/logs/log.go
func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
