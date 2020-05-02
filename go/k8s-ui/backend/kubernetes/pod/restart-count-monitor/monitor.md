
# Problem
监听特定 namespace 下所有 deployments 内的所有 pods 的重启次数(该 pod 内的所有 containers 重启次数总和)，如果重启次数大于设定阈值，
就给钉钉发送该 namespace/deployment/pod 的详细信息，包括 deployment name 和 pod name，以及各个 container 的分别重启次数，和该 pod 最新十条 event。 


# Solution


