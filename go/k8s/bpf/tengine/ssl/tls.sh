

# sh tls.sh

# 生成 ca.pem/ca-key.pem 根证书
cfssl gencert -initca ca-csr.json | cfssljson -bare ca

# 生成服务端证书 server.pem/server-key.pem
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem \
     --config=ca-config.json -profile=kubernetes \
     server-csr.json | cfssljson -bare server
