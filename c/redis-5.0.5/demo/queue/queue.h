
#ifndef REDIS_5_0_5_QUEUE_H
#define REDIS_5_0_5_QUEUE_H

#include <stdbool.h>

#define MAX_QUEUE 10;

typedef int Item;
typedef struct node {
    Item item;
    struct node *next;
} Node;

typedef struct queue {
    Node *header;
    Node *tail;
    int items;
} Queue;

void initialize(Queue *q);
bool isFull(const Queue *q);
bool isEmpty(const Queue *q);
int count(const Queue *q);
bool push(Item item, Queue *q);
bool pop(Item *item, Queue *q);
void empty(Queue *q);

#endif //REDIS_5_0_5_QUEUE_H
