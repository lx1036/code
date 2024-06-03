

# bpf debug
为方便 debug bpf，需要开启 `bpf monitor`:
```md
1. 需要找出 minikube node1 中的 nginx pod ip，如 10.244.0.48
2. 需要找出对应的 endpoint id: `cilium endpoint list`，如 1776
3. 然后开启 monitor: `cilium monitor --related-to 1776`
4. minikube node1 上访问 nginx pod ip: `curl -I 10.244.0.48`
```

观察 bpf log:
```shell
-> endpoint 1776 flow 0xbd84431c , identity host->78310 state new ifindex lxc4b5a37a98f73 orig-ip 10.244.0.126: 10.244.0.126:38878 -> 10.244.0.48:80 tcp SYN
-> stack flow 0xbf3a9706 , identity 78310->host state reply ifindex 0 orig-ip 0.0.0.0: 10.244.0.48:80 -> 10.244.0.126:38878 tcp SYN, ACK
-> endpoint 1776 flow 0xbd84431c , identity host->78310 state established ifindex lxc4b5a37a98f73 orig-ip 10.244.0.126: 10.244.0.126:38878 -> 10.244.0.48:80 tcp ACK
-> endpoint 1776 flow 0xbd84431c , identity host->78310 state established ifindex lxc4b5a37a98f73 orig-ip 10.244.0.126: 10.244.0.126:38878 -> 10.244.0.48:80 tcp ACK
-> stack flow 0xbf3a9706 , identity 78310->host state reply ifindex 0 orig-ip 0.0.0.0: 10.244.0.48:80 -> 10.244.0.126:38878 tcp ACK
-> endpoint 1776 flow 0xbd84431c , identity host->78310 state established ifindex lxc4b5a37a98f73 orig-ip 10.244.0.126: 10.244.0.126:38878 -> 10.244.0.48:80 tcp ACK, FIN
-> stack flow 0xbf3a9706 , identity 78310->host state reply ifindex 0 orig-ip 0.0.0.0: 10.244.0.48:80 -> 10.244.0.126:38878 tcp ACK, FIN
-> endpoint 1776 flow 0xbd84431c , identity host->78310 state established ifindex lxc4b5a37a98f73 orig-ip 10.244.0.126: 10.244.0.126:38878 -> 10.244.0.48:80 tcp ACK
```

