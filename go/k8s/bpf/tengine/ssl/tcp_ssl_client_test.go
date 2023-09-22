package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 该测试可以
func TestHTTPSSLClient(test *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 跳过客户端验证服务端证书，仅用于测试环境
		// 设置SNI字段, 只有是域名时(IP 不行)，Handshake Protocol: Client Hello 报文里才有 Extension: server_name
		ServerName: "kubernetes.default.svc.cluster",
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get("https://127.0.0.1:5005")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印响应内容
	fmt.Println("Response:", string(body))
}
func TestTCPSSLClient(test *testing.T) {
	caPemPath, _ := filepath.Abs("ca.pem")
	rootPEM, _ := os.ReadFile(caPemPath)
	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:5005", &tls.Config{
		RootCAs: roots,
		// 跳过客户端验证服务端证书 CA 不在 CA list 中，仅用于测试环境，不需要验证 "x509: certificate signed by unknown authority"
		InsecureSkipVerify: true,
		ServerName:         "kubernetes.default.svc.cluster",
	})
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	_, err = conn.Write([]byte("hello server"))
	if err != nil {
		panic("failed to write: " + err.Error())
	}
	serverData, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	logrus.Infof("server data is %s", string(serverData))

}

func TestHTTPSSLServer(test *testing.T) {
	// 加载SSL证书和私钥
	serverPem, _ := filepath.Abs("../../ssl/server.pem")
	serverKeyPem, _ := filepath.Abs("../../ssl/server-key.pem")
	cert, err := tls.LoadX509KeyPair(serverPem, serverKeyPem)
	if err != nil {
		fmt.Println("Error loading certificate:", err)
		return
	}

	// 创建一个自定义的TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// 创建一个自定义的HTTP服务器
	server := &http.Server{
		Addr:      "localhost:5005",
		TLSConfig: tlsConfig,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, client!"))
	})

	fmt.Println("Server is listening on localhost:5005")

	// 启动服务器
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}

func TestTCPSSLServer(test *testing.T) {
	// 加载SSL证书和私钥
	serverPem, _ := filepath.Abs("server.pem")
	serverKeyPem, _ := filepath.Abs("server-key.pem")
	cert, err := tls.LoadX509KeyPair(serverPem, serverKeyPem)
	if err != nil {
		fmt.Println("Error loading certificate:", err)
		return
	}

	// 创建一个自定义的TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// 创建一个监听器
	listener, err := tls.Listen("tcp", "localhost:5005", tlsConfig)
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on localhost:5005")

	for {
		// 接受连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取客户端请求
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	requestData, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	logrus.Infof("client data is %s", string(requestData))

	// 处理请求并返回响应
	response := []byte("HTTP/1.1 200 OK\\r\\n\\r\\nHello, client!")
	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("Error writing response:", err)
		return
	}
}
