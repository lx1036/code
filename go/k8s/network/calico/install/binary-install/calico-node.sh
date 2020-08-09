# Calico 二进制安装: https://juejin.im/post/6844904071611023374



# 从 ctl 镜像中 copy 出来，也可以从 git release 下载。
CALICO_CTL_IMAGE=calico/ctl:v3.12.0
docker pull ${CALICO_CTL_IMAGE}
docker create --name calico-ctl-create ${CALICO_CTL_IMAGE}
sudo docker cp calico-ctl-create:/calicoctl /usr/local/bin/calicoctl
docker rm calico-ctl-create

ETCD_ENPOINTS="http://10.9.144.49:2379"
# 这里配置calicoctl连接数据源，根据实际情况变化
sudo mkdir -p /etc/calico
sudo sh -c "cat > /etc/calico/calicoctl.cfg" << EOF
apiVersion: projectcalico.org/v3
kind: CalicoAPIConfig
metadata:
spec:
  datastoreType: etcdv3
  etcdEndpoints: ${ETCD_ENPOINTS}
EOF


CALICO_NODE_IMAGE=calico/node:v3.12.0
docker pull ${CALICO_NODE_IMAGE}
docker create --name calico-node-create  ${CALICO_NODE_IMAGE}
# calico-node(felix confd)
sudo docker cp calico-node-create:/bin/calico-node /usr/local/bin/calico-node
# felix,felix 里面所需要的环境变量与calico node 重叠但是名称不同，所以直接使用脚本方式。
# 详见：https://github.com/projectcalico/node/blob/release-v3.12/filesystem/etc/service/available/felix/run
sudo docker cp calico-node-create:/etc/service/available/felix/run /usr/local/bin/calico-felix
# bird，用于节点互联的组件，使用由confd生成的配置文件。
sudo docker cp calico-node-create:/usr/bin/bird /usr/local/bin/bird
# confd configurations，confd 的模板等，confd 从这些模板动态生成 bird 等所需的配置文件。
sudo docker cp calico-node-create:/etc/calico/confd /etc/calico/confd
docker rm calico-node-create


# 主要配置一个变量，其他变量去 calico.env 里面修改。
# all support env,default values are referenced: https://docs.projectcalico.org/reference/node/configuration
sudo sh -c "cat > /etc/calico/calico.env" << EOF
NODENAME=$(hostname)
NO_DEFAULT_POOLS=false
IP=""
IP6=""
IP_AUTODETECTION_METHOD=first-found
IP6_AUTODETECTION_METHOD=first-found
DISABLE_NODE_IP_CHECK=false
AS=
CALICO_DISABLE_FILE_LOGGING=false
CALICO_ROUTER_ID=""
DATASTORE_TYPE=etcdv3
WAIT_FOR_DATASTORE=false
CALICO_NETWORKING_BACKEND=bird
CALICO_IPV4POOL_CIDR=192.168.0.0/16
CALICO_IPV6POOL_CIDR=""
CALICO_IPV4POOL_BLOCK_SIZE=26
CALICO_IPV6POOL_BLOCK_SIZE=122
CALICO_IPV4POOL_IPIP=Always
CALICO_IPV4POOL_VXLAN=Never
CALICO_IPV4POOL_NAT_OUTGOING=true
CALICO_IPV6POOL_NAT_OUTGOING=false
CALICO_IPV4POOL_NODE_SELECTOR="all()"
CALICO_IPV6POOL_NODE_SELECTOR="all()"
CALICO_STARTUP_LOGLEVEL=ERROR
CLUSTER_TYPE=""
ETCD_ENDPOINTS=${ETCD_ENPOINTS}
ETCD_DISCOVERY_SRV=""
ETCD_KEY_FILE=""
ETCD_CERT_FILE=""
ETCD_CA_CERT_FILE=""
CALICO_MANAGE_CNI=false
FELIX_LOGSEVERITYSCREEN=INFO
EOF


sudo sh -c "cat > /etc/systemd/system/calico-felix.service" << EOF
[Unit]
Description=Calico Felix agent
After=syslog.target network.target

