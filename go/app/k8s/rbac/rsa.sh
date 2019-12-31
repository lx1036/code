
openssl genrsa -out backend.key 2048
openssl req -new -key backend.key -out backend.csr -subj "/CN=backend/O=dev"
openssl x509 -req -in backend.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -out backend.crt -days 365
