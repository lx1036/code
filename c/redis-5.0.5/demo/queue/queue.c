
#include "queue.h"
#include <stdio.h>
#include <stdlib.h>

void copyToNode(Item i, Node *pNode);

void copyToItem(Node *pNode, Item *pInt);

void initialize(Queue *queue) {
    queue->items = 0;
    queue->header = queue->tail = NULL;
}

bool isFull(const Queue *queue) {
    return queue->items == MAX_QUEUE;
}

bool isEmpty(const Queue *queue) {
    return queue->items == 0;
}

int count(const Queue *queue) {
    return queue->items;
}

bool push(Item item, Queue *queue) {
    Node *pNode;

    if (isFull(queue)) return false;

    pNode = (Node *) malloc(sizeof(Node));

    if (pNode == NULL) {
        fprintf(stderr, "Unable to allocate available memory!\n");
        exit(1);
    }

    copyToNode(item, pNode);
    pNode->next = NULL;

    if (isEmpty(queue)) {
        queue->header = pNode;
    } else {
        queue->tail->next = pNode;
    }

    queue->items++;

    return true;
}

bool pop(Item *item, Queue *queue) {
    if (isEmpty(queue)) return false;

    Node *pNode;
    copyToItem(queue->header, item);

    pNode = queue->header;
    queue->header = queue->header->next;
    free(pNode);
    queue->items--;

    if (queue->items == 0) {
        queue->header = queue->tail = NULL;
    }

    return true;
}

void empty(Queue *queue) {
    Item item;

    while (!isEmpty(queue)) {
        pop(&item, queue);
    }
}

void copyToNode(Item item, Node *pNode) {
    pNode->item = item;
}

void copyToItem(Node *pNode, Item *item) {
    *item = pNode->item;
}
