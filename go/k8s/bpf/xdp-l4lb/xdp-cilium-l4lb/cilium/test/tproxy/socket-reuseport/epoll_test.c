


/*
 * golang epoll 原理: https://mp.weixin.qq.com/s/xt0Elppc_OaDFnTI_tW3hg
 *
 * (1)epoll:
 * • 每次处理的 sock_fd 数量无上限
   • loop thread 通过 epoll_create 操作创建一个 epoll 池子
   • loop thread 通过 epoll_ctl 每次将一个待监听的 sock_fd 添加到 epoll 池中，linux 使用 rbtree 红黑树存储 sock_fd
   • 每当 sock_fd 红黑树中有 sock_fd 就绪事件到达时，会唤醒 loop thread. 同时内核会将处于就绪态的 sock_fd 直接告知 loop thread，无需额外遍历
 *
 * (2)select
• 一次可以处理多个 fd，体现多路. 但 fd 数量有限，最多 1024 个
• loop thread 通过 select 将一组 fd 提交到内核做监听
• 当 fd 中无 io event 就绪时，loop thread 会陷入阻塞
• 每当这组 fd 中有 io event 到达时，内核会唤醒 loop thread
• loop thread 无法精准感知到哪些 fd 就绪，需要遍历一轮 fd 列表，时间复杂度 O(N)
• 托付给内核的 fd 列表只具有一轮交互的时效. 新的轮次中，loop thread 需要重新将监听的 fd 列表再传递给内核一次
 *
 *
 */



#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <sys/socket.h>
#include <sys/epoll.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>

#define MAXLINE 1024
#define LISTENQ 20

#define ngx_memzero(buf, n)       (void) memset(buf, 0, n)
#define ngx_nonblocking(s)  fcntl(s, F_SETFL, fcntl(s, F_GETFL) | O_NONBLOCK)

