

# FUSE(Filesystem in Userspace)
Linux 文件系统为每个文件分配两个数据结构：inode 和 dentry。

该 fusefs-client 可以直接本地与 polefs meta/master 组件一起使用。




### Fuse 内核模块
linux fuse: https://github.com/torvalds/linux/blob/master/fs/fuse

### Fuse 用户态模块
fuse 用户态框架：
https://github.com/jacobsa/fuse
https://github.com/bazil/fuse

使用 fuse 用户态框架的 fusefs 程序：fusefs



## 基本概念
inode(index node): 则反映了文件系统对象中的一般元数据信息。Inode is a node in VFS tree.
dentry(directory entry)目录项: 则是反映出某个文件系统对象在全局文件系统树中的位置。总之，就是文件名到 inode 的 mapping。


## 参考文献
**[自制文件系统 — 来看看文件系统的样子](https://mp.weixin.qq.com/s/7qq3AugMKqjlwx26PT20sw)**

**[自制文件系统 — 02 FUSE 框架，开发者的福音](https://mp.weixin.qq.com/s/HvbMxNiVudjNPRgYC8nXyg)**

**[自制文件系统 —— 03 Go实战：hello world 的文件系统](https://mp.weixin.qq.com/s/Yf6yBoEQe6ijMlPgZ6P2sA)**

**[自制文件系统 — 04 HelloFS 进阶 分布式加密文件系统](https://mp.weixin.qq.com/s/rxabk_o5YuSko8SM8EdouA)**

**[自制文件系统 —— 05 总结：一切都为了狙击“文件”](https://mp.weixin.qq.com/s/x7WZmFULZ1AKXu6Kgw0P-Q)**

**[基于Fuse的用户态文件系统性能优化几点建议](https://zhuanlan.zhihu.com/p/68085075)**

**[VFS中的file，dentry和inode](https://bean-li.github.io/vfs-inode-dentry/)**
