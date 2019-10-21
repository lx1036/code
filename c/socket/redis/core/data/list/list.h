
#ifndef SOCKET_API_LIST_H
#define SOCKET_API_LIST_H







/**
 * API
 */


#define listLength(l) ((l)->len)
#define listFirst(l) ((l)->head)
#define listLast(l) ((l)->tail)
#define listPrevNode(n) ((n)->prev)
#define listNextNode(n) ((n)->next)
#define listNodeValue(n) ((n)->value)


list *listCreate(void);

#endif //SOCKET_API_LIST_H
