


## Cilium






## Cilium 功能点
Cilium 通过 BGP 宣告 pod cidr：https://github.com/cilium/cilium/pull/16525
Cilium 多个 ippool 问题，目前还不支持，一个K8s里部署多个Cilium不同网段实例，貌似也不支持：https://github.com/cilium/cilium/issues/13227

本地安装 cilium CLI:
```shell
curl -L --remote-name-all https://github.com/cilium/cilium-cli/releases/latest/download/cilium-darwin-amd64.tar.gz{,.sha256sum}
shasum -a 256 -c cilium-darwin-amd64.tar.gz.sha256sum
sudo tar xzvfC cilium-darwin-amd64.tar.gz /usr/local/bin
rm cilium-darwin-amd64.tar.gz{,.sha256sum}
```


## Troubleshoot
(1)Cilium 支持多个网段问题，或者k8s 里部署多个不同网段的 Cilium实例？


