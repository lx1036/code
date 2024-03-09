package main

import (
	"github.com/sirupsen/logrus"
	"net"
)

// *[解析 Golang 网络 IO 模型之 EPOLL](https://mp.weixin.qq.com/s/xt0Elppc_OaDFnTI_tW3hg)*

// curl localhost:8080
func main() {
	// 创建一个 tcp 端口监听器
	l, _ := net.Listen("tcp", ":8080")
	// 主动轮询模型
	for {
		// 等待 tcp 连接到达
		conn, _ := l.Accept()
		// 开启一个 goroutine 负责一笔客户端请求的处理
		go serve(conn)
	}
}

func serve(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	_, err := conn.Read(buf)
	//_, err := io.ReadFull(conn, buf)
	//buf, err := io.ReadAll(conn)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Infof("read %s from client", string(buf))
	}

	_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\npong"))
}
