
#include <stdio.h>

enum server {
    SERVER_A = 0,
    SERVER_B = 1,
    MAX_SERVERS,
};

#ifndef ARRAY_SIZE
#define ARRAY_SIZE(arr) (sizeof(arr) / sizeof(arr[0]))
#endif

// IDEA 本地 RUN 调试时，Source File 里写 cilium/test/tproxy/socket-lookup/main_test.c
int main() {
    int server_fds[] = {[0 ... MAX_SERVERS - 1] = -1};
    int size = ARRAY_SIZE(server_fds);
    printf("%d\n", size); // 2

    int cnt = ~0;
    printf("%d\n", cnt); // -1
}

