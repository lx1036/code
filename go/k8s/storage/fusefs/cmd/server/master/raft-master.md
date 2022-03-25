
# raft-master 组件
该组件仅仅是一个 raft cluster，重要功能包括：
* 提供一个 http api，供 meta node 注册。
* 提供一个 http api，供 create/delete volume。在创建 volume 时，还需要根据 meta node 使用率选择 3 个 meta node，然后在每一个 meta node 上分别创建 3 次包含 inode(start, end) 的 meta partition。

# TODO
* 把 tiglabs/raft 替换为 hashicorp/raft

