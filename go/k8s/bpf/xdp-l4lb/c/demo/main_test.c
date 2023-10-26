
#include <stdio.h>

// cc -o main_test.o main_test.c
int main(int argc, char const *argv[])
{
    int num = 1;
    int *packet_count;
    // *packet_count = 1; // 报错
    packet_count = &num;
    (*packet_count)++;
    printf("%d\n", *packet_count); // 2
    if ((*packet_count)++ & 1) // 3 & 1 = 1, 按位与操作, false
        printf("ok\n");

    printf("%d\n", 1+2); // 3
    return 0;
}