[Service]
User=root
EnvironmentFile=/etc/calico/calico.env
ExecStartPre=/usr/bin/mkdir -p /var/run/calico
ExecStartPre=/usr/local/bin/calico-node -startup
ExecStart=/usr/local/bin/calico-felix
KillMode=process
Restart=on-failure
LimitNOFILE=32000

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl enable calico-felix
sudo systemctl start calico-felix


# 安装 calico-felix service
sudo sh -c "cat > /etc/systemd/system/calico-confd.service" << EOF
[Unit]
Description=Calico confd
After=syslog.target network.target

[Service]
User=root
EnvironmentFile=/etc/calico/calico.env
ExecStartPre=/usr/bin/mkdir -p /var/run/calico
ExecStart=/usr/local/bin/calico-node -confd
KillMode=process
Restart=on-failure
LimitNOFILE=32000

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl enable calico-confd
sudo systemctl start calico-confd

# 安装 calico-confd service
sudo sh -c "cat > /etc/systemd/system/calico-confd.service" << EOF
[Unit]
Description=Calico confd
After=syslog.target network.target

[Service]
User=root
EnvironmentFile=/etc/calico/calico.env
ExecStartPre=/usr/bin/mkdir -p /var/run/calico
ExecStart=/usr/local/bin/calico-node -confd
KillMode=process
Restart=on-failure
LimitNOFILE=32000

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl enable calico-confd
sudo systemctl start calico-confd

# 安装 bird service
sudo sh -c "cat > /etc/systemd/system/bird.service" << EOF
[Unit]
Description=BIRD internet routing daemon
After=syslog.target network.target

[Service]
User=root
EnvironmentFile=/etc/calico/calico.env
ExecStartPre=/usr/bin/mkdir -p /var/run/calico
ExecStart=/usr/local/bin/bird -R -s /var/run/calico/bird.ctl -d -c /etc/calico/confd/config/bird.cfg
KillMode=process
Restart=on-failure
LimitNOFILE=32000

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl enable bird
sudo systemctl start bird


# calico libnetwork-plugin 安装
CALICO_LIBNETWORK_PLUGIN_IMAGE=calico/libnetwork-plugin
docker pull ${CALICO_LIBNETWORK_PLUGIN_IMAGE}
docker create --name calico-libnetwork-plugin-create ${CALICO_LIBNETWORK_PLUGIN_IMAGE}
sudo docker cp calico-libnetwork-plugin-create:/libnetwork-plugin /usr/local/bin/calico-libnetwork-plugin
docker rm calico-libnetwork-plugin-create

sudo sh -c "cat > /etc/systemd/system/calico-libnetwork-plugin.service" << EOF
[Unit]
Description=Calico libnetwork plugin
After=syslog.target network.target calico-felix.service
Requires=calico-felix.service

[Service]
User=root
EnvironmentFile=/etc/calico/calico.env
ExecStartPre=/usr/bin/mkdir -p /var/run/calico
ExecStart=/usr/local/bin/calico-libnetwork-plugin
KillMode=process
Restart=on-failure
LimitNOFILE=32000

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl enable calico-libnetwork-plugin
sudo systemctl start calico-libnetwork-plugin


sudo sh -c "cat > /etc/calico/global-network-policy-allow-all.yaml" << EOF
apiVersion: projectcalico.org/v3
kind: GlobalNetworkPolicy
metadata:
  name: allow-all
spec:
  selector: all()
  ingress:
  - action: Allow
  egress:
  - action: Allow
EOF
sudo calicoctl apply -f /etc/calico/global-network-policy-allow-all.yaml

# 在这之前需要设置 docker.service 增加连接至etcd相关配置， 例如： --cluster-store=etcd://192.168.2.21:2379
# 修改 /usr/lib/systemd/system/docker.service (centos下，不同系统路径和配置可能不同)，
#在 ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock 行增加 --cluster-store=etcd://192.168.2.21:2379
systemctl daemon-reload
systemctl restart docker
docker network create --driver calico --ipam-driver calico-ipam --subnet 192.168.0.0/16 cali_net
