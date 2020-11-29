

certs:
	ansible-playbook -i inventory/wayne.conf playbook/certs.yaml -e "group=master"

etcd-cluster:
	ansible-playbook -i inventory/etcd.conf 02-etcd.yaml -e "group=master"

etcd-demo:
	ansible-playbook -i inventory/wayne.conf 02-etcd.yaml -e "group=wayne"
