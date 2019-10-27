
#ifndef SOCKET_API_SKIPLIST_H
#define SOCKET_API_SKIPLIST_H

#include "../../sds/sds.h"



typedef struct skiplistNode {
    sds element; // 存储字符串类型数据
    double score; // 排序分值
    struct skiplistNode *backward; // 后退指针,当前节点最底层的前一个节点
    struct skiplistLevel {

    } level[];
} skiplistNode;

typedef struct skiplist {

};


/**
 * API
 */
 skiplist *create(void);

#endif //SOCKET_API_SKIPLIST_H
