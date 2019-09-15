#include <stdio.h>

struct Test {
    char* type;
    int64_t name;
} test;

int main(void) {
    printf("%d bytes \n", sizeof(int));
    printf("%d bytes \n", sizeof(int64_t));
    printf("%d bytes \n", sizeof(test));
}
