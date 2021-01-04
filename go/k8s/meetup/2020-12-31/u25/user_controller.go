package main

import "time"

type UserController struct {
	metrics *Metrics
}

func (controller *UserController) register() {

	start := time.Now()
	controller.metrics.recordTimestamp("/register", start)

	// 省略业务代码
	time.Sleep(time.Second * 1)
	// 省略业务代码

	controller.metrics.recordResponseTime("/register", time.Since(start))

}

func (controller *UserController) login(username, password string) {
	start := time.Now()
	controller.metrics.recordTimestamp("/login", start)

	// 省略业务代码
	time.Sleep(time.Second * 1)
	// 省略业务代码

	controller.metrics.recordResponseTime("/login", time.Since(start))
}
