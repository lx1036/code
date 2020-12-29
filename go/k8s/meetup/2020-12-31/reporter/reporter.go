package reporter

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

type Console struct {
	
}

func (console *Console) StartRepeatedReport()  {
	
	
	go wait.Until(func() {
	
	}, period, stopCh)
}


type Email struct {
	
}

func (email *Email) AddToAddress(address []string)  {

}

func (email *Email) StartDailyReport() {

	go wait.Until(func() {
	
	}, time.Hour * 24, stopCh)
}