// 验证可用:
// cc -o epoll_test epoll_test.c
// ./epoll_test 8024
int main(int argc, char* argv[]) {
    int portnumber;
    if (argc == 2) {
        if((portnumber = atoi(argv[1])) < 0) {
            fprintf(stderr,"Usage:%s portnumber/a/n",argv[0]);
            return 1;
        }
    } else {
        fprintf(stderr,"Usage:%s portnumber/a/n",argv[0]);
        return 1;
    }

    //声明epoll_event结构体的变量,ev用于注册事件,数组用于回传要处理的事件
    struct epoll_event ev, events[20];
    int epfd;
    epfd = epoll_create(256); // 1. 创建一个 epoll pool
    
    int sock_fd;
    sock_fd = socket(AF_INET, SOCK_STREAM, 0);
    //把socket设置为非阻塞方式
    if (ngx_nonblocking(sock_fd) == -1) {
        perror("set nonblock error");
        return 1;
    }
    //设置与要处理的事件相关的文件描述符
    ev.data.fd = sock_fd;
    //设置监听的事件类型为EPOLLIN，即读事件, 就是说关心读事件
    // EPOLLET: 将EPOLL设为边缘触发(Edge Triggered)模式，这是相对于水平触发(Level Triggered)来说的
    ev.events = EPOLLIN|EPOLLET;
    //注册epoll事件
    epoll_ctl(epfd, EPOLL_CTL_ADD, sock_fd, &ev); // 2. 将一个待监听的 fd 注册/删除到 epoll pool 中
    struct sockaddr_in serveraddr;
    ngx_memzero(&serveraddr, sizeof(serveraddr)); // 初始化结构体 &serveraddr
    serveraddr.sin_family = AF_INET;
    // serveraddr.sin_addr.s_addr= htonl(INADDR_ANY);
    char *local_addr = "127.0.0.1";
    inet_aton(local_addr, &(serveraddr.sin_addr));//htons(portnumber);
    serveraddr.sin_port = htons(portnumber);
    //绑定 ip:port
    bind(sock_fd, (struct sockaddr*)&serveraddr, sizeof(serveraddr));
    //监听连接请求
    listen(sock_fd, LISTENQ);

    socklen_t addrlen;
    struct sockaddr_in clientaddr;
    int sockfd;
    char buffer[BUFSIZ];
    ssize_t n; // number of bytes read or written
    int nfds, connfd;
    int i, nread;
    for (;;) {
        // 等待事件发生 https://man7.org/linux/man-pages/man2/epoll_wait.2.html
        nfds = epoll_wait(epfd, events, 20, -1); // 3. 当 epoll pool 中有 fd read/write 读写事件就绪时, 唤醒 loop thread 来处理数据
        printf("nfds=%d\n", nfds);
        //处理所发生的所有事件
        for (i = 0; i < nfds; i++) {
            if(events[i].data.fd == sock_fd) { // 如果新监测到一个SOCKET用户连接到了绑定的SOCKET端口，建立新的连接
                connfd = accept(sock_fd, (struct sockaddr*)&clientaddr, &addrlen);
                if(connfd < 0){
                    perror("connfd<0");
                    continue;
                }
                if (ngx_nonblocking(connfd) == -1) {
                    perror("set connfd nonblock error");
                    exit(1);
                }
                printf("accept a new connection from %s:%d\n", inet_ntoa(clientaddr.sin_addr), ntohs(clientaddr.sin_port));
                //设置与要处理的事件相关的文件描述符
                ev.data.fd = connfd;
                ev.events = EPOLLIN|EPOLLET;
                //注册epoll事件
                epoll_ctl(epfd, EPOLL_CTL_ADD, connfd, &ev);
            } else if (events[i].events&EPOLLIN) {// 如果是已经连接的用户，并且收到数据，那么进行读数据
                if ((sockfd = events[i].data.fd) < 0)
                    continue;
                
                n = 0;    
                while ((nread = read(sockfd, buffer + n, BUFSIZ)) > 0) {    
                    n += nread;
                }
                if (nread == -1 && errno != EAGAIN) {    
                    perror("read error");  
                    close(sockfd);
                    printf("read data error or closed by peer!\n");
                    //删除已关闭连接的socket文件描述符
                    epoll_ctl(epfd, EPOLL_CTL_DEL, sockfd, NULL);
                    continue;  
                }
                printf("read %ld data bytes from fd=%d\n", n, sockfd);
                // if ((n = read(sockfd, buffer, BUFSIZ)) <= 0) {
                //     close(sockfd);
                //     printf("read data error or closed by peer!\n");
                //     //删除已关闭连接的socket文件描述符
                //     epoll_ctl(epfd, EPOLL_CTL_DEL, sockfd, NULL);
                //     continue;
                // }
                //设置用于注测的写操作事件
                ev.data.fd = sockfd;
                ev.events = EPOLLOUT|EPOLLET;
                if (epoll_ctl(epfd, EPOLL_CTL_MOD, sockfd, &ev) == -1) {    
                    perror("epoll_ctl: mod");    
                }
            } else if (events[i].events&EPOLLOUT) {// 如果有数据发送
                sockfd = events[i].data.fd;
                // char *data = "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\nHello World"; // 验证这样不行
                sprintf(buffer, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\nHello World", 11);  
                int data_size = strlen(buffer);
                int nwrite;
                n = data_size;
                while (n > 0) {
                    nwrite = write(sockfd, buffer + data_size - n, n);
                    printf("write data to fd=%d %d bytes\n", sockfd, nwrite);
                    if (nwrite < n) {
                        if (nwrite == -1 && errno != EAGAIN) {
                            perror("write error");
                        }
                        break;
                    }
                    n -= nwrite;
                }
                // n = write(sockfd, data, sizeof(data));
                close(sockfd);
                // ev.data.fd=sockfd;
                // ev.events=EPOLLIN|EPOLLET;
                // epoll_ctl(epfd,EPOLL_CTL_MOD,sockfd,&ev);
            }
        }
    }

    close(epfd);
    close(sock_fd);
    return 0;
}
