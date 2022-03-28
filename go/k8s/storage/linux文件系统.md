
# Linux 文件系统
Linux 文件系统为每个文件分配两个数据结构：inode 和 dentry。

inode(index node): 文件的索引节点。文件的唯一标识，是唯一的，记录文件的创建修改时间、*数据在磁盘的位置*。该数据结构存储在磁盘中。
dentry(directory entry): 是 inode 和文件名的映射，文件可以有多个文件名，所以一个 inode 有多个 dentry。该
数据结构是存放在内存中，不在硬盘中，与 inode 的另一个重要区别。

linux 每次读写磁盘为一个 page，即一个数据块，默认为 4KB，大大提高磁盘读写效率。*文件系统的基本操作单位是数据块。*




# 参考文献
**[一口气搞懂文件系统](https://mp.weixin.qq.com/s/qJdoXTv_XS_4ts9YuzMNIw)**
