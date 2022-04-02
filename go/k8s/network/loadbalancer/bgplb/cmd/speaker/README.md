

# LoadBalancer BGP Speaker
(1) bgppeer crd 用户自定义

```yaml

apiVersion: bgplb.k9s.io/v1
kind: BGPPeer
metadata:
  name: nodeA # 必须是 node name
spec:
  peerAddress: 10.0.0.1
  peerASN: 64501
  peerPort: 1790
  sourceAddress: 20.0.0.1
  myASN: 64500
  sourcePort: 1791

```


