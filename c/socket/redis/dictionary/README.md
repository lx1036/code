# 字典(HashTable)的作用
**[HashTable](https://user-gold-cdn.xitu.io/2018/7/23/164c4dcd14c00534?imageView2/0/w/1280/h/960/format/webp/ignore-error/1)**

**[redis hash](https://juejin.im/post/5b53ee7e5188251aaa2d2e16#heading-2)**
Redis 的 hash 数据结构，内部使用 HashTable 存储的，原理是：对键的 hash 值，作为数组 data[] 的下标，
数组 data[] 的元素是一个单向链表（为了解决 hash 冲突），单向链表的每一个元素是 dictEntry 结构体，该结构体内的 *next 指针指向下一个 dictEntry 结构体。


