package reporter

import (
	"time"
	
	"k8s-lx1036/k8s/meetup/2020-12-31/storage"
	
	"k8s.io/apimachinery/pkg/util/wait"
)

type Reporter interface {

}


type Console struct {
	
}

func (console *Console) StartRepeatedReport()  {
	
	
	go wait.Until(func() {
	
	}, period, stopCh)
}

func NewConsole(storage storage.Storage) Reporter {
	// 1. 按照给定的时间区间，从数据库中拉取数据
	allRequestInfos := storage.getAllRequestInfosByDuration(start, end)
	
	// 2. 根据原始数据，得到统计数据
	requestStatMap := map[string]RequestStat{}
	
	// 3. 将数据显示到终端
}


type Email struct {
	ToAddress []string
}

func (email *Email) AddToAddress(address []string)  {
	email.ToAddress = append(email.ToAddress, address...)
}

func (email *Email) StartDailyReport() {

	go wait.Until(func() {
	
	}, time.Hour * 24, stopCh)
}

func NewMail(storage storage.Storage) Reporter {
	requestInfos := storage.GetRequestInfos()
	
}






