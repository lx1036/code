

curl -X POST http://localhost:8001/services/baidu-service/plugins \
    --data "name=mtls-auth"  \
    --data "config.ca_certificates=`[{"id": "fdac360e-7b19-4ade-a553-6dd22937c82f" }, { "id": "aabc360e-7b19-5aab-1231-6da229a7b82f"} ]`" \
    --data "config.authenticated_group_by=CN"
