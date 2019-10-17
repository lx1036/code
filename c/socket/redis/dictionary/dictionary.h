
#include <stdint.h>

#ifndef __DICT_H
#define __DICT_H

#define DICT_OK 0
#define DICT_ERR 1

typedef struct dictEntry {
    void *key; // 键值对的键
    union {
        void *val; // 键值对的值，指针可以指向 string,hash,list,set,sorted-set
        uint64_t u64; //
        int64_t s64; // 键的过期时间
        double d;
    } v;
    struct dictEntry *next; // 键hash后冲突，next 指针指向下一个dictEntry，单链表解决hash冲突
} dictEntry;

typedef struct dictHashTable {
    dictEntry **table; // ??
    unsigned long size; // table数组大小
    unsigned long sizemask;
} dictHashTable;


