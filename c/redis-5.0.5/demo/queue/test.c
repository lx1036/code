
/**
 * @see http://www.duokan.com/reader/www/app.html?id=77cfef1a3f944d64a501bf33f5b26428
 */
#include <stdio.h>
#include "queue.h"
#include <unistd.h>

int main(void) {
    Queue line;
    Item temp;
    char ch;

    /*int ch = getchar();
    printf("%d test", ch);
    _exit(1);*/

    initialize(&line);
    puts("Type a to add, Type d to delete, Type q to exit\n");

    while ((ch = getchar()) != 'q') {
        if (ch == 'a') {
            scanf("%d", &temp);

            if (!isEmpty(&line)) {
                push(temp, &line);
            } else {
                puts("Queue is full\n");
            }
        } else {
            if (isEmpty(&line)) {
                puts("Nothing to delete");
            } else {
                pop(&temp, &line);
                printf("Remove %d from queue\n", temp);
            }
        }

        printf("%d items in queue\n", count(&line));
    }

    empty(&line);
    return 0;
}
