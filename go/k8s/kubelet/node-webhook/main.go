package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"k8s-lx1036/k8s/kubelet/node-webhook/node"

	"k8s.io/klog/v2"
)

var (
	certFile = flag.String("tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	keyFile  = flag.String("tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	port     = flag.String("port", "8443", "serve tls listen port")
)

// debug: go run . --kubeconfig=`echo $HOME`/.kube/config --tls-cert-file=../deploy/release/tls/loadbalancer-webhook.pem --tls-private-key-file=../deploy/release/tls/loadbalancer-webhook-key.pem
// debug in local: go run . --kubeconfig=`echo $HOME`/.kube/config (代码临时注销关闭tls)
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	tlsConfig, err := getTlsConfig()
	if err != nil {
		klog.Errorf("tls config err: %v", err)
		return
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		node.Serve(writer, request, node.MutatePod)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	klog.Infof("start listen port: %s", *port)
	server := &http.Server{
		Addr:      fmt.Sprintf(":%s", *port),
		TLSConfig: tlsConfig,
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		klog.Errorf("serve tls err: %v", err)
	}
}

func getTlsConfig() (*tls.Config, error) {
	sCert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}

	return tlsConfig, nil
}
