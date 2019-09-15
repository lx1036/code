
#include <stdio.h>

void interchange(int *u, int *v);
int add(int x, int y);

int main(void) {
    int x = 5, y = 10;
    printf("Original x=%d y=%d \n", x, y);
    interchange(&x, &y);

    int sum = add(x, y);
    printf("Sum = %d \n", sum);
    printf("Now x=%d y=%d \n", x, y);
}

void interchange(int *u, int *v) {
    int temp = *u;
    *u = *v;
    *v = temp;
}

int add(int x, int y) {
    x++;
    y++;

    return x + y;
}
