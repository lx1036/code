#include <stdio.h>

int main(void) {
    int ch;

    while ((ch = getchar()) != '#') {
        putchar(ch);
    }

    return 0;
}
