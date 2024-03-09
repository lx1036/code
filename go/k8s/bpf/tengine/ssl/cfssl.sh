

# @see https://kubernetes.io/docs/tasks/administer-cluster/certificates/
# sh cfssl.sh

curl -L https://github.com/cloudflare/cfssl/releases/download/v1.6.4/cfssl_1.6.4_darwin_amd64 -o cfssl
chmod +x cfssl
mv cfssl /usr/local/bin/
curl -L https://github.com/cloudflare/cfssl/releases/download/v1.6.4/cfssljson_1.6.4_darwin_amd64 -o cfssljson
chmod +x cfssljson
mv cfssljson /usr/local/bin/
curl -L https://github.com/cloudflare/cfssl/releases/download/v1.6.4/cfssl-certinfo_1.6.4_darwin_amd64 -o cfssl-certinfo
chmod +x cfssl-certinfo
mv cfssl-certinfo /usr/local/bin/
