package common

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
	"runtime"
)

// 打开url
func openURL(urlAddr string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", " /c start "+urlAddr)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", urlAddr)
	} else {
		return
	}
	err := cmd.Start()
	if err != nil {
		//logger.Log.Errorw("打开浏览器错误", "err", err)
		log.Error(err)
	}
}
