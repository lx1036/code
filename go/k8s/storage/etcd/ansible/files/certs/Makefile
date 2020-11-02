
ca:
	cfssl gencert -initca ca-csr.json | cfssljson -bare ca

ca-renew:
	cfssl gencert -renewca -ca=ca.pem -ca-key=ca-key.pem | cfssljson -bare renewca

