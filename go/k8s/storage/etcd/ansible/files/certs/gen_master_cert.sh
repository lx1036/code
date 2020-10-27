


./gen_master_csr.sh $1
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes \
    $1-csr.json | cfssljson -bare $1
