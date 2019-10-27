#include <stdio.h>
#include <sys/socket.h>
#include <stdint.h>
#include <stdlib.h>
#include <netinet/in.h>

int make_socket(uint16_t port) {
    int sock;

    sock = socket(PF_INET, SOCK_STREAM, 0); // SOCK_STREAM 表示字节流，对应TCP
    if (sock < 0) {
        perror("socket");
        exit(EXIT_FAILURE);
    }

    struct sockaddr_in name;
    name.sin_family = AF_INET;
    name.sin_port = htons(port);
    name.sin_addr.s_addr = htonl(INADDR_ANY);
    if (bind(sock, (const struct sockaddr *) &name, sizeof(name)) < 0) {
        perror("bind");
        exit(EXIT_FAILURE);
    }

    return sock;
}


int main(int argc, char **argv) {
    int socket_fd = make_socket(11111);
    exit(0);
}

