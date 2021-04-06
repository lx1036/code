





## 基本概念
boltdb 所有修改操作都是修改内存中的数据结构，只有 commit() 之后才会进行落盘，所以内存和磁盘有两种
数据结构来存储所有数据，内存中是B+tree，磁盘中是[]byte，通过unsafe函数把[]byte转换为对应的结构体。

### Storage Architecture
https://dbdb.io/db/boltdb
There are only a few types in BoltDB: DB, Bucket, Tx, and Cursor. 
* The DB is a single file represented by pages on disk. Each page is commonly 4096 Bytes. 
  The first two pages of the file store the metadata that keeps track of version, transaction id, and page size of the database, 
  as well as the locations of the freelist and the first page id of data. 
  The third page stores a freelistthat keeps track the page id's of the free pages. 
  The rest of the pages are the collection of buckets that store the key/value paris. 

* A Bucket is a collection of unique keys that are associated with values. Each bucket is represented using a B+ tree. 
  When accessing a B+ tree, the nodes in the corresponding page are fetched into memory. 
  The B+ tree used in BoltDB is different from common B+tree in the following aspects: 
  While B+ trees typically have n+1 values and n keys in each node, the number of values and the number of keys are equal in BoltDB B+ tree. 
  In BoltDB B+ tree, the value field in a non-leaf node stores the page id of its child node and the corresponding key stores the first key in the child node. 
  There are no pointers between sibling leaf nodes in BoltDB B+ tree.

### 内存中数据结构

#### node





## 参考文献
**[boltdb源码阅读](https://zhuanlan.zhihu.com/p/346954004)**

**[boltdb 源码分析](https://youjiali1995.github.io/storage/boltdb/)**

**[boltdb B+ tree](https://youjiali1995.github.io/database/CMU-15445/)**



## TroubleShooting
**[database file size not updating?](https://github.com/boltdb/bolt/issues/308)**




## boltdb 客户端(TODO: 研究下)
**[boltdbweb web客户端](https://github.com/evnix/boltdbweb)**

**[boltdb buckets 扩展客户端](https://github.com/joyrexus/buckets)**

**[nested buckets 客户端](https://github.com/abhigupta912/mbuckets)**

**[simplebolt 扩展了几个数据结构](https://github.com/xyproto/simplebolt)**

